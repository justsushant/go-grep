/*
Copyright © 2024 justsushant

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	fileNameFlag = "fileName"
	ignoreCaseFlag = "ignoreCase"
	searchDirFlag = "searchDir"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "grep",
	Short: "command line program that implements Unix grep like functionality",
	Run: func(cmd *cobra.Command, args []string) {
		fileName, err := cmd.Flags().GetString(fileNameFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}
		ignoreCase, err := cmd.Flags().GetBool(ignoreCaseFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}
		searchDir, err := cmd.Flags().GetBool(searchDirFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}

		result := run(os.DirFS("."), cmd.InOrStdin(), args, fileName, ignoreCase, searchDir)
		if result != "" {
			fmt.Fprintln(cmd.OutOrStdout(), result)
		}
		os.Exit(0)
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grep.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP(fileNameFlag, "o", "", "writes output to the file")
	rootCmd.Flags().BoolP(ignoreCaseFlag, "i", false, "ignores case")
	rootCmd.Flags().BoolP(searchDirFlag, "r", false, "searches directory")
}