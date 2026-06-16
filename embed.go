package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed templates
var templateFS embed.FS

func templateFileSystem() http.FileSystem {
	sub, err := fs.Sub(templateFS, "templates")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
