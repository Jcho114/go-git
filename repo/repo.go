package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Core struct {
		FormatVersion int  `toml:"repositoryformatversion"`
		FileMode      bool `toml:"filemode"`
		Bare          bool `toml:"bare"`
	} `toml:"core"`
}

func defaultConfig() *Config {
	config := &Config{}
	config.Core.FormatVersion = 0
	config.Core.FileMode = false
	config.Core.Bare = false

	return config
}

func parseConfig(filepath string) (*Config, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = toml.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Write(filepath string) error {
	bytes, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

type Repository struct {
	Worktree string
	Gitdir   string
	Config   *Config
}

func NewRepository(path string, force bool) (*Repository, error) {
	worktree := path
	gitdir := filepath.Join(path, ".git")
	var config *Config

	_, err := os.Stat(gitdir)
	pathexists := !errors.Is(err, os.ErrNotExist)
	if !pathexists && !force {
		return nil, fmt.Errorf("not a git repository %s", path)
	}

	cfgfilepath := filepath.Join(gitdir, "config")
	_, err = os.Stat(cfgfilepath)
	pathexists = !errors.Is(err, os.ErrNotExist)
	if pathexists {
		config, err = parseConfig(cfgfilepath)
		if err != nil {
			return nil, err
		}
	} else if force {
		config = defaultConfig()
	} else {
		return nil, fmt.Errorf("config file missing")
	}

	if !force && config.Core.FormatVersion != 0 {
		err := fmt.Errorf("unsupported repositoryformatversion: %d", config.Core.FormatVersion)
		return nil, err
	}

	repo := &Repository{
		Worktree: worktree,
		Gitdir:   gitdir,
		Config:   config,
	}

	return repo, nil
}
