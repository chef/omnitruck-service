/*
Copyright © 2022 Will Fisher <will.fisher@progress.com>
This file is part of Omnitruck API Wrapper
*/
package cmd

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/chef/omnitruck-service/services"
	"github.com/chef/omnitruck-service/services/opensource"
	"github.com/chef/omnitruck-service/services/trial"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type CliConfig struct {
	Opensource ServiceDef `yaml:"opensource"`
	Trial      ServiceDef `yaml:"trial"`
}

type ServiceDef struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
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
		L := log.WithField("pkg", "cmd/start")
		log.SetOutput(os.Stdout)
		// log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)

		var wg sync.WaitGroup

		if cliConfig.Opensource.Enabled {
			L.Info("Starting Opensource API")
			os_api := opensource.NewServer(services.Config{
				Name:   cliConfig.Opensource.Name,
				Listen: cliConfig.Opensource.Listen,
				Log:    L.WithField("pkg", cliConfig.Opensource.Name),
			})
			os_api.Start(&wg)
		}
		if cliConfig.Trial.Enabled {
			L.Info("Starting Trial API")
			trial_api := trial.NewServer(services.Config{
				Name:   cliConfig.Trial.Name,
				Listen: cliConfig.Trial.Listen,
				Log:    L.WithField("pkg", cliConfig.Trial.Name),
			})
			trial_api.Start(&wg)
		}
		wg.Wait()
	},
}

func initConfig() {
	log.Info("Init Config")
	if cfgFile != "" {
		// Use config file from the flag
		yamlFile, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("cfgFile", cfgFile).Error("Unable to read config file")
			return
		}

		err = yaml.Unmarshal(yamlFile, &cliConfig)
		if err != nil {
			log.WithError(err).WithField("cfgFile", cfgFile).Error("Error parsing config file")
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
	startCmd.PersistentFlags().StringVar(&cfgFile, "config", "./.omnitruck.yml", "config file")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
