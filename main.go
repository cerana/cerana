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
	cmdSnapshot.PersistentPreRun = dummyPersistentPreRun
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

	cmdCreate := genCommand("create", "Create a ZFS dataset or volume",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			createTypeName, _ := cmd.Flags().GetString("type")
			createType, err := getDMUType(createTypeName)
			if err != nil {
				log.Fatal("invalid type")
			}

			propJSON, _ := cmd.Flags().GetString("props")
			var props map[string]interface{}
			if err := json.Unmarshal([]byte(propJSON), &props); err != nil {
				log.Fatal("bad prop json")
			}

			if volsize, ok := props["volsize"]; ok {
				if volsizeFloat, ok := volsize.(float64); ok {
					props["volsize"] = uint64(volsizeFloat)
				}
			}

			if volblocksize, ok := props["volblocksize"]; ok {
				if volblocksizeFloat, ok := volblocksize.(float64); ok {
					props["volblocksize"] = uint64(volblocksizeFloat)
				}
			}

			return create(name, createType, props)
		},
	)
	cmdCreate.Flags().StringP("type", "t", "zfs", "zfs or zvol")
	cmdCreate.Flags().StringP("props", "p", "{}", "create properties")

	cmdSend := genCommand("send", "Generate a send stream from a snapshot",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			largeBlockOK, _ := cmd.Flags().GetBool("largeblock")
			embedOK, _ := cmd.Flags().GetBool("embed")
			fromSnap, _ := cmd.Flags().GetString("fromsnap")

			outputFD := os.Stdout.Fd()
			output, _ := cmd.Flags().GetString("output")
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

			return send(name, outputFD, fromSnap, largeBlockOK, embedOK)
		},
	)
	cmdSend.Flags().StringP("output", "o", "/dev/stdout", "output file")
	cmdSend.Flags().StringP("fromsnap", "f", "", "full snap name to send incremental from")
	cmdSend.Flags().BoolP("embed", "e", false, "embed data")
	cmdSend.Flags().BoolP("largeblock", "l", false, "large block")

	cmdClone := genCommand("clone", "Creates a clone from a snapshot",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			origin, _ := cmd.Flags().GetString("origin")
			if origin == "" {
				log.Fatal("missing origin snapshot")
			}

			var props map[string]interface{}
			propJSON, _ := cmd.Flags().GetString("props")
			if err := json.Unmarshal([]byte(propJSON), &props); err != nil {
				log.Fatal("bad prop json")
			}

			return clone(name, origin, props)
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
			recurse, _ := cmd.Flags().GetBool("recurse")
			depth, _ := cmd.Flags().GetUint64("depth")
			typesList, _ := cmd.Flags().GetString("types")

			types := map[string]bool{}
			if typesList != "" {
				for _, t := range strings.Split(typesList, ",") {
					types[t] = true
				}
			}

			m, err := list(name, types, recurse, depth)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", out)
			return nil
		})
	cmdList.PersistentPreRun = dummyPersistentPreRun
	cmdList.Flags().StringP("types", "t", "", "Comma seperated list of dataset types to list.")
	cmdList.Flags().BoolP("recurse", "r", false, "Recurse to all levels.")
	cmdList.Flags().Uint64P("depth", "d", 0, "Recursion depth limit, 0 for unlimited")

	cmdGet := genCommand("get", "Get dataset properties",
		func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			recurse, _ := cmd.Flags().GetBool("recurse")
			depth, _ := cmd.Flags().GetUint64("depth")
			typesList, _ := cmd.Flags().GetString("types")

			types := map[string]bool{}
			if typesList != "" {
				for _, t := range strings.Split(typesList, ",") {
					types[t] = true
				}
			}

			m, err := properties(name, types, recurse, depth)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", out)
			return nil
		})
	cmdGet.Flags().StringP("types", "t", "", "Comma seperated list of dataset types to list.")
	cmdGet.Flags().BoolP("recurse", "r", false, "Recurse to all levels.")
	cmdGet.Flags().Uint64P("depth", "d", 0, "Recursion depth limit, 0 for unlimited")

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
