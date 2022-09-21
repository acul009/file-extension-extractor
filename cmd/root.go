/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acul009/file-extension-extractor/copier"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "file-extension-extractor",
	Short: "A tool used to copy files based on their extension",
	Long:  `A tool used to copy files based on their extension`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if _, err := os.Stat(args[0]); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("the given path %s does not exist", args[0])
			} else {
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		extensions, err := cmd.Flags().GetStringArray("extensions")
		if err != nil {
			panic(err)
		}
		blacklist, err := cmd.Flags().GetBool("blacklist")
		if err != nil {
			panic(err)
		}
		routines, err := cmd.Flags().GetInt("parralell")
		if err != nil {
			panic(err)
		}
		buffer, err := cmd.Flags().GetInt("buffer")
		if err != nil {
			panic(err)
		}
		move, err := cmd.Flags().GetBool("move")
		if err != nil {
			panic(err)
		}
		copier.StartCopy(filepath.ToSlash(args[0]), filepath.ToSlash(args[1]), extensions, blacklist, routines, buffer, move)
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.file-extension-extractor.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringArrayP("extensions", "e", []string{}, "Each extension flag adds an extension to scan for. e.g.: -r pdf")
	rootCmd.Flags().BoolP("blacklist", "b", false, "Use Blacklist instead of whitelist")
	rootCmd.Flags().BoolP("move", "m", false, "Move instead of copy")
	rootCmd.Flags().IntP("parralell", "p", 1, "Use multiple concurrent goroutines")
	rootCmd.Flags().Int("buffer", 4, "How many kilobytes the copy buffer should use")
}
