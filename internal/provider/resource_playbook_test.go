package provider

import (
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testPlaybook() playbook {
	return playbook{
		Playbook:         types.StringValue("playbook.yaml"),
		TempInventoryDir: types.StringValue("/tmp/inventory-1"),
		Verbosity:        types.Int64Value(0),
	}
}

func assertArgs(t *testing.T, got []string, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("args = %q, want %q", got, want)
	}
}

func TestBuildPlaybookArgsPassesExtraVarsVerbatim(t *testing.T) {
	p := testPlaybook()
	p.ExtraVars = types.StringValue("plain: value\n")

	args := p.buildPlaybookArgs("")

	assertArgs(t, args, []string{"-i", "/tmp/inventory-1", "-e", "plain: value\n", "playbook.yaml"})
}

func TestBuildPlaybookArgsAppendsWriteOnlyExtraVarsFileAfterPlaintextExtraVars(t *testing.T) {
	p := testPlaybook()
	p.ExtraVars = types.StringValue("plain: value\n")

	args := p.buildPlaybookArgs("/tmp/extra-vars-1.yaml")

	assertArgs(t, args, []string{
		"-i", "/tmp/inventory-1",
		"-e", "plain: value\n",
		"-e", "@/tmp/extra-vars-1.yaml",
		"playbook.yaml",
	})
}

func TestBuildPlaybookArgsOmitsExtraVarsFileWhenThereAreNoWriteOnlyExtraVars(t *testing.T) {
	p := testPlaybook()

	args := p.buildPlaybookArgs("")

	assertArgs(t, args, []string{"-i", "/tmp/inventory-1", "playbook.yaml"})
}

func TestWriteExtraVarsFileWritesTheGivenContent(t *testing.T) {
	filePath, cleanup := writeExtraVarsFile("ansible_become_password: secret\n")
	t.Cleanup(cleanup)

	assertFileContent(t, filePath, "ansible_become_password: secret\n")
}

func TestWriteExtraVarsFileCleanupRemovesTheFile(t *testing.T) {
	filePath, cleanup := writeExtraVarsFile("ansible_become_password: secret\n")

	cleanup()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("extra vars file %s should be removed by cleanup, got err = %v", filePath, err)
	}
}

func TestNewWriteOnlyVarsMapsHostAndGroupVariablesByName(t *testing.T) {
	wo := newWriteOnlyVars(playbook{
		InventoryHosts: []inventoryHost{
			{Name: types.StringValue("host-a"), VariablesWO: types.StringValue("host_secret: a\n")},
		},
		InventoryGroups: []inventoryGroup{
			{Name: types.StringValue("group-a"), VariablesWO: types.StringValue("group_secret: b\n")},
		},
	})

	if got := wo.hosts["host-a"]; got != "host_secret: a\n" {
		t.Errorf("hosts[host-a] = %q, want %q", got, "host_secret: a\n")
	}

	if got := wo.groups["group-a"]; got != "group_secret: b\n" {
		t.Errorf("groups[group-a] = %q, want %q", got, "group_secret: b\n")
	}
}

func TestNewWriteOnlyVarsOmitsEntriesWithoutWriteOnlyVariables(t *testing.T) {
	wo := newWriteOnlyVars(playbook{
		InventoryHosts: []inventoryHost{
			{Name: types.StringValue("host-a"), Variables: types.StringValue("var_a: value\n")},
		},
		InventoryGroups: []inventoryGroup{
			{Name: types.StringValue("group-a")},
		},
	})

	if _, ok := wo.hosts["host-a"]; ok {
		t.Error("hosts should not contain host-a when it has no write-only variables")
	}

	if _, ok := wo.groups["group-a"]; ok {
		t.Error("groups should not contain group-a when it has no write-only variables")
	}
}

func TestNewWriteOnlyVarsCarriesExtraVars(t *testing.T) {
	wo := newWriteOnlyVars(playbook{ExtraVarsWO: types.StringValue("extra_secret: c\n")})

	if got := wo.extraVars.ValueString(); got != "extra_secret: c\n" {
		t.Errorf("extraVars = %q, want %q", got, "extra_secret: c\n")
	}
}
