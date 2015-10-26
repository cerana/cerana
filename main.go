package main

import (
	"flag"
	"fmt"
	"os"
)

var zfs *os.File

func init() {
	z, err := os.OpenFile("/dev/zfs", os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	zfs = z
}

var funcs = map[string]func(string) error{
	"exists":  exists,
	"destroy": destroy,
	"unmount": unmount,
}

func main() {
	name := flag.String("name", "", "dataset name")
	cmd := flag.String("cmd", "", "command to invoke")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nAvailable commands:\n")
		for cmd := range funcs {
			fmt.Fprintln(os.Stderr, "  ", cmd)
		}
	}
	flag.Parse()

	if *cmd == "" {
		fmt.Println("error: cmd is required")
		return
	}
	fn, ok := funcs[*cmd]
	if !ok {
		fmt.Println("error: uknown command", *cmd)
		return
	}

	if err := fn(*name); err != nil {
		fmt.Println("error:", err)
	}
}
