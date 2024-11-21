//go:build int_test

package main

import (
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/cfg"
	"os"
	"sort"
	"strings"
	"testing"
)

func testInstanceList(t *testing.T, args []string, expectedResult bool, expectedIDs []string) {
	if err := os.Setenv(cfg.FileEnvVar, "../../test/resources/aem.yml"); err != nil {
		t.Fatal(err)
	}

	cli := NewCLI()
	cmd := cli.cmd
	cmd.SetArgs(args)

	defer func() {
		err := recover()
		actualResult := err != nil
		if actualResult != expectedResult {
			t.Errorf("InstanceList(%v) = %v; want %v", args, actualResult, expectedResult)
		} else if actualResult && expectedResult {
			actualIDs := []string{}
			instances := cli.outputResponse.Data["instances"].([]pkg.Instance)
			for _, i := range instances {
				actualIDs = append(actualIDs, i.ID())
			}
			sort.SliceStable(actualIDs, func(i, j int) bool {
				return strings.Compare(actualIDs[i], actualIDs[j]) < 0
			})
			sort.SliceStable(expectedIDs, func(i, j int) bool {
				return strings.Compare(expectedIDs[i], expectedIDs[j]) < 0
			})
			result := len(actualIDs) == len(expectedIDs)
			for i := range actualIDs {
				result = result && actualIDs[i] == expectedIDs[i]
			}
			if !result {
				t.Errorf("InstanceList(%v) = %v; want %v", args, actualIDs, expectedIDs)
			}
		}
	}()

	_ = cmd.Execute()
}

func TestAllInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list"}, true, []string{"local_author", "local_publish", "remote_author_dev-auth", "remote_publish_dev-pub1", "remote_publish_dev-pub2"})
}

func TestAuthorInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-A"}, true, []string{"local_author", "remote_author_dev-auth"})
}

func TestPublishInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-P"}, true, []string{"local_publish", "remote_publish_dev-pub1", "remote_publish_dev-pub2"})
}

func TestLocalInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-L"}, true, []string{"local_author", "local_publish"})
}

func TestLocalAuthorInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-L", "-A"}, true, []string{"local_author"})
}

func TestRemoteInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-R"}, true, []string{"remote_author_dev-auth", "remote_publish_dev-pub1", "remote_publish_dev-pub2"})
}

func TestRemotePublishInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-R", "-P"}, true, []string{"remote_publish_dev-pub1", "remote_publish_dev-pub2"})
}

func TestDevInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-C", "dev"}, true, []string{"remote_author_dev-auth", "remote_publish_dev-pub1", "remote_publish_dev-pub2"})
}

func TestPublishDevInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-P", "-C", "dev"}, true, []string{"remote_publish_dev-pub1", "remote_publish_dev-pub2"})
}

func TestInstanceByID(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_author", "-I", "remote_publish_dev-pub2"}, true, []string{"local_author", "remote_publish_dev-pub2"})
}

func TestInstanceByURL(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-U", "http://admin:admin@127.0.0.1:4502", "-U", "http://admin:admin@127.0.0.1:4503", "-U", "test_author=http://admin:admin@127.0.0.1:4502"}, true, []string{"remote_adhoc_1", "remote_adhoc_2", "test_author"})
}

func TestConflictByIDAndAuthor(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_author", "-A"}, false, []string{})
}

func TestConflictByIDAndLocal(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_author", "-L"}, false, []string{})
}

func TestConflictByIDAndClassifier(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_author", "-C", "dev"}, false, []string{})
}

func TestConflictIDAndURL(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_author", "-U", "http://admin:admin@127.0.0.1:4502"}, false, []string{})
}

func TestConflictAuthorAndPublish(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-A", "-P"}, false, []string{})
}

func TestConflictAndLocalAndRemote(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-L", "-R"}, false, []string{})
}
