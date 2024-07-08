/*
Copyright © 2022 Will Fisher <will.fisher@progress.com>
This file is part of Omnitruck API Wrapper
*/
package cmd

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/services"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/progress-platform-services/platform-common/plogger"
	"github.com/spf13/cobra"
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
		secret := awsutils.GetSecret(os.Getenv("CONFIG"), os.Getenv("REGION"))
		err := json.Unmarshal([]byte(secret), &serviceConfig)
		if err != nil {
			logger.Fatal(err.Error())
		}
		if cliConfig.Opensource.Enabled {
			os_api := services.New(services.Config{
				Name:   cliConfig.Opensource.Name,
				Listen: cliConfig.Opensource.Listen,
				Log: logger.With(map[string]interface{}{
					"pkg": cliConfig.Opensource.Name,
				}),
				Mode:          services.Opensource,
				ServiceConfig: serviceConfig,
			})
			os_api.Start(&wg)
		}
		if cliConfig.Trial.Enabled {
			trial_api := services.New(services.Config{
				Name:   cliConfig.Trial.Name,
				Listen: cliConfig.Trial.Listen,
				Log: logger.With(map[string]interface{}{
					"pkg": cliConfig.Trial.Name,
				}),
				Mode:          services.Trial,
				ServiceConfig: serviceConfig,
			})
			trial_api.Start(&wg)
		}
		if cliConfig.Commercial.Enabled {
			commercial_api := services.New(services.Config{
				Name:   cliConfig.Commercial.Name,
				Listen: cliConfig.Commercial.Listen,
				Log: logger.With(map[string]interface{}{
					"pkg": cliConfig.Commercial.Name,
				}),
				Mode:          services.Commercial,
				ServiceConfig: serviceConfig,
			})
			commercial_api.Start(&wg)
		}
		wg.Wait()
	},
}

func setupLogging() plogger.ILogger {
	plog, _ := plogger.NewLogger(plogger.LoggerConfig{
		LogToStdout: true,
		LogLevel:    "INFO",
	})
	return plog
}

func initConfig() {
	log, _ := plogger.NewLogger(plogger.LoggerConfig{
		LogToStdout: true,
		LogLevel:    "INFO",
	})
	files, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, f := range files {
		log.Info(f.Name())
	}
	if cfgFile != "" {
		// Use config file from the flag
		yamlFile, err := os.ReadFile(cfgFile)
		if err != nil {
			log.Error("Unable to read config file", err , map[string]interface{}{
				"cfgFile":cfgFile,
			})
			return
		}

		err = yaml.Unmarshal(yamlFile, &cliConfig)
		if err != nil {
			log.Error("Error parsing config file", err , map[string]interface{}{
				"cfgFile":cfgFile,
			})
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
