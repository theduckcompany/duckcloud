package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/theduckcompany/duckcloud/src/server"
)

func GetOrCreateConfig(binaryName string) (*server.Config, error) {
	configRelPath := path.Join(binaryName, "config.json")

	cfgFullPath, err := xdg.SearchConfigFile(configRelPath)
	if err != nil {
		return writeDefaultConfig(configRelPath)
	}

	fmt.Printf("load config from %q\n", cfgFullPath)
	raw, err := os.ReadFile(cfgFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the config at %q: %w", cfgFullPath, err)
	}

	cfg := server.Config{}
	err = json.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid content for config %q: %w", cfgFullPath, err)
	}

	return &cfg, nil
}

func writeDefaultConfig(cfgRelPath string) (*server.Config, error) {
	cfgFullPath, err := xdg.ConfigFile(cfgRelPath)
	if err != nil {
		return nil, err
	}

	cfg := server.NewDefaultConfig()

	raw, err := json.MarshalIndent(cfg, "\t", "\t")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(cfgFullPath, raw, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to save the default config at %q: %w", cfgFullPath, err)
	}

	fmt.Printf("write default config at %q\n", cfgFullPath)

	return cfg, nil
}
