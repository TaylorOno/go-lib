package config

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	profile        string
	paths          = []string{".", "../.."}
	onConfigChange func()
	Registry       *Configuration
)

func init() {
	flag.StringVar(&profile, "profiles", "", "profiles to use if any as a comma separated list")
}

// Configuration holds the application configuration
type Configuration struct {
	*viper.Viper
}

// InitConfig initializes the application configuration must be called AFTER any flags have been registered to preserver config precedence order.
func InitConfig(_ context.Context) {
	registry := viper.New()

	// configure flags
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := registry.BindPFlags(pflag.CommandLine)
	if err != nil {
		slog.Error("failed to bind flags", slog.String("error", err.Error()))
	}

	pflag.VisitAll(func(f *pflag.Flag) {
		registry.RegisterAlias(strings.ReplaceAll(f.Name, "-", "_"), f.Name)
	})

	// configure environment variables
	registry.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	registry.AutomaticEnv()

	// register config file
	registry.SetConfigName("config")
	registry.SetConfigType("yaml")

	// add additional paths
	for _, path := range paths {
		registry.AddConfigPath(path)
	}

	// find and read the base config file
	err = registry.ReadInConfig()
	if err != nil {
		slog.Error("failed to read config file", slog.String("error", err.Error()))
	}

	// apply profile-specific overrides config files
	for _, env := range registry.GetStringSlice("profiles") {
		registry.SetConfigName(fmt.Sprintf("config-%s", env))
		registry.MergeInConfig()
	}

	// watch for config changes and allow dynamic reload
	if onConfigChange != nil {
		registry.WatchConfig()
		registry.OnConfigChange(func(e fsnotify.Event) {
			onConfigChange()
		})
	}

	Registry = &Configuration{Viper: registry}
}

// AddConfigPath adds a path to search for config files
// Must be called before InitConfig
func AddConfigPath(path string) {
	if Registry != nil {
		panic("cannot add config path after InitConfig has been called")
	}

	paths = append(paths, path)
}

// OnConfigChange registers a function that will run when a config change is detected.
// Must be called before InitConfig
func OnConfigChange(run func()) {
	if Registry != nil {
		panic("cannot register config change handler after InitConfig has been called")
	}

	onConfigChange = run
}

func GetLogLevel(key string) slog.Level {
	if Registry == nil {
		return slog.LevelInfo
	}

	levelStr := Registry.GetString(key)
	if levelStr == "" {
		return slog.LevelInfo
	}

	var level slog.Level
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		slog.Warn("failed to parse log level, defaulting to INFO", slog.String("key", key), slog.String("value", levelStr), slog.String("error", err.Error()))
		return slog.LevelInfo
	}

	return level
}
