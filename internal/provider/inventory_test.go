package provider

import (
	"os"
	"path"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func buildTestInventory(t *testing.T, hosts []inventoryHost, groups []inventoryGroup, wo writeOnlyVars) string {
	t.Helper()

	dir := buildPlaybookInventory("inventory-test-*", hosts, groups, wo)
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("fail to clean up temp inventory %s: %v", dir, err)
		}
	})

	return dir
}

func assertFileContent(t *testing.T, filePath string, want string) {
	t.Helper()

	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("fail to read %s: %v", filePath, err)
	}

	if string(got) != want {
		t.Errorf("%s content = %q, want %q", filePath, string(got), want)
	}
}

func TestBuildPlaybookInventoryWritesHostVariablesToVarsFileInHostDirectory(t *testing.T) {
	dir := buildTestInventory(t, []inventoryHost{
		{Name: types.StringValue("host-a"), Variables: types.StringValue("var_a: value\n")},
	}, nil, writeOnlyVars{})

	assertFileContent(t, path.Join(dir, "host_vars", "host-a", "vars.yaml"), "var_a: value\n")
}

func TestBuildPlaybookInventoryWritesHostWriteOnlyVariablesToSeparateFile(t *testing.T) {
	dir := buildTestInventory(t, []inventoryHost{
		{Name: types.StringValue("host-a"), Variables: types.StringValue("var_a: value\n")},
	}, nil, writeOnlyVars{
		hosts: map[string]string{"host-a": "ansible_become_password: secret\n"},
	})

	assertFileContent(t, path.Join(dir, "host_vars", "host-a", "vars.yaml"), "var_a: value\n")
	assertFileContent(t, path.Join(dir, "host_vars", "host-a", "vars_wo.yaml"), "ansible_become_password: secret\n")
}

func TestBuildPlaybookInventoryWritesGroupWriteOnlyVariablesToSeparateFile(t *testing.T) {
	dir := buildTestInventory(t, []inventoryHost{
		{Name: types.StringValue("host-a"), Groups: []types.String{types.StringValue("group-a")}},
	}, []inventoryGroup{
		{Name: types.StringValue("group-a"), Variables: types.StringValue("group_var: value\n")},
	}, writeOnlyVars{
		groups: map[string]string{"group-a": "vault_token: secret\n"},
	})

	assertFileContent(t, path.Join(dir, "group_vars", "group-a", "vars.yaml"), "group_var: value\n")
	assertFileContent(t, path.Join(dir, "group_vars", "group-a", "vars_wo.yaml"), "vault_token: secret\n")
}

func TestBuildPlaybookInventoryWritesOnlyTheWriteOnlyFileWhenHostHasNoPlaintextVariables(t *testing.T) {
	dir := buildTestInventory(t, []inventoryHost{
		{Name: types.StringValue("host-a")},
	}, nil, writeOnlyVars{
		hosts: map[string]string{"host-a": "ansible_password: secret\n"},
	})

	assertFileContent(t, path.Join(dir, "host_vars", "host-a", "vars_wo.yaml"), "ansible_password: secret\n")

	if _, err := os.Stat(path.Join(dir, "host_vars", "host-a", "vars.yaml")); !os.IsNotExist(err) {
		t.Errorf("vars.yaml should not exist when no plaintext variables are set, got err = %v", err)
	}
}

func TestBuildPlaybookInventoryCreatesNoVarsDirectoryWhenNoVariablesAreSet(t *testing.T) {
	dir := buildTestInventory(t, []inventoryHost{
		{Name: types.StringValue("host-a"), Groups: []types.String{types.StringValue("group-a")}},
	}, []inventoryGroup{
		{Name: types.StringValue("group-a")},
	}, writeOnlyVars{})

	for _, varsDir := range []string{"host_vars", "group_vars"} {
		if _, err := os.Stat(path.Join(dir, varsDir)); !os.IsNotExist(err) {
			t.Errorf("%s should not exist when no variables are set, got err = %v", varsDir, err)
		}
	}
}
