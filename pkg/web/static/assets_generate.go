// +build ignore

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"

	"github.com/gargath/pleiades/pkg/web/static"
)

func main() {
	err := vfsgen.Generate(static.Assets, vfsgen.Options{
		PackageName:  "static",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
