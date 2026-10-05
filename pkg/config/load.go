package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/spf13/viper"
)

var envReplacer = strings.NewReplacer(".", "_", "-", "_")

// Load parses flags, environment, and config file into a Config.
func Load(args []string) (Config, error) {
	fs := flags()
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	v := viper.New()
	v.SetEnvPrefix("SYNAC")
	v.SetEnvKeyReplacer(envReplacer)
	v.AutomaticEnv()

	setDefaults(v)

	rootFlag, _ := fs.GetString(keyRoot)
	searchRoot := rootFlag
	rootExplicit := searchRoot != ""
	if searchRoot == "" {
		searchRoot = os.Getenv("SYNAC_ROOT")
		rootExplicit = searchRoot != ""
	}
	if searchRoot == "" {
		searchRoot = detectRoot()
	}

	configFile, _ := fs.GetString(keyConfig)
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName(".synac")

		if searchRoot != "" {
			v.AddConfigPath(searchRoot)
		}
		if !rootExplicit {
			v.AddConfigPath(".")
		}

		if explicit := preferredConfigFile(searchRoot, rootExplicit); explicit != "" {
			v.SetConfigFile(explicit)
		}
	}

	var notFound viper.ConfigFileNotFoundError
	if err := v.ReadInConfig(); err != nil {
		if !errors.As(err, &notFound) {
			return Config{}, errors.WithStack(fmt.Errorf("read config: %w", err))
		}
	}

	if err := v.BindPFlags(fs); err != nil {
		return Config{}, errors.WithStack(fmt.Errorf("bind flags: %w", err))
	}

	if v.InConfig(keyRoot) {
		return Config{}, errors.WithStack(fmt.Errorf("unsupported config key %q: set the root with --root or SYNAC_ROOT", keyRoot))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, errors.WithStack(fmt.Errorf("decode config: %w", err))
	}

	rootValue := rootFlag
	if rootValue == "" {
		rootValue = os.Getenv("SYNAC_ROOT")
	}
	root, err := resolveRoot(rootValue)
	if err != nil {
		return Config{}, err
	}
	cfg.Root = root

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
