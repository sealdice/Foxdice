package ui

import "embed"

var (
	//go:embed all:nav/dist
	NavAndDoc embed.FS
	//go:embed all:demo/dist
	Demo embed.FS
	//go:embed all:admin/dist/spa
	Admin embed.FS
)
