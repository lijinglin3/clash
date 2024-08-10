package constant

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

const Name = "clash"

// Path is used to get the configuration path
//
// on Unix systems, `$HOME/.config/clash`.
// on Windows, `%USERPROFILE%/.config/clash`.
var Path = func() *iPath {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir, _ = os.Getwd()
	}

	homeDir = path.Join(homeDir, ".config", Name)

	if _, err = os.Stat(homeDir); err != nil {
		if configHome, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
			homeDir = path.Join(configHome, Name)
		}
	}
	return &iPath{homeDir: homeDir, configFile: "config.yaml"}
}()

type iPath struct {
	homeDir    string
	configFile string
}

// SetHomeDir is used to set the configuration path
func SetHomeDir(root string) {
	Path.homeDir = root
}

// SetConfig is used to set the configuration file
func SetConfig(file string) {
	Path.configFile = file
}

func (p *iPath) HomeDir() string {
	return p.homeDir
}

func (p *iPath) Config() string {
	return p.configFile
}

// Resolve return a absolute path or a relative path with homedir
func (p *iPath) Resolve(path string) string {
	if !filepath.IsAbs(path) {
		return filepath.Join(p.HomeDir(), path)
	}

	return path
}

// IsSubPath return true if path is a subpath of homedir
func (p *iPath) IsSubPath(path string) bool {
	homedir := p.HomeDir()
	path = p.Resolve(path)
	rel, err := filepath.Rel(homedir, path)
	if err != nil {
		return false
	}

	return !strings.Contains(rel, "..")
}

func (p *iPath) MMDB() string {
	return path.Join(p.homeDir, "Country.mmdb")
}

func (p *iPath) OldCache() string {
	return path.Join(p.homeDir, ".cache")
}

func (p *iPath) Cache() string {
	return path.Join(p.homeDir, "cache.db")
}
