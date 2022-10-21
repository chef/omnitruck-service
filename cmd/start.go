/*
Copyright © 2022 Will Fisher <will.fisher@progress.com>
This file is part of Omnitruck API Wrapper
*/
package cmd

import (
	"os"
	"sync"

	"github.com/chef/omnitruck-service/services"
	"github.com/chef/omnitruck-service/services/opensource"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		L := log.WithField("pkg", "cmd/opensource")
		log.SetOutput(os.Stdout)
		// log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)

		os_start := opensource.NewOpensourceServer(services.Config{
			Listen: ":3000",
		})

		var wg sync.WaitGroup

		L.Info("Starting Opensourcestart")

		os_start.Start(&wg)

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
