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
	"strings"

	"github.com/spf13/cobra"
)

var (
	fileNameFlag         = "fileName"
	ignoreCaseFlag       = "ignoreCase"
	searchDirFlag        = "searchDir"
	linesBeforeMatchFlag = "linesBeforeMatch"
	linesAfterMatchFlag  = "linesAfterMatch"
	lineCountFlag        = "lineCount"
	includeFileFlag      = "include-file"
	excludeFileFlag      = "exclude-file"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "grep",
	Short: "command line program that implements Unix grep like functionality",
	Run: func(cmd *cobra.Command, args []string) {
		a := os.Args[1:]
		if len(a) < 2 {
			fmt.Println("error: Missing required arguments")
			cmd.Usage()
			os.Exit(1)
		}
		keyword := a[0]
		path := a[1]

		fileWriteName, err := cmd.Flags().GetString(fileNameFlag)
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
		linesBeforeMatch, err := cmd.Flags().GetInt(linesBeforeMatchFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}
		linesAfterMatch, err := cmd.Flags().GetInt(linesAfterMatchFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}
		lineCount, err := cmd.Flags().GetBool(lineCountFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}
		includeExt, err := cmd.Flags().GetString(includeFileFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}
		excludeExt, err := cmd.Flags().GetString(excludeFileFlag)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), err)
		}

		input := &GrepInput{
			keyword:          keyword,
			path:             path,
			fileWriteName:    fileWriteName,
			linesBeforeMatch: linesBeforeMatch,
			linesAfterMatch:  linesAfterMatch,
			ignoreCase:       ignoreCase,
			searchDir:        searchDir,
			lineCount:        lineCount,
			stdin:            cmd.InOrStdin(),
			output:           cmd.OutOrStdout(),
		}

		if includeExt != "" {
			input.includeExt = strings.Split(includeExt, ",")
		}
		if excludeExt != "" {
			input.excludeExt = strings.Split(excludeExt, ",")
		}

		run(os.DirFS("/"), input)
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
	rootCmd.Flags().IntP(linesAfterMatchFlag, "A", 0, "includes the line(s) after the match")
	rootCmd.Flags().IntP(linesBeforeMatchFlag, "B", 0, "includes the line(s) before the match")
	rootCmd.Flags().BoolP(lineCountFlag, "C", false, "includes the line count")
	rootCmd.Flags().StringP(includeFileFlag, "", "", "only include relevant file types")
	rootCmd.Flags().StringP(excludeFileFlag, "", "", "exclude all provided file types")
}
