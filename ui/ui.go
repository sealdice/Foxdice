package ui

import "embed"

var (
	//go:embed all:pkg/nav/dist
	NavAndDoc embed.FS
	//go:embed all:pkg/demo/dist
	Demo embed.FS
	Bot  embed.FS
)
