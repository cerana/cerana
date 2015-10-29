package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	cobra "github.com/spf13/cobra"
)

var zfs *os.File

func init() {
	z, err := os.OpenFile("/dev/zfs", os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	zfs = z
}

func main() {
	root := &cobra.Command{
		Use:  "gozfs",
		Long: "gozfs is the cli for testing the go zfs ioctl api",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Use == "gozfs" {
				return
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Fatal(err)
			}
			if name == "" {
				log.Fatal("missing name")
			}
		},
		Run: help,
	}
	root.PersistentFlags().StringP("name", "n", "", "dataset name")

	cmdExists := &cobra.Command{
		Use:   "exists",
		Short: "Test for dataset existence.",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			if err := exists(name); err != nil {
				log.Fatal(err)
			} else {
				log.Info("exists")
			}
		},
	}

	cmdDestroy := &cobra.Command{
		Use:   "destroy",
		Short: "Destroys a dataset or volume.",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			deferFlag, _ := cmd.Flags().GetBool("defer")
			if err := destroy(name, deferFlag); err != nil {
				log.Fatal(err)
			} else {
				log.Info("destroyed")
			}
		},
	}
	cmdDestroy.Flags().BoolP("defer", "d", false, "defer destroy")

	root.AddCommand(
		cmdExists,
		cmdDestroy,
	)
	if err := root.Execute(); err != nil {
		log.Fatal("root execute failed:", err)
	}
}

func help(cmd *cobra.Command, _ []string) {
	if err := cmd.Help(); err != nil {
		log.Fatal("help failed:", err)
	}
}
