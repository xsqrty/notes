package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
	"github.com/xsqrty/notes/pkg/config/formatter"
	"github.com/xsqrty/notes/pkg/config/mode"
	"github.com/xsqrty/notes/pkg/config/size"
	"github.com/xsqrty/notes/pkg/help"
	"time"
)

type Config struct {
	Mode    mode.Mode `env:"MODE" envDefault:"dev" envDescription:"Application mode: dev, prod"`
	DB      DBConfig
	Auth    AuthConfig
	Server  ServerConfig
	Logger  LoggerConfig
	Cors    CorsConfig
	Swag    SwagConfig
	Metrics MetricsConfig
	Version string
	AppName string
}

type MetricsConfig struct {
	Port            int           `env:"METRICS_PORT" envDefault:"9090" envDescription:"Metrics port"`
	Host            string        `env:"METRICS_HOST" envDefault:"localhost" envDescription:"Metrics host"`
	Namespace       string        `env:"METRICS_NAMESPACE" envDefault:"" envDescription:"Metrics namespace"`
	Subsystem       string        `env:"METRICS_SUBSYSTEM" envDefault:"" envDescription:"Metrics subsystem"`
	ShutdownTimeout time.Duration `env:"METRICS_SHUTDOWN_TIMEOUT" envDefault:"5s" envDescription:"Metrics graceful shutdown timeout"`
}

type SwagConfig struct {
	Port            int           `env:"SWAG_PORT" envDefault:"1323" envDescription:"Swagger port"`
	Host            string        `env:"SWAG_HOST" envDefault:"localhost" envDescription:"Swagger host"`
	ShutdownTimeout time.Duration `env:"SWAG_SHUTDOWN_TIMEOUT" envDefault:"5s" envDescription:"Swagger graceful shutdown timeout"`
}

type CorsConfig struct {
	AllowedOrigins   []string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*" envDescription:"Allowed origins"`
	AllowedMethods   []string `env:"CORS_ALLOWED_METHODS" envDefault:"GET,POST,PUT,DELETE,OPTIONS" envDescription:"Allowed methods"`
	AllowedHeaders   []string `env:"CORS_ALLOWED_HEADERS" envDefault:"Accept,Content-Type,Authorization" envDescription:"Allowed methods"`
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true" envDescription:"Allow credentials"`
	MaxAge           int      `env:"CORS_MAX_AGE" envDefault:"300" envDescription:"Max age"`
}

type AuthConfig struct {
	AccessTokenExp  time.Duration `env:"ACCESS_TOKEN_EXPIRES" envDefault:"15m" envDescription:"Access token expiration"`
	RefreshTokenExp time.Duration `env:"REFRESH_TOKEN_EXPIRES" envDefault:"1h" envDescription:"Refresh token expiration"`
	PasswordCost    int           `env:"PASSWORD_COST" envDefault:"8" envDescription:"Password cost"`
}

type LoggerConfig struct {
	Stdout         bool                `env:"LOG_STDOUT" envDefault:"true" envDescription:"Logger stdout"`
	StdoutFormater formatter.Formatter `env:"LOG_STDOUT_FORMATTER" envDefault:"" envDescription:"Logger stdout formatter: json, pretty"`
	FileOut        bool                `env:"LOG_FILE_OUT" envDefault:"false" envDescription:"Logger file output"`
	DirOut         string              `env:"LOG_DIR" envDefault:"" envDescription:"Logger directory output"`
	Level          string              `env:"LOG_LEVEL" envDefault:"" envDescription:"Logger level: debug, info, warn, error"`
}

type ServerConfig struct {
	Port            int           `env:"PORT" envDefault:"8080" envDescription:"Server port"`
	Host            string        `env:"HOST" envDefault:"0.0.0.0" envDescription:"Server host"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s" envDescription:"Server graceful shutdown timeout"`
	LimitReqJson    size.Bytes    `env:"LIMIT_REQ_JSON" envDefault:"100kb" envDescription:"Limit request json size"`
}

type DBConfig struct {
	DSN string `env:"DSN" envDescription:"Data source name (pg connection)"`
}

type cmdArgs struct {
	printVersion bool
	printHelp    bool
}

var (
	Version = "dev"
	AppName = "app"
	args    *cmdArgs
)

func init() {
	args = parseCmdArgs()
}

func NewConfig() (*Config, error) {
	var config Config
	err := env.Parse(&config)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if config.Logger.Level == "" {
		switch config.Mode {
		case mode.Dev:
			config.Logger.Level = "debug"
		case mode.Prod:
			config.Logger.Level = "info"
		}
	}

	if config.Logger.StdoutFormater == formatter.None {
		switch config.Mode {
		case mode.Dev:
			config.Logger.StdoutFormater = formatter.Pretty
		case mode.Prod:
			config.Logger.StdoutFormater = formatter.Json
		}
	}

	config.Version = Version
	config.AppName = AppName

	return &config, nil
}

func (_ *Config) PrintVersion() bool {
	return args.printVersion
}

func (_ *Config) PrintHelp() bool {
	return args.printHelp
}

func (c *Config) Help() string {
	return help.GetHelp(c)
}

func parseCmdArgs() *cmdArgs {
	var args cmdArgs

	pflag.BoolVarP(&args.printVersion, "version", "v", false, "print app version")
	pflag.BoolVarP(&args.printHelp, "help", "h", false, "print help")
	pflag.Parse()

	return &args
}
