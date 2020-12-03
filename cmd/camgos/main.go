package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"go.uber.org/zap"

	"github.com/darkclainer/camgo/pkg/querier"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	codeErrorArgs = iota + 1
	codeInternalError
)

func exitf(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

type Config struct {
	ZapConfig string
	Host      string

	Remote querier.RemoteConfig
	Cached querier.CachedConfig
}

func (c *Config) ZapConf() (*zap.Config, error) {
	if c.ZapConfig == "" {
		defaultConf := zap.NewDevelopmentConfig()
		return &defaultConf, nil
	}
	var zapConf zap.Config
	if err := json.Unmarshal([]byte(c.ZapConfig), &zapConf); err != nil {
		return nil, err
	}
	return &zapConf, nil
}

func getConfig() (*Config, *zap.Config, error) {
	pflag.StringP("config", "c", "config.yaml", "path to local config")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, nil, err
	}
	viper.SetEnvPrefix("CAMGO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("host", "localhost:8080")
	viper.BindEnv("remote.host")
	viper.BindEnv("cached.path")
	viper.BindEnv("cached.inmemory")

	configPath := viper.GetString("config")
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s", configPath)
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, nil, fmt.Errorf("error while unmarshaling config: %w", err)
	}
	zapConf, err := conf.ZapConf()
	if err != nil {
		return nil, nil, err
	}
	return &conf, zapConf, nil
}

func main() {
	conf, zapConf, err := getConfig()
	if err != nil {
		exitf(codeErrorArgs, "Failure while parsing arguments: %s", err)
	}
	logger, err := zapConf.Build()
	if err != nil {
		exitf(codeErrorArgs, "Failure while instatiating logger: %s", err)
	}
	defer logger.Sync()

	fmt.Printf("%+v\n", conf)

	logger.Info("Starting server")
	server, err := New(logger, conf)
	if err != nil {
		exitf(codeInternalError, "Can not initialize server: %s", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if err := server.Close(context.Background()); err != nil {
			logger.Error("Chutdown error", zap.Error(err))
			return
		}
	}()

	servePath := fmt.Sprintf("http://%s", conf.Host)
	logger.Info(fmt.Sprintf("Listeing started on %s\n", servePath))
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", zap.Error(err))
		}
	}
	logger.Info("Closed")
}
