/*
Copyright © 2025 Juan David Cabrera Duran juandavid.juandis@gmail.com
*/
package cmd

import (
	"fmt"
	"orders-service/internal/config"
	"orders-service/pkg/logger"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.New(env)
		cfg, err := config.Load(configFile, env)
		if err != nil {
			log.Fatal("Failed to load configuration", "error", err)
		}
		log.Info(fmt.Sprintf("version: %s", cfg.Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

}
