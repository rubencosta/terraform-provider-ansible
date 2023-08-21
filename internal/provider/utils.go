package provider

import (
	"log"
	"os"
	"strings"
)

func CreateVerboseSwitch(verbosity int) string {
	verbose := ""

	if verbosity == 0 {
		return verbose
	}

	verbose += "-"
	verbose += strings.Repeat("v", verbosity)

	return verbose
}

func RemoveDir(dirname string) {
	err := os.RemoveAll(dirname)
	if err != nil {
		log.Printf("Fail to remove dir %s: %v", dirname, err)
	}
}
