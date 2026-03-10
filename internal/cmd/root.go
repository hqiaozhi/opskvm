package cmd

import (
	"context"
	"crypto/tls"
	"opskvm/internal/controller/files"
	"opskvm/internal/controller/kvm"
	"opskvm/internal/controller/mirrors"
	"opskvm/internal/controller/system"
	"opskvm/internal/controller/totp"
	"opskvm/internal/controller/users"
	"opskvm/internal/controller/wake"
	"opskvm/internal/controller/webshell"
	_ "opskvm/internal/packed"
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
	Debug    bool   `short:"d" name:"debug" brief:"debug mode,default false (true/false)" orphan:"true"`
	RootPath string `short:"D" name:"rootpath" default:"/data/opskvm/"  brief:"root path (save data)"`
	Enroll   bool   `short:"" name:"enroll" brief:"defaut false (true/false)" orphan:"true"`
	Ssl      bool   `short:"S" name:"ssl" brief:"defaut true (true/false)" orphan:"true"`
	Swagger  bool   `short:"" name:"swagger" brief:"defaut false (true/false)" orphan:"true"`
	Cors     bool   `short:"" name:"cors" brief:"defaut false (true/false)" orphan:"true"`

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
	Ch9329Path  string `short:"" name:"ch9329path" default:"" brief:"ch9329 path, empty for auto detect"`
	VideoPath   string `short:"" name:"videopath" default:"" brief:"video path, empty for auto detect"`
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
	service.New(in.RootPath, in.Debug)

	s := g.Server()
	if !in.Ssl {
		err = NewSSLGenerator().Generate(ctx, in.RootPath, "cn", []string{}, []string{})
		if err != nil {
			return nil, err
		}
		s.SetHTTPSAddr(in.Host + ":" + in.Port)
		g.Log().Infof(ctx, "HTTPS enabled CertFile: %s KeyFile: %s", in.RootPath+"server.crt", in.RootPath+"server.key")
		s.EnableHTTPS(in.RootPath+"server.crt", in.RootPath+"server.key", &tls.Config{
			InsecureSkipVerify: true,
		})

	} else {
		s.SetAddr(in.Host + ":" + in.Port)
	}

	if in.Swagger {
		s.SetOpenApiPath("/api.json")
		s.SetSwaggerPath("/swagger")
	}

	// 下载服务
	staticDir := fileService.GetFileManagerService(in.RootPath).GetStorageDir()
	s.AddStaticPath("/downloads", staticDir)

	// 设置静态服务（web）
	// 这里是打包最终的目录
	// 例如命令：gf packed resource/public/html/dist internal/packed/data.go -n packed
	// resource/public/html/dist的最后的目录为dist
	// 需要导入 _ "opskvm/internal/packed" 才能直接使用
	s.SetServerRoot("dist") // 设置后访问 / 即可访问到dist下的静态资源
	// 重写
	s.SetRewriteMap(map[string]string{
		"/register":              "/",
		"/admin":                 "/",
		"/admin/remote":          "/",
		"/admin/file":            "/",
		"/admin/shell":           "/",
		"/admin/mirrors":         "/",
		"/admin/wol":             "/",
		"/admin/settings":        "/",
		"/admin/settings/system": "/",
		"/admin/settings/mouse":  "/",
		"/admin/settings/camera": "/",
		"/admin/settings/user":   "/",
		"/admin/settings/totp":   "/",
	})

	// 路由对象注册
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse, CORSMode(in.Cors), MiddlewareAuth)

		group.Bind(
			kvm.NewV1(),
			users.NewV1(),
			files.NewV1(),
			system.NewV1(),
			wake.NewV1(),
			mirrors.NewV1(),
			webshell.NewV1(),
			totp.NewV1(),
		)
	})
	s.SetGraceful(true)
	s.Run()

	// 等待service清理完毕才退出
	<-service.Svc.Done
	return
}

func CORSMode(mode bool) func(r *ghttp.Request) {
	if mode {
		return MiddlewareCORS
	}
	return func(r *ghttp.Request) { r.Middleware.Next() }
}
