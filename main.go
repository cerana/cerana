package main

import (
	"encoding/json"
	"fmt"
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

type handler func(*cobra.Command, []string) error

func genCommand(use, short string, fn handler) *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			log.Fatal(err)
		} else {
			log.Info("success")
		}
	}
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run:   run,
	}

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

	cmdExists := genCommand("exists", "Test for dataset existence.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			return exists(name)
		})

	cmdDestroy := genCommand("destroy", "Destroys a dataset or volume.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			deferFlag, _ := cmd.Flags().GetBool("defer")
			return destroy(name, deferFlag)
		})
	cmdDestroy.Flags().BoolP("defer", "d", false, "defer destroy")

	cmdHolds := genCommand("holds", "Retrieve list of user holds on the specified snapshot.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			holds, err := holds(name)
			if err == nil {
				fmt.Println(holds)
			}
			return err
		})

	cmdSnapshot := genCommand("snapshot <snap1> <snap2> ...", "Creates a snapshot of a dataset or volume",
		func(cmd *cobra.Command, snaps []string) error {
			zpool, _ := cmd.Flags().GetString("zpool")
			if zpool == "" {
				log.Fatal("zpool required")
			}
			if len(snaps) == 0 {
				log.Fatal("at least one snapshot name required")
			}

			var props map[string]string
			propJSON, _ := cmd.Flags().GetString("props")
			if err := json.Unmarshal([]byte(propJSON), &props); err != nil {
				log.Fatal("bad prop json")
			}

			errlist, err := snapshot(zpool, snaps, props)
			if errlist != nil {
				log.Error(errlist)
			}
			return err
		},
	)
	cmdSnapshot.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Use PersistentPreRun here to replace the root one
	}
	cmdSnapshot.Flags().StringP("zpool", "z", "", "zpool")
	cmdSnapshot.Flags().StringP("props", "p", "{}", "snapshot properties")

	cmdRollback := genCommand("rollback", "Rollback this filesystem or volume to its most recent snapshot.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			result, err := rollback(name)
			if err == nil {
				fmt.Println(result)
			}
			return err
		},
	)

	root.AddCommand(
		cmdExists,
		cmdDestroy,
		cmdHolds,
		cmdSnapshot,
		cmdRollback,
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
