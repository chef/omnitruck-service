/*
Copyright © 2022 Will Fisher <will.fisher@progress.com>
This file is part of Omnitruck API Wrapper
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/services"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type CliConfig struct {
	Opensource ServiceDef  `yaml:"opensource"`
	Trial      ServiceDef  `yaml:"trial"`
	Commercial ServiceDef  `yaml:"commercial"`
	Logging    LoggingConf `yaml:"logging"`
}

type ServiceDef struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

type LoggingConf struct {
	Format string `yaml:"format"`
}

var (
	cfgFile   string
	cliConfig CliConfig
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := setupLogging()

		var wg sync.WaitGroup
		var serviceConfig config.ServiceConfig
		secret := awsutils.GetSecret(os.Getenv("CONFIG"), os.Getenv("REGION"), logger)
		err := json.Unmarshal([]byte(secret), &serviceConfig)
		if err != nil {
			logger.Fatal(err.Error())
		}
		if cliConfig.Opensource.Enabled {
			openSource_logger := setupLogging()
			os_api := services.New(services.Config{
				Name:          cliConfig.Opensource.Name,
				Listen:        cliConfig.Opensource.Listen,
				Log:           openSource_logger.With(zap.String("pkg", cliConfig.Opensource.Name)),
				Mode:          services.Opensource,
				ServiceConfig: serviceConfig,
			})
			os_api.Start(&wg)
		}
		if cliConfig.Trial.Enabled {
			trial_logger := setupLogging()
			trial_api := services.New(services.Config{
				Name:          cliConfig.Trial.Name,
				Listen:        cliConfig.Trial.Listen,
				Log:           trial_logger.With(zap.String("pkg", cliConfig.Trial.Name)),
				Mode:          services.Trial,
				ServiceConfig: serviceConfig,
			})
			trial_api.Start(&wg)
		}
		if cliConfig.Commercial.Enabled {
			commercial_logger := setupLogging()
			commercial_api := services.New(services.Config{
				Name:          cliConfig.Commercial.Name,
				Listen:        cliConfig.Commercial.Listen,
				Log:           commercial_logger.With(zap.String("pkg", cliConfig.Commercial.Name)),
				Mode:          services.Commercial,
				ServiceConfig: serviceConfig,
			})
			commercial_api.Start(&wg)
		}
		wg.Wait()
	},
}

func setupLogging() *zap.Logger {
	var logger *zap.Logger
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:    cliConfig.Logging.Format,
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        zapcore.OmitKey,
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      zapcore.OmitKey,
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	logger, err := config.Build()
	if err != nil {
		logger.Error("error while creating a logger: " + err.Error())
		return nil
	}
	logger = logger.With(zap.String("pkg", "cmd/start"))
	return logger
}

func initConfig() {
	var log *zap.Logger
	files, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, f := range files {
		// log.Info(f.Name())
		fmt.Println(f.Name())
	}
	if cfgFile != "" {
		// Use config file from the flag
		yamlFile, err := os.ReadFile(cfgFile)
		if err != nil {
			log.With(zap.String("cfgfile", cfgFile)).Error("error while reading the config file", zap.Error(err))
			return
		}

		err = yaml.Unmarshal(yamlFile, &cliConfig)
		if err != nil {
			log.With(zap.String("cfgfile", cfgFile)).Error("error while unmarshing the config file", zap.Error(err))
			return
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
	startCmd.PersistentFlags().StringVar(&cfgFile, "config", "./omnitruck.yml", "config file")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
