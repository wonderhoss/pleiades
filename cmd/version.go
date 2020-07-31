package main

import (
	"fmt"
	"strings"
)

var VERSION = "undefined"

func version() string {
	if strings.Contains(VERSION, ".") {
		return fmt.Sprintf("version %s", VERSION)
	} else {
		return fmt.Sprintf("build sha:%s", VERSION)
	}
}
