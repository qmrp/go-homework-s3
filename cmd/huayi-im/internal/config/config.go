package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	Server ServerConfig `yaml:"server" mapstructure:"SERVER"`
	Log    LogConfig    `yaml:"log" mapstructure:"LOG"`
	DB     DBConfig     `yaml:"db" mapstructure:"DB"`
	Redis  RedisConfig  `yaml:"redis" mapstructure:"REDIS"`
	Admin  AdminConfig  `yaml:"admin" mapstructure:"ADMIN"`
	Jwt    JWTConfig    `yaml:"jwt" mapstructure:"JWT"`
	API    APIConfig    `yaml:"api" mapstructure:"API"` // 添加 API 配置
}

// APIConfig API 配置
type APIConfig struct {
	BaseURL string `yaml:"base_url" mapstructure:"BASE_URL"`
}

type JWTConfig struct {
	SecretKey string `yaml:"secret_key" mapstructure:"SECRET_KEY"`
	ExpiresIn int    `yaml:"expires_in" mapstructure:"EXPIRES_IN"`
	Issuer    string `yaml:"issuer" mapstructure:"ISSUER"`
	Algorithm string `yaml:"algorithm" mapstructure:"ALGORITHM"`
}

type AdminConfig struct {
	Username string `yaml:"username" mapstructure:"USERNAME"` // ✅ yaml:"username"
	Password string `yaml:"password" mapstructure:"PASSWORD"` // ✅ yaml:"password"
	Nickname string `yaml:"nickname" mapstructure:"NICKNAME"` // ✅ yaml:"nickname"
	Email    string `yaml:"email" mapstructure:"EMAIL"`       // ✅ yaml:"email"
}

// ServerConfig 服务配置
type ServerConfig struct {
	Addr         string        `yaml:"addr" mapstructure:"ADDR"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"WRITE_TIMEOUT"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level     string `yaml:"level" mapstructure:"LEVEL"`
	Output    string `yaml:"output" mapstructure:"OUTPUT"`
	FilePath  string `yaml:"file_path" mapstructure:"FILE_PATH"`
	MaxSize   int    `yaml:"max_size" mapstructure:"MAX_SIZE"`
	MaxBackup int    `yaml:"max_backup" mapstructure:"MAX_BACKUP"`
	MaxAge    int    `yaml:"max_age" mapstructure:"MAX_AGE"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Driver          string        `yaml:"driver" mapstructure:"DRIVER"`
	DSN             string        `yaml:"dsn" mapstructure:"DSN"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"CONN_MAX_LIFETIME"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr" mapstructure:"ADDR"`
	Password string `yaml:"password" mapstructure:"PASSWORD"`
	DB       int    `yaml:"db" mapstructure:"DB"`
	PoolSize int    `yaml:"pool_size" mapstructure:"POOL_SIZE"`
}

var Cfg Config

// Load 加载配置
func Load() *Config {
	// 从环境变量获取运行环境，默认为 dev
	env := viper.GetString("ENV")
	if env == "" {
		env = "dev" // 默认开发环境
	}

	// 根据环境变量设置配置文件名
	configName := "config." + env
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// 添加多个配置文件搜索路径
	viper.AddConfigPath("./configs/")
	viper.AddConfigPath("../configs/")
	viper.AddConfigPath("./cmd/huayi-im/configs/")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()      // 自动读取环境变量
	viper.SetEnvPrefix("APP") // 环境变量前缀：APP_SERVER_ADDR
	viper.AllowEmptyEnv(true)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，则尝试使用默认配置
		fmt.Printf("警告：未找到配置文件 %s，将使用环境变量或默认值\n", configName)

		// 尝试查找可执行文件所在目录的configs文件夹
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			configPath := filepath.Join(execDir, "configs")
			viper.AddConfigPath(configPath)

			// 再次尝试读取配置文件
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("再次尝试加载配置失败：%s\n", err.Error())
			} else {
				fmt.Printf("成功从 %s 加载配置\n", viper.ConfigFileUsed())
			}
		}
	} else {
		fmt.Printf("成功从 %s 加载配置\n", viper.ConfigFileUsed())
	}

	// 反序列化配置
	if err := viper.Unmarshal(&Cfg); err != nil {
		panic("解析配置失败：" + err.Error())
	}

	return &Cfg
}
