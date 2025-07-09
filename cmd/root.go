package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var Version = "v0.0.0"

var rootCmd = &cobra.Command{
	Use:     "mtc-cli",
	Short:   "The MoreThanCertified CLI",
	Long:    `This program is used to validate your MoreThanCertified lesson tasks interactively on your local machine.`,
	Version: Version,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/mtc/mtc.yaml)")
	rootCmd.PersistentFlags().StringP("api-base-url", "l", viper.GetString("api_base_url"), "API base URL")
	viper.BindPFlag("api_base_url", rootCmd.PersistentFlags().Lookup("api-base-url"))
}

func initConfig() {
	// First, handle the global config file exactly as before.
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		cobra.CheckErr(err)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(home, ".config", "mtc")
		err = os.MkdirAll(configDir, os.ModePerm)
		cobra.CheckErr(err)

		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath(configDir)
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MTC")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				viper.Set("api_base_url", "https://app.morethancertified.com/api/v1")
				err := viper.SafeWriteConfigAs(filepath.Join(configDir, "config.json"))
				cobra.CheckErr(err)
			} else {
				cobra.CheckErr(err)
			}
		}
	}

	// Now, look for a project-specific config in the current directory and merge it.
	wd, err := os.Getwd()
	cobra.CheckErr(err)
	viper.AddConfigPath(wd)
	viper.SetConfigName(".mtc")
	viper.SetConfigType("json")

	// MergeInConfig will not error if the file doesn't exist, which is perfect for us.
	viper.MergeInConfig()
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
