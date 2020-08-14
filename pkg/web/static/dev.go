// +build dev

package static

import (
	"net/http"
)

// Assets is the local filesystem implementation of the web frontend
var Assets http.FileSystem = http.Dir("web/dist/")
