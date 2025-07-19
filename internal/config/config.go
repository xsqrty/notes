package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
	"github.com/xsqrty/notes/pkg/config/formatter"
	"github.com/xsqrty/notes/pkg/config/mode"
	"github.com/xsqrty/notes/pkg/config/size"
	"github.com/xsqrty/notes/pkg/help"
)

// Config is a central configuration for the application, defining environment-based settings and services' parameters.
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

// MetricsConfig represents the configuration for metrics.
type MetricsConfig struct {
	Port            int           `env:"METRICS_PORT"             envDefault:"9090"      envDescription:"Metrics port"`
	Host            string        `env:"METRICS_HOST"             envDefault:"localhost" envDescription:"Metrics host"`
	Namespace       string        `env:"METRICS_NAMESPACE"        envDefault:""          envDescription:"Metrics namespace"`
	Subsystem       string        `env:"METRICS_SUBSYSTEM"        envDefault:""          envDescription:"Metrics subsystem"`
	ShutdownTimeout time.Duration `env:"METRICS_SHUTDOWN_TIMEOUT" envDefault:"5s"        envDescription:"Metrics graceful shutdown timeout"`
}

// SwagConfig represents the configuration for the Swagger HTTP server.
type SwagConfig struct {
	Port            int           `env:"SWAG_PORT"             envDefault:"1323"      envDescription:"Swagger port"`
	Host            string        `env:"SWAG_HOST"             envDefault:"localhost" envDescription:"Swagger host"`
	ShutdownTimeout time.Duration `env:"SWAG_SHUTDOWN_TIMEOUT" envDefault:"5s"        envDescription:"Swagger graceful shutdown timeout"`
}

// CorsConfig represents the configuration options for Cross-Origin Resource Sharing (CORS).
type CorsConfig struct {
	AllowedOrigins   []string `env:"CORS_ALLOWED_ORIGINS"   envDefault:"*"                                 envDescription:"Allowed origins"`
	AllowedMethods   []string `env:"CORS_ALLOWED_METHODS"   envDefault:"GET,POST,PUT,DELETE,OPTIONS"       envDescription:"Allowed methods"`
	AllowedHeaders   []string `env:"CORS_ALLOWED_HEADERS"   envDefault:"Accept,Content-Type,Authorization" envDescription:"Allowed methods"`
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true"                              envDescription:"Allow credentials"`
	MaxAge           int      `env:"CORS_MAX_AGE"           envDefault:"300"                               envDescription:"Max age"`
}

// AuthConfig holds authentication-related configuration settings.
type AuthConfig struct {
	AccessTokenExp  time.Duration `env:"ACCESS_TOKEN_EXPIRES"  envDefault:"15m" envDescription:"Access token expiration"`
	RefreshTokenExp time.Duration `env:"REFRESH_TOKEN_EXPIRES" envDefault:"1h"  envDescription:"Refresh token expiration"`
	PasswordCost    int           `env:"PASSWORD_COST"         envDefault:"8"   envDescription:"Password cost"`
}

// LoggerConfig represents the configuration settings for the logger.
type LoggerConfig struct {
	Stdout         bool                `env:"LOG_STDOUT"           envDefault:"true"  envDescription:"Logger stdout"`
	StdoutFormater formatter.Formatter `env:"LOG_STDOUT_FORMATTER" envDefault:""      envDescription:"Logger stdout formatter: json, pretty"`
	FileOut        bool                `env:"LOG_FILE_OUT"         envDefault:"false" envDescription:"Logger file output"`
	DirOut         string              `env:"LOG_DIR"              envDefault:""      envDescription:"Logger directory output"`
	Level          string              `env:"LOG_LEVEL"            envDefault:""      envDescription:"Logger level: debug, info, warn, error"`
}

// ServerConfig defines the configuration for the server.
type ServerConfig struct {
	Port              int           `env:"PORT"                envDefault:"8080"    envDescription:"Server port"`
	Host              string        `env:"HOST"                envDefault:"0.0.0.0" envDescription:"Server host"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT"    envDefault:"30s"     envDescription:"Server graceful shutdown timeout"`
	LimitReqJson      size.Bytes    `env:"LIMIT_REQ_JSON"      envDefault:"100kb"   envDescription:"Limit request json size"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" envDefault:"10s"     envDescription:"Read header timeout"`
}

// DBConfig holds the database configuration, including the data source name for PostgreSQL connections.
type DBConfig struct {
	DSN string `env:"DSN" envDescription:"Data source name (pg connection)"`
}

// cmdArgs represents the structure for storing command-line argument flags.
// It holds flags for printing version and help information.
type cmdArgs struct {
	printVersion bool
	printHelp    bool
}

var (
	// Version indicates the current version of the application, defaulting to "dev". Overridden by ldflags
	Version = "dev"
	// AppName holds the name of the application, defaulting to "app". Overridden by ldflags
	AppName = "app"
	// args stores the command-line arguments and related flags for the application.
	args *cmdArgs
)

// init initializes the application by parsing command-line arguments and populating the global `args` variable.
func init() {
	args = parseCmdArgs()
}

// NewConfig initializes Config structure by parsing environment variables and setting defaults based on the application mode.
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

// PrintVersion determines whether the application version information should be printed.
func (*Config) PrintVersion() bool {
	return args.printVersion
}

// PrintHelp determines whether the help information should be printed.
func (*Config) PrintHelp() bool {
	return args.printHelp
}

// Help generates a string containing the command line arguments and environment variable information for the Config.
func (c *Config) Help() string {
	return help.GetHelp(c)
}

// parseCmdArgs parses command-line arguments and returns a pointer to a cmdArgs struct containing parsed values.
func parseCmdArgs() *cmdArgs {
	var args cmdArgs

	pflag.BoolVarP(&args.printVersion, "version", "v", false, "print app version")
	pflag.BoolVarP(&args.printHelp, "help", "h", false, "print help")
	pflag.Parse()

	return &args
}
