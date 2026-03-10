package main

import (
	"opskvm/internal/cmd"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
)

const (
	VERSION = "v0.0.2"
)

func main() {
	os.Setenv("TZ", "Asia/Shanghai")
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
	err = root.AddObject(
		cmd.Install,
		cmd.Uninstall,
	)
	if err != nil {
		panic(err)
	}

	root.Run(gctx.New())
}
