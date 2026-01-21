package conf

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const configFileName = "config"

type JWTConfig struct {
	SecretKey     string        `mapstructure:"opskvm_secret_key"`     // 密钥（必须保密）
	Issuer        string        `mapstructure:"opskvm_issuer"`         // 签发者
	Audience      string        `mapstructure:"opskvm_audience"`       // 受众
	ExpireHours   time.Duration `mapstructure:"opskvm_expire_hours"`   // 过期时间（小时）
	RefreshHours  time.Duration `mapstructure:"opskvm_refresh_hours"`  // 刷新令牌过期时间（小时）
	SigningMethod string        `mapstructure:"opskvm_signing_method"` // 签名算法（HS256/HS512）
}

type LoginUser struct {
	Username string `mapstructure:"username"` // 用户名
	Password string `mapstructure:"password"` // 密码
}

type AppConfig struct {
	Name               string        `mapstructure:"opskvm_name"`
	Host               string        `mapstructure:"opskvm_host"`
	Port               int           `mapstructure:"opskvm_port"`
	Mode               string        `mapstructure:"opskvm_mode"`                 // 运行模式（debug/release/test）
	VideoPath          string        `mapstructure:"opskvm_video_path"`           // 视频设备路径（默认/dev/video0）
	Ch9329Path         string        `mapstructure:"opskvm_ch9329_path"`          // 键盘鼠标设备路径（默认/dev/ttyUSB0）
	VideoWidth         int           `mapstructure:"opskvm_video_width"`          // 视频宽度（默认1920）
	VideoHeight        int           `mapstructure:"opskvm_video_height"`         // 视频高度（默认1080）
	VideoFPS           int           `mapstructure:"opskvm_video_fps"`            // 视频帧率（默认30）
	KMhidMode          string        `mapstructure:"opskvm_kmhid_mode"`           // 键盘鼠标模式（otg/ch9329）
	ReadTimeout        time.Duration `mapstructure:"opskvm_read_timeout"`         // 读取超时
	WriteTimeout       time.Duration `mapstructure:"opskvm_write_timeout"`        // 写入超时
	IdleTimeout        time.Duration `mapstructure:"opskvm_idle_timeout"`         // 空闲超时
	MaxMultipartMemory int64         `mapstructure:"opskvm_max_multipart_memory"` // 最大上传内存（默认10485760）
}

type Config struct {
	JWT   JWTConfig `mapstructure:"opskvm_jwt"`
	App   AppConfig `mapstructure:"opskvm_app"`
	Login LoginUser `mapstructure:"opskvm_login"`
}

// LoadConfig 初始化并加载配置
func New() (*Config, *viper.Viper, string) {
	// 设置 viper 读取 config.yaml
	v := viper.New()
	v.SetConfigName(configFileName)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/opskvm/")
	v.AddConfigPath("./conf")

	// 设置默认值
	v.SetDefault("opskvm_app.opskvm_video_path", "/dev/video0")
	v.SetDefault("opskvm_app.opskvm_video_width", 1920)
	v.SetDefault("opskvm_app.opskvm_video_height", 1080)
	v.SetDefault("opskvm_app.opskvm_video_fps", 30)
	v.SetDefault("opskvm_app.opskvm_mode", "debug")
	v.SetDefault("opskvm_app.opskvm_max_multipart_memory", 10485760)

	// 创建配置变量
	var config Config

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using default configuration")
		} else {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	// 解析配置，无论配置文件是否存在，都会使用默认值
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	configPath := v.ConfigFileUsed()
	if configPath == "" {
		configPath = "./config.yaml"
	}
	log.Println("Initial configuration loaded successfully.")

	return &config, v, configPath
}

// WatchConfigChanges 启动一个 goroutine 来监听配置文件变化并自动重新加载
func (c *Config) WatchConfigChanges(v *viper.Viper) {
	// 注意：此方法现在使用的是全局viper实例，在实际使用中应该传入正确的viper实例
	// 为了兼容性暂时保留此实现
	if v.ConfigFileUsed() != "" {
		v.WatchConfig()

		// 设置配置变化时的回调函数
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Config file changed: %s Op: %s", e.Name, e.Op.String())

			// 尝试重新加载配置到临时变量
			if err := v.Unmarshal(&c); err != nil {
				log.Printf("Error unmarshalling updated config: %v. Keeping old config.", err)
				return // 如果新配置有错误，保持旧配置不变
			}
			log.Println("Configuration reloaded successfully and applied.")
		})
	} else {
		log.Println("No config file to watch, skipping config watching.")
	}
}
