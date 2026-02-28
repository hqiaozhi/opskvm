package main

import (
	_ "opskvm/internal/packed"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"

	"opskvm/internal/cmd"
)

const (
	VERSION = "v0.0.1"
)

func main() {
	g.I18n().SetLanguage("zh-CN")
	err := gtime.SetTimeZone("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	cmd.SetVersion(VERSION)
	root, err := gcmd.NewFromObject(cmd.ROOT)
	if err != nil {
		panic(err)
	}
	root.Run(gctx.New())
}
