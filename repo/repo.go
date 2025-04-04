package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Config struct {
	Core struct {
		FormatVersion int  `ini:"repositoryformatversion"`
		FileMode      bool `ini:"filemode"`
		Bare          bool `ini:"bare"`
	} `ini:"core"`
}

func defaultConfig() *Config {
	config := &Config{}
	config.Core.FormatVersion = 0
	config.Core.FileMode = false
	config.Core.Bare = false

	return config
}

func parseConfig(filepath string) (*Config, error) {
	var cfg Config
	err := ini.MapTo(&cfg, filepath)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Write(filepath string) error {
	inicfg := ini.Empty()

	err := ini.ReflectFrom(inicfg, c)
	if err != nil {
		return err
	}

	err = inicfg.SaveTo(filepath)
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

func FindRepository(path string, required bool) (*Repository, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	gitpath := filepath.Join(abspath, ".git")
	info, err := os.Stat(gitpath)
	if err != nil {
		return nil, err
	}

	if info.Mode().IsDir() {
		return NewRepository(path, false)
	}

	parent, err := filepath.Abs(filepath.Join(path, ".."))
	if err != nil {
		return nil, err
	}

	if parent == path {
		var err error
		if required {
			err = fmt.Errorf("no git directory")
		}
		return nil, err
	}

	return FindRepository(parent, required)
}
