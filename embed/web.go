package embed

import (
	"embed"
	"io/fs"
)

//go:embed web/dist/*
var webAssets embed.FS

func WebFS() (fs.FS, error) {
	return fs.Sub(webAssets, "web/dist")
}
