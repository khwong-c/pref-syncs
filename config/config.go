package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/samber/do/v2"
	"github.com/samber/oops"
)

type Config struct {
	Port int `default:"7086"`
	IDP  struct {
		Enable   bool   `default:"false" usage:"Enable embedded IDP for testing."`
		User     string `default:"root"  usage:"Default user of the Local IDP."`
		Password string `default:"root"  usage:"Default password of the Local IDP."`
	}
}

func LoadConfig(do.Injector) (*Config, error) {
	cfg := Config{}
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: false,
		SkipFiles:    true,
		SkipEnv:      false,
		SkipFlags:    true,
		EnvPrefix:    "PREF_SYNC",
	})
	if err := loader.Load(); err != nil {
		return nil, oops.Wrap(err)
	}
	return &cfg, nil
}
