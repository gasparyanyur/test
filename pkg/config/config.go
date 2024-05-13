package config

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tel-io/tel/v2"
)

const (
	configFilePathEnv = "CONFIG_FILE_PATH"
)

type Options struct {
	Dir       string // dir where configs located
	File      string // config file name, set to (name).(yaml|json)
	Type      string // yaml or json
	validator *validator.Validate
}

func (o *Options) fill() error {

	if os.Getenv(configFilePathEnv) != "" {
		o.Dir = path.Dir(os.Getenv(configFilePathEnv))
		o.File = path.Base(os.Getenv(configFilePathEnv))
		o.Type = strings.ToLower(path.Ext(os.Getenv(configFilePathEnv)))
		o.Type = strings.ReplaceAll(o.Type, ".", "")
	}

	if o.Type != "json" && o.Type != "yaml" {
		return errors.Errorf("bad format %s, must be json or yaml", o.Type)
	}
	if !strings.HasSuffix(o.Dir, string(os.PathSeparator)) {
		o.Dir += string(os.PathSeparator)
	}

	o.validator = validator.New()

	return nil
}

func Parse(ctx context.Context, opts Options, configStruct interface{}) error {

	t := reflect.TypeOf(configStruct)
	if t.Kind() != reflect.Ptr {
		return errors.New("configStruct arg must be pointer")
	}
	if t.Elem().Kind() != reflect.Struct {
		return errors.New("configStruct arg must be pointer to struct")
	}

	err := opts.fill()
	if err != nil {
		return fmt.Errorf("fill options: %w", err)
	}

	configPath := opts.Dir + opts.File
	err = fileExists(configPath)
	if err != nil {
		return fmt.Errorf("verify config file: %w", err)
	}

	tel.FromCtx(ctx).Info("load config from file", tel.String("path", configPath))
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("extract absolute path: %w", err)
	}

	// load config file
	viper.SetConfigFile(configPath)
	viper.SetConfigType(opts.Type)
	err = viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	tel.FromCtx(ctx).Info("config from file loaded successfully")

	// load env vars
	viper.AllowEmptyEnv(true)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// merge file and env vars and parse data to config struct
	err = viper.Unmarshal(configStruct)
	if err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	// validate config
	err = opts.validator.Struct(configStruct)
	if err != nil {
		return err
	}

	return nil
}

func fileExists(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file not exists: " + absPath)
		}
		return err
	}
	if info.IsDir() {
		return errors.New("must be file: " + absPath)
	}
	if info.Size() == 0 {
		return errors.New("file is empty: " + absPath)
	}

	return nil
}
