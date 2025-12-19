package cmd

import (
	"opskvm/internal/router"
	"opskvm/internal/svc"
	"strconv"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var RootCmd = &cobra.Command{
	Short: "app [flag] [args]",
	Run:   startRun,
}

func startRun(cmd *cobra.Command, args []string) {
	svcCtx := svc.New(cmd.Context())

	appcfg := svcCtx.Conf.App
	addr := appcfg.Host + ":" + strconv.Itoa(appcfg.Port)
	engine := router.New(addr, appcfg.Mode, svcCtx)

	// 2. 注册业务路由（核心：解耦路由定义与引擎实现）
	engine.RegisterRoutes(router.RegisterBusinessRoutes)

	// 3. 启动服务器（阻塞，支持优雅关闭）
	engine.Run()
}

func init() {
	// 初始化根命令，这一步会自动添加 completion 命令
	RootCmd.CompletionOptions.DisableDefaultCmd = true
}
