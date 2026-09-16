package provider

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"gopkg.in/ini.v1"
)

const defaultHostGroup = "ungrouped"

// writeOnlyVars carries the write-only variable values for a playbook run. They
// are read from the resource config rather than the plan, and are deliberately
// kept out of the resource model so they can never be written to state.
type writeOnlyVars struct {
	// hosts maps a host name to its yaml encoded write-only variables.
	hosts map[string]string
	// groups maps a group name to its yaml encoded write-only variables.
	groups map[string]string
	// extraVars holds yaml encoded write-only variables for the whole play.
	extraVars types.String
}

func buildPlaybookInventory(inventoryDest string, hosts []inventoryHost, groups []inventoryGroup, wo writeOnlyVars) string {
	destinationDir, err := os.MkdirTemp("", inventoryDest)
	if err != nil {
		log.Fatalf("Fail to create temp inventory directory: %v", err)
	}
	inventoryFileInfo, err := os.Create(path.Join(destinationDir, "hosts"))
	if err != nil {
		log.Fatalf("Fail to create inventory file: %v", err)
	}

	inventoryFileName := inventoryFileInfo.Name()
	log.Printf("Inventory %s was created", inventoryFileName)

	inventory, err := ini.Load(inventoryFileName)
	if err != nil {
		log.Printf("Fail to read inventory: %v", err)
	}

	inventoryMap := make(map[string][]string)
	for _, h := range hosts {
		hostGroups := h.Groups
		hostName := h.Name.ValueString()
		if len(h.Groups) == 0 {
			hostGroups = []types.String{types.StringValue(defaultHostGroup)}
		}
		for _, group := range hostGroups {
			g := group.ValueString()
			_, ok := inventoryMap[g]
			if !ok {
				inventoryMap[g] = []string{}
			}
			if !slices.Contains(inventoryMap[g], hostName) {
				inventoryMap[g] = append(inventoryMap[g], hostName)
			}
		}
		writeVarsFiles(destinationDir, "host_vars", hostName, h.Variables, wo.hosts[hostName])
	}
	for _, g := range groups {
		name := g.Name.ValueString() + ":children"
		for _, c := range g.Children {
			childName := c.ValueString()
			_, ok := inventoryMap[name]
			if !ok {
				inventoryMap[name] = []string{}
			}
			if !slices.Contains(inventoryMap[name], childName) {
				inventoryMap[name] = append(inventoryMap[name], childName)
			}
		}
		groupName := g.Name.ValueString()
		writeVarsFiles(destinationDir, "group_vars", groupName, g.Variables, wo.groups[groupName])
	}

	for k, v := range inventoryMap {
		_, err := inventory.NewRawSection(k, strings.Join(v, "\n"))
		if err != nil {
			log.Fatalf("Fail to create inventory section: %v", err)
		}
	}

	err = inventory.SaveTo(inventoryFileName)
	if err != nil {
		log.Fatalf("Fail to create inventory: %v", err)
	}

	return destinationDir
}

// writeVarsFiles writes the variables of a single inventory entry into
// <destinationDir>/<varsDir>/<name>/. Ansible loads every file in an entry's
// vars directory, which lets the plaintext and write-only values stay in
// separate files so the provider never has to parse or merge yaml itself.
// "vars_wo.yaml" sorts after "vars.yaml", so write-only values take precedence
// on conflicting keys.
func writeVarsFiles(destinationDir string, varsDir string, name string, variables types.String, writeOnly string) {
	files := map[string]string{}
	if !variables.IsNull() {
		files["vars.yaml"] = variables.ValueString()
	}

	if writeOnly != "" {
		files["vars_wo.yaml"] = writeOnly
	}

	if len(files) == 0 {
		return
	}

	entryDir := path.Join(destinationDir, varsDir, name)

	err := os.MkdirAll(entryDir, 0755)
	if err != nil {
		log.Fatalf("Fail to create %s dir for %s: %v", varsDir, name, err)
	}

	for fileName, content := range files {
		err = os.WriteFile(path.Join(entryDir, fileName), []byte(content), 0600)
		if err != nil {
			log.Fatalf("Fail to create %s file %s for %s: %v", varsDir, fileName, name, err)
		}
	}
}

// writeExtraVarsFile writes yaml encoded write-only extra variables to a
// private temporary file so the value never becomes part of the
// ansible-playbook command line, which the provider stores in state as "cmd".
// The returned cleanup removes the file, and should run as soon as the playbook
// has finished.
func writeExtraVarsFile(content string) (string, func()) {
	file, err := os.CreateTemp("", "extra-vars-*.yaml")
	if err != nil {
		log.Fatalf("Fail to create write-only extra vars file: %v", err)
	}

	filePath := file.Name()
	cleanup := func() {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Fail to remove write-only extra vars file %s: %v", filePath, err)
		}
	}

	if _, err := file.WriteString(content); err != nil {
		cleanup()
		log.Fatalf("Fail to write write-only extra vars file: %v", err)
	}

	if err := file.Close(); err != nil {
		cleanup()
		log.Fatalf("Fail to close write-only extra vars file: %v", err)
	}

	return filePath, cleanup
}

// newWriteOnlyVars collects the write-only values out of a playbook decoded
// from the resource config. Write-only attributes are null in both the plan and
// the state, so the config is the only place they can be read from.
func newWriteOnlyVars(cfg playbook) writeOnlyVars {
	wo := writeOnlyVars{
		hosts:     map[string]string{},
		groups:    map[string]string{},
		extraVars: cfg.ExtraVarsWO,
	}

	for _, h := range cfg.InventoryHosts {
		if !h.VariablesWO.IsNull() {
			wo.hosts[h.Name.ValueString()] = h.VariablesWO.ValueString()
		}
	}

	for _, g := range cfg.InventoryGroups {
		if !g.VariablesWO.IsNull() {
			wo.groups[g.Name.ValueString()] = g.VariablesWO.ValueString()
		}
	}

	return wo
}
