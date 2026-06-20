/*
Copyright © 2025 Motalleb Fallahnezha

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"os"

	"github.com/fmotalleb/go-tools/git"
	"github.com/fmotalleb/go-tools/log"
	"github.com/spf13/cobra"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/service"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "north-outage",
	Short:   "A brief description of your application",
	Version: git.String(),
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if verbose, err := cmd.Flags().GetBool("verbose"); err != nil {
			return err
		} else if verbose {
			log.SetDebugDefaults()
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		// ctx = (ctx, os.Interrupt, os.Kill)
		var err error
		if ctx, err = log.WithNewEnvLogger(ctx); err != nil {
			return err
		}
		var cfg *config.Config
		var cfgPath string
		if cfgPath, err = cmd.Flags().GetString("config"); err != nil {
			return err
		}
		if cfg, err = config.ReadConfig(ctx, cfgPath); err != nil {
			return err
		}
		ctx = config.Attach(ctx, cfg)
		return service.Serve(ctx)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("config", "c", "./config.toml", "config path")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "output will be more verbose")
}
