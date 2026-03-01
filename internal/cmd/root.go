package cmd

import (
	"context"
	"opskvm/internal/controller/files"
	"opskvm/internal/controller/kvm"
	"opskvm/internal/controller/users"
	"opskvm/internal/service"
	fileService "opskvm/internal/service/files"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// GF is the management object for `gf` command line tool.
var ROOT = Init{}

type Init struct {
	g.Meta `name:"main"`
}

type CIintInput struct {
	g.Meta `name:"main"`
	// APP配置
	Host     string `short:"H" name:"host" default:"0.0.0.0" brief:"server host"`
	Port     string `short:"P" name:"port" default:"8080"  brief:"port of http server"`
	Username string `short:"u" name:"username" default:"admin"  brief:"login username"`
	Password string `short:"p" name:"password" default:"admin123"  brief:"login password"`

	// JWT配置
	SecretKey     string        `short:"s" name:"secretkey" default:"hv4cW0kHLoigQcmVlHACmOIwVFaIQhd0qIf7SXgy4sffFRcmere85VKZrtbuMcH9"  brief:"jwt secret key"`
	Issuer        string        `short:"i" name:"issuer" default:"opskvm"  brief:"jwt issuer"`
	Audience      string        `short:"a" name:"audience" default:"opskvm"  brief:"jwt audience"`
	ExpireHours   time.Duration `short:"e" name:"expire" default:"2"  brief:"jwt expire hours"`
	RefreshHours  time.Duration `short:"r" name:"refresh" default:"12"  brief:"jwt refresh hours"`
	SigningMethod string        `short:"m" name:"signingmethod" default:"HS256"  brief:"jwt signing method"`

	// OTG和鼠键配置
	hidMode     string `short:"" name:"hidmode" default:"otg" brief:"hid mode otg/ch9329"`
	UdcName     string `short:"" name:"udcname" default:"opskvm" brief:"udc name"`
	Ch9329Path  string `short:"" name:"ch9329path" default:"/dev/ttyUSB0" brief:"ch9329 path"`
	VideoPath   string `short:"" name:"videopath" default:"auto" brief:"video path"`
	VideoWidth  int    `short:"" name:"videowidth" default:"1920" brief:"video width"`
	VideoHeight int    `short:"" name:"videoheight" default:"1080" brief:"video height"`
	VideoFPS    uint32 `short:"" name:"videofps" default:"30" brief:"video fps"`

	// 查看视频设备
	VideoDevice bool `short:"l" name:"list" brief:"find video device" orphan:"true"`

	// 指定日志级别
	LogLevel string `short:"" name:"loglevel" default:"info" brief:"set log level"`

	// 版本
	Version bool `short:"v" name:"version" brief:"show version information of current binary" orphan:"true"`
}

type CInitOutput struct{}

func (c Init) Index(ctx context.Context, in CIintInput) (out *CInitOutput, err error) {
	err = c.cmdLine(ctx, in)
	if err != nil {
		return nil, err
	}

	// 初始化Service
	service.New()

	s := g.Server()
	s.SetGraceful(true)
	s.SetAddr(in.Host + ":" + in.Port)
	s.SetOpenApiPath("/api.json")
	s.SetSwaggerPath("/swagger")

	staticDir := fileService.GetFileManagerService().GetStorageDir()
	s.AddStaticPath("/downloads", staticDir)

	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse, MiddlewareCORS)
		group.Bind(
			kvm.NewV1(),
			users.NewV1(),
			files.NewV1(),
		)
	})

	s.Run()

	// 等待service清理完毕才退出
	<-service.Svc.Done
	return
}
