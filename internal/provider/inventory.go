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

func buildPlaybookInventory(inventoryDest string, hosts []inventoryHost, groups []inventoryGroup) string {
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
		if !h.Variables.IsNull() {
			err := os.MkdirAll(path.Join(destinationDir, "host_vars"), 0755)
			if err != nil {
				log.Fatalf("Fail to create host_vars dir: %v", err)
			}
			err = os.WriteFile(path.Join(destinationDir, "host_vars", h.Name.ValueString()), []byte(h.Variables.ValueString()), 0644)
			if err != nil {
				log.Fatalf("Fail to create host_vars file: %v", err)
			}
		}
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
		if !g.Variables.IsNull() {
			err := os.MkdirAll(path.Join(destinationDir, "group_vars"), 0755)
			if err != nil {
				log.Fatalf("Fail to create group_vars dir: %v", err)
			}
			err = os.WriteFile(path.Join(destinationDir, "group_vars", g.Name.ValueString()), []byte(g.Variables.ValueString()), 0644)
			if err != nil {
				log.Fatalf("Fail to create group_vars file: %v", err)
			}
		}
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
