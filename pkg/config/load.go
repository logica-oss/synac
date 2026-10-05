package config

import (
	"fmt"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/spf13/viper"
)

var envReplacer = strings.NewReplacer(".", "_", "-", "_")

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
		v.AddConfigPath(".")

		if explicit := preferredConfigFile(searchRoot); explicit != "" {
			v.SetConfigFile(explicit)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return Config{}, errors.WithStack(fmt.Errorf("read config: %w", err))
		}
	}

	if err := v.BindPFlags(fs); err != nil {
		return Config{}, errors.WithStack(fmt.Errorf("bind flags: %w", err))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, errors.WithStack(fmt.Errorf("decode config: %w", err))
	}

	root, err := resolveRoot(cfg.Root)
	if err != nil {
		return Config{}, err
	}
	cfg.Root = root

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
