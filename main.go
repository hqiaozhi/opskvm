package main

import (
	_ "opskvm/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"opskvm/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
