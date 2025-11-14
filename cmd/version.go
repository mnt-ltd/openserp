/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "unknown"
	GitHash = "unknown"
	BuildAt = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of openserp",
	Long:  `Print the version number of openserp application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
		fmt.Println("AppName: ", "openserp")
		fmt.Println("Version: ", Version)
		fmt.Println("BuildAt: ", BuildAt)
		fmt.Println("GitHash: ", GitHash)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
