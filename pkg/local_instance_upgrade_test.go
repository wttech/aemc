package pkg

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wttech/aemc/pkg/common/pathx"
)

// fakeJava stands in for the JVM running the Quickstart JAR: '-unpack' lays out the scripts of 'crx-quickstart/bin',
// any other call is the upgrade boot, served by TestLocalInstanceUpgradeBootHelper.
const fakeJava = `#!/bin/sh
case " $* " in
  *" -unpack "*)
    mkdir -p crx-quickstart/bin crx-quickstart/conf crx-quickstart/logs
    for script in start quickstart; do printf '#!/bin/sh\nexit 0\n' > crx-quickstart/bin/$script; done
    printf '#!/bin/sh\ncd "$(dirname "$0")/.."\n[ -f boot.pid ] || exit 4\ntouch stopped\nkill "$(cat boot.pid)"\n' > crx-quickstart/bin/stop
    printf '#!/bin/sh\nexit 3\n' > crx-quickstart/bin/status
    chmod +x crx-quickstart/bin/*
    ;;
  *)
    exec "$AEMC_TEST_BINARY" -test.run='^TestLocalInstanceUpgradeBootHelper$' -- "$@"
    ;;
esac
`

const upgradeBootStoppedMarker = "crx-quickstart/stopped"

// TestLocalInstanceUpgradeBootHelper is not a test but the upgrade boot run by fakeJava (the helper process pattern
// of the Go 'os/exec' tests): it answers any request like a started instance until the stop script kills it.
func TestLocalInstanceUpgradeBootHelper(t *testing.T) {
	if os.Getenv("AEMC_TEST_UPGRADE_BOOT") != "1" {
		t.Skip("run only as the upgrade boot")
	}
	port := os.Args[slices.Index(os.Args, "-p")+1]
	_ = os.WriteFile("crx-quickstart/boot.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{}"))
	})
	_ = http.ListenAndServe("127.0.0.1:"+port, nil)
	os.Exit(1)
}

func newTestUpgradeInstance(t *testing.T) (LocalInstance, string, string) {
	if runtime.GOOS == "windows" {
		t.Skip("fake JVM is a shell script")
	}
	dir := t.TempDir()
	javaHome := filepath.Join(dir, "jdk")
	require.NoError(t, os.MkdirAll(filepath.Join(javaHome, "bin"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(javaHome, "bin", "java"), []byte(fakeJava), 0755))
	libDir := filepath.Join(dir, "lib")
	require.NoError(t, os.MkdirAll(libDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(libDir, "license.properties"), []byte{}, 0644))
	t.Setenv("AEMC_TEST_BINARY", os.Args[0])
	t.Setenv("AEMC_TEST_UPGRADE_BOOT", "1")

	aem := DefaultAEM()
	aem.vendorManager.javaManager.HomeDir = javaHome
	aem.vendorManager.quickstart.DistFile = filepath.Join(libDir, "cq-quickstart-*.jar")
	aem.vendorManager.quickstart.LicenseFile = filepath.Join(libDir, "license.properties")
	im := aem.InstanceManager()
	im.LocalOpts.UnpackDir = filepath.Join(dir, "instance")
	im.CheckOpts.Skip = true
	// a custom password makes the upgrade boot be awaited until it answers, as on a real instance
	i := im.New("local_author", fmt.Sprintf("http://127.0.0.1:%d", freePort(t)), "admin", "custom_password")
	return *i.Local(), libDir, i.HTTP().Port()
}

func freePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func setTestDistFile(t *testing.T, libDir string, name string) {
	files, err := filepath.Glob(filepath.Join(libDir, "cq-quickstart-*.jar"))
	require.NoError(t, err)
	for _, file := range files {
		require.NoError(t, os.Remove(file))
	}
	require.NoError(t, os.WriteFile(filepath.Join(libDir, name), []byte{}, 0644))
}

func TestLocalInstanceUpgradeStopsBootWithStopScript(t *testing.T) {
	li, libDir, _ := newTestUpgradeInstance(t)
	setTestDistFile(t, libDir, "cq-quickstart-6.6.2.jar")
	require.NoError(t, li.Create())
	setTestDistFile(t, libDir, "cq-quickstart-6.6.3.jar")

	require.NoError(t, li.Upgrade())
	require.FileExists(t, filepath.Join(li.Dir(), upgradeBootStoppedMarker), "upgrade boot stopped with the stop script")
}

func TestLocalInstanceUpgradeStaysOutdatedAfterFailedBoot(t *testing.T) {
	li, libDir, port := newTestUpgradeInstance(t)
	setTestDistFile(t, libDir, "cq-quickstart-6.6.2.jar")
	require.NoError(t, li.Create())
	setTestDistFile(t, libDir, "cq-quickstart-6.6.3.jar")

	occupied, err := net.Listen("tcp", "127.0.0.1:"+port)
	require.NoError(t, err)
	require.Error(t, li.Upgrade(), "upgrade boot cannot start on an occupied port")
	require.NoError(t, occupied.Close())
	upgradeNeeded, err := li.IsUpgradeNeeded()
	require.NoError(t, err)
	require.True(t, upgradeNeeded, "a failed upgrade boot leaves the instance outdated, so the upgrade runs again")

	require.NoError(t, li.Upgrade())
	upgradeNeeded, err = li.IsUpgradeNeeded()
	require.NoError(t, err)
	require.False(t, upgradeNeeded)
}

func TestLocalInstanceUpgradeJarArgs(t *testing.T) {
	aem := DefaultAEM()
	i := aem.InstanceManager().New("local_author", "http://127.0.0.1:4502", "admin", "admin")
	li := *i.Local()
	li.JvmOpts = []string{"-Xmx4g"}
	li.RunModes = []string{"local"}
	li.StartOpts = []string{"-Dfoo=bar"}

	require.Equal(t, []string{
		"-Xmx4g",
		"-jar", pathx.Canonical("/lib/cq-quickstart-6.6.3.jar"),
		"-nofork", "-nointeractive", "-use-control-port",
		"-r", "author,local",
		"-p", "4502",
		"-Dfoo=bar",
	}, li.upgradeJarArgs("/lib/cq-quickstart-6.6.3.jar"))
}

func TestLocalInstanceUpgradeRefusedInServiceMode(t *testing.T) {
	li, libDir, _ := newTestUpgradeInstance(t)
	setTestDistFile(t, libDir, "cq-quickstart-6.6.2.jar")
	require.NoError(t, li.Create())
	setTestDistFile(t, libDir, "cq-quickstart-6.6.3.jar")
	li.LocalOpts().ServiceMode = true

	require.ErrorContains(t, li.Upgrade(), "service mode")
	upgradeNeeded, err := li.IsUpgradeNeeded()
	require.NoError(t, err)
	require.True(t, upgradeNeeded)
}
