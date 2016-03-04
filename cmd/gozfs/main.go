package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/gozfs"
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
			ok, err := gozfs.Exists(name)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("does not exist")
			}
			return nil
		})

	cmdDestroy := genCommand("destroy", "Destroys a dataset or volume.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			recursive, _ := cmd.Flags().GetBool("recursive")
			recursiveClones, _ := cmd.Flags().GetBool("recursiveclones")
			forceUnmount, _ := cmd.Flags().GetBool("forceunmount")
			deferDestroy, _ := cmd.Flags().GetBool("defer")

			ds, err := gozfs.GetDataset(name)
			if err != nil {
				return err
			}

			opts := &gozfs.DestroyOptions{
				Recursive:       recursive,
				RecursiveClones: recursiveClones,
				ForceUnmount:    forceUnmount,
				Defer:           deferDestroy,
			}
			return ds.Destroy(opts)
		})
	cmdDestroy.Flags().BoolP("defer", "d", false, "defer destroy")
	cmdDestroy.Flags().BoolP("recursive", "r", false, "recursively destroy datasets")
	cmdDestroy.Flags().BoolP("recursiveclones", "c", false, "recursively destroy clones")
	cmdDestroy.Flags().BoolP("forceunmount", "f", false, "force unmount")

	cmdHolds := genCommand("holds", "Retrieve list of user holds on the specified snapshot.",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			ds, err := gozfs.GetDataset(name)
			if err != nil {
				return err
			}
			holds, err := ds.Holds()
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

			ds, err := gozfs.GetDataset(nameParts[0])
			if err != nil {
				return err
			}

			return ds.Snapshot(nameParts[1], recursive)
		},
	)
	cmdSnapshot.Flags().StringP("props", "p", "{}", "snapshot properties")
	cmdSnapshot.Flags().BoolP("recursive", "r", false, "recurisvely create snapshots")

	cmdRollback := genCommand("rollback", "Rollback this filesystem or volume to its most recent snapshot.",
		func(cmd *cobra.Command, args []string) error {
			destroyMoreRecent, _ := cmd.Flags().GetBool("destroyrecent")
			name, _ := cmd.Flags().GetString("name")

			ds, err := gozfs.GetDataset(name)
			if err != nil {
				return err
			}

			return ds.Rollback(destroyMoreRecent)
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

				_, err := gozfs.CreateVolume(name, volsize, props)
				return err
			}
			_, err := gozfs.CreateFilesystem(name, props)
			return err
		},
	)
	cmdCreate.Flags().StringP("type", "t", "zfs", "zfs or zvol")
	cmdCreate.Flags().Uint64P("volsize", "s", 0, "volume size")
	cmdCreate.Flags().StringP("props", "p", "{}", "create properties")

	cmdSend := genCommand("send", "Generate a send stream from a snapshot",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			output, _ := cmd.Flags().GetString("output")

			outputWriter := os.Stdout
			if output != "/dev/stdout" {
				outputFile, err := os.Create(output)
				if err != nil {
					return err
				}
				defer outputFile.Close()

				outputWriter = outputFile
			} else {
				// If sending on stdout, don't log anything else unless there's
				// an error
				log.SetLevel(log.ErrorLevel)
			}

			ds, err := gozfs.GetDataset(name)
			if err != nil {
				return err
			}

			return ds.Send(outputWriter)
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

			ds, err := gozfs.GetDataset(origin)
			if err != nil {
				return err
			}
			_, err = ds.Clone(name, props)
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

			ds, err := gozfs.GetDataset(name)
			if err != nil {
				return err
			}

			failedName, err := ds.Rename(newName, recursive)
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

			var datasets []*gozfs.Dataset
			var err error
			switch dsType {
			case "all":
				datasets, err = gozfs.Datasets(name)
			case gozfs.DatasetFilesystem:
				datasets, err = gozfs.Filesystems(name)
			case gozfs.DatasetSnapshot:
				datasets, err = gozfs.Snapshots(name)
			case gozfs.DatasetVolume:
				datasets, err = gozfs.Volumes(name)
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
	cmdList.Flags().StringP("type", "t", "all", strings.Join([]string{"all", gozfs.DatasetFilesystem, gozfs.DatasetVolume, gozfs.DatasetSnapshot}, ","))

	cmdGet := genCommand("get", "Get dataset properties",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")

			ds, err := gozfs.GetDataset(name)
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
