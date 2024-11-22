//go:build int_test

package main

import (
	"github.com/wttech/aemc/pkg"
	"sort"
	"strings"
	"testing"
)

func testInstanceList(t *testing.T, args []string, expectedResult bool, expectedIDs []string) {
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
	testInstanceList(t, []string{"instance", "list", "--output-value", "NONE"}, true, []string{"local_author", "local_publish"})
}

func TestAuthorInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-A", "--output-value", "NONE"}, true, []string{"local_author"})
}

func TestPublishInstances(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-P", "--output-value", "NONE"}, true, []string{"local_publish"})
}

func TestInstanceByID(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_author", "--output-value", "NONE"}, true, []string{"local_author"})
}

func TestInstanceByURL(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-U", "http://admin:admin@127.0.0.1:4502", "-U", "http://admin:admin@127.0.0.1:4503", "-U", "test_author=http://admin:admin@127.0.0.1:4502", "--output-value", "NONE"}, true, []string{"remote_adhoc_1", "remote_adhoc_2", "test_author"})
}

func TestInstanceByIDOrURL(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-I", "local_publish", "-U", "http://admin:admin@127.0.0.1:4502", "--output-value", "NONE"}, true, []string{"local_publish", "remote_adhoc"})
}

func TestAuthorInstanceByURL(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-U", "dev-auth_author=http://admin:admin@127.0.0.1:4502", "-U", "dev-pub1_publish=http://admin:admin@127.0.0.1:4503", "-U", "dev-pub2_publish=http://admin:admin@127.0.0.1:4504", "-A", "--output-value", "NONE"}, true, []string{"dev-auth_author"})
}

func TestPublishInstanceByURL(t *testing.T) {
	testInstanceList(t, []string{"instance", "list", "-U", "dev-auth_author=http://admin:admin@127.0.0.1:4502", "-U", "dev-pub1_publish=http://admin:admin@127.0.0.1:4503", "-U", "dev-pub2_publish=http://admin:admin@127.0.0.1:4504", "-P", "--output-value", "NONE"}, true, []string{"dev-pub1_publish", "dev-pub2_publish"})
}
