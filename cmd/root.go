/*
Copyright © 2023 Chris Collins collins.christopher@gmail.com

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
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/clcollins/bulk-jira-from-yaml/pkg/config"
	"github.com/clcollins/bulk-jira-from-yaml/pkg/jira"
)

var appName string = "bulk-jira-from-yaml"
var shortDescription = "Create bulk Jira tickets from a YAML file"
var longDescription = `Create bulk Jira tickets from a YAML file`

var verbose bool
var cfgFile string
var yamlFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: shortDescription,
	Long:  longDescription,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) { jira.Run() },
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s)", appName))

	rootCmd.PersistentFlags().StringVar(&yamlFile, "input", "", "YAML-formatted representation of Jira cards to be created in bulk.")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home + fmt.Sprintf("/.config/%s", appName))
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf("%s", appName))
	}

	viper.SetEnvPrefix("BJY")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if verbose {
		fmt.Fprintln(os.Stderr, err)
	}

	err := viper.Unmarshal(&config.AppConfig)
	if err != nil {
		panic(err)
	}

	for _, x := range []*string{
		&config.AppConfig.Username,
		&config.AppConfig.Token,
		&config.AppConfig.Host,
	} {
		_, err := url.Parse(*x)
		if err != nil {
			panic(err)
		}
	}
}
