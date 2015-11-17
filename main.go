package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	cobra "github.com/spf13/cobra"
)

type handler func(*cobra.Command, []string) error

func genCommand(use, short string, fn handler) *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			log.Fatal(err)
		} else {
			log.Debug("success")
		}
	}
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run:   run,
	}

}

func main() {
	dummyPersistentPreRun := func(cmd *cobra.Command, args []string) {
		// Use PersistentPreRun here to replace the root one
	}
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
			recursive, _ := cmd.Flags().GetBool("recursive")
			recursiveClones, _ := cmd.Flags().GetBool("recursiveclones")
			forceUnmount, _ := cmd.Flags().GetBool("forceunmount")
			deferDestroy, _ := cmd.Flags().GetBool("defer")

			d, err := GetDataset(name)
			if err != nil {
				return err
			}

			opts := &DestroyOptions{
				Recursive:       recursive,
				RecursiveClones: recursiveClones,
				ForceUnmount:    forceUnmount,
				Defer:           deferDestroy,
			}
			return d.Destroy(opts)
		})
	cmdDestroy.Flags().BoolP("defer", "d", false, "defer destroy")
	cmdDestroy.Flags().BoolP("recursive", "r", false, "recursively destroy datasets")
	cmdDestroy.Flags().BoolP("recursiveclones", "c", false, "recursively destroy clones")
	cmdDestroy.Flags().BoolP("forceunmount", "f", false, "force unmount")

	cmdHolds := genCommand("holds", "Retrieve list of user holds on the specified snapshot.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			holds, err := holds(name)
			if err == nil {
				fmt.Println(holds)
			}
			return err
		})

	cmdSnapshot := genCommand("snapshot", "Creates a snapshot of a dataset or volume",
		func(cmd *cobra.Command, snaps []string) error {
			recursive, _ := cmd.Flags().GetBool("recursive")
			name, _ := cmd.Flags().GetString("name")
			nameParts := strings.Split(name, "@")
			if len(nameParts) != 2 {
				log.Fatal("invalid snapshot name")
			}

			var props map[string]string
			propJSON, _ := cmd.Flags().GetString("props")
			if err := json.Unmarshal([]byte(propJSON), &props); err != nil {
				log.Fatal("bad prop json")
			}

			d, err := GetDataset(nameParts[0])
			if err != nil {
				return err
			}

			return d.Snapshot(nameParts[1], recursive)
		},
	)
	cmdSnapshot.Flags().StringP("props", "p", "{}", "snapshot properties")
	cmdSnapshot.Flags().BoolP("recursive", "r", false, "recurisvely create snapshots")

	cmdRollback := genCommand("rollback", "Rollback this filesystem or volume to its most recent snapshot.",
		func(cmd *cobra.Command, args []string) error {
			destroyMoreRecent, _ := cmd.Flags().GetBool("destroyrecent")
			name, _ := cmd.Flags().GetString("name")

			d, err := GetDataset(name)
			if err != nil {
				return err
			}

			return d.Rollback(destroyMoreRecent)
		},
	)
	cmdRollback.Flags().BoolP("destroyrecent", "r", false, "destroy more recent snapshots and their clones")

	cmdCreate := genCommand("create", "Create a ZFS dataset or volume",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			volsize, _ := cmd.Flags().GetUint64("volsize")
			createTypeName, _ := cmd.Flags().GetString("type")

			propJSON, _ := cmd.Flags().GetString("props")
			var props map[string]interface{}
			if err := json.Unmarshal([]byte(propJSON), &props); err != nil {
				log.Fatal("bad prop json")
			}

			if createTypeName == "zvol" {
				if volblocksize, ok := props["volblocksize"]; ok {
					if volblocksizeFloat, ok := volblocksize.(float64); ok {
						props["volblocksize"] = uint64(volblocksizeFloat)
					}
				}

				_, err := CreateVolume(name, volsize, props)
				return err
			} else {
				_, err := CreateFilesystem(name, props)
				return err
			}
		},
	)
	cmdCreate.Flags().StringP("type", "t", "zfs", "zfs or zvol")
	cmdCreate.Flags().Uint64P("volsize", "s", 0, "volume size")
	cmdCreate.Flags().StringP("props", "p", "{}", "create properties")

	cmdSend := genCommand("send", "Generate a send stream from a snapshot",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			output, _ := cmd.Flags().GetString("output")

			outputFD := os.Stdout.Fd()
			if output != "/dev/stdout" {
				outputFile, err := os.Create(output)
				if err != nil {
					return err
				}
				defer outputFile.Close()

				outputFD = outputFile.Fd()
			} else {
				// If sending on stdout, don't log anything else unless there's
				// an error
				log.SetLevel(log.ErrorLevel)
			}

			d, err := GetDataset(name)
			if err != nil {
				return err
			}

			return d.Send(outputFD)
		},
	)
	cmdSend.Flags().StringP("output", "o", "/dev/stdout", "output file")

	cmdClone := genCommand("clone", "Creates a clone from a snapshot",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			origin, _ := cmd.Flags().GetString("origin")
			if origin == "" {
				log.Fatal("missing origin snapshot name")
			}

			var props map[string]interface{}
			propJSON, _ := cmd.Flags().GetString("props")
			if err := json.Unmarshal([]byte(propJSON), &props); err != nil {
				log.Fatal("bad prop json")
			}

			d, err := GetDataset(origin)
			if err != nil {
				return err
			}
			_, err = d.Clone(name, props)
			return err
		},
	)
	cmdClone.Flags().StringP("origin", "o", "", "name of origin snapshot")
	cmdClone.Flags().StringP("props", "p", "{}", "snapshot properties")

	cmdRename := genCommand("rename", "Rename a dataset",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			newName, _ := cmd.Flags().GetString("newname")
			recursive, _ := cmd.Flags().GetBool("recursive")

			failedName, err := rename(name, newName, recursive)
			if failedName != "" {
				log.Error(failedName)
			}
			return err
		},
	)
	cmdRename.Flags().StringP("newname", "o", "", "new name of dataset")
	cmdRename.Flags().BoolP("recursive", "r", false, "recursively rename snapshots")

	cmdList := genCommand("list", "List filesystems, volumes, snapshots and bookmarks.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			dsType, _ := cmd.Flags().GetString("type")
			if dsType == "" {
				dsType = "all"
			}

			var datasets []*Dataset
			var err error
			switch dsType {
			case "all":
				datasets, err = Datasets(name)
			case DatasetFilesystem:
				datasets, err = Filesystems(name)
			case DatasetSnapshot:
				datasets, err = Snapshots(name)
			case DatasetVolume:
				datasets, err = Volumes(name)
			}
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(datasets, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", out)
			return nil
		})
	cmdList.PersistentPreRun = dummyPersistentPreRun
	cmdList.Flags().StringP("type", "t", "all", strings.Join([]string{"all", DatasetFilesystem, DatasetVolume, DatasetSnapshot}, ","))

	cmdGet := genCommand("get", "Get dataset properties",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")

			ds, err := GetDataset(name)
			if err != nil {
				return err
			}

			out, err := json.MarshalIndent(ds, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", out)
			return nil
		})

	root.AddCommand(
		cmdExists,
		cmdDestroy,
		cmdHolds,
		cmdSnapshot,
		cmdRollback,
		cmdCreate,
		cmdSend,
		cmdClone,
		cmdRename,
		cmdList,
		cmdGet,
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
