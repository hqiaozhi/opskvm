package main

import (
	_ "opskvm/internal/packed"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	"opskvm/internal/cmd"
)

const (
	VERSION = "v0.0.1"
)

func main() {
	g.I18n().SetLanguage("zh-CN")
	cmd.SetVersion(VERSION)
	root, err := gcmd.NewFromObject(cmd.ROOT)
	if err != nil {
		panic(err)
	}
	root.Run(gctx.New())
}
