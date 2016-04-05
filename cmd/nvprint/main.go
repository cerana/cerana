package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/cerana/cerana/zfs/nv"
)

func main() {
	skip := flag.Int("skip", 0, "number of leading bytes to skip")
	indent := flag.String("indent", "  ", "indentation value")
	flag.Parse()

	if *skip > 0 {
		buf := make([]byte, *skip)
		i, err := io.ReadFull(os.Stdin, buf)
		if i != *skip {
			fmt.Println("failed to skip leading bytes")
			return
		}
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	var r io.Reader
	files := flag.Args()
	if len(files) == 0 {
		r = os.Stdin
	} else {
		readers := make([]io.Reader, len(files))
		for i := range files {
			f, err := os.Open(files[i])
			if err != nil {
				panic(err)
			}
			readers[i] = f
		}
		r = io.MultiReader(readers...)
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err)
		return
	}

	var out bytes.Buffer

	err = nv.PrettyPrint(&out, buf, *indent)
	if err != nil {
		fmt.Println(err)
		return
	}
	out.WriteString("\n")
	out.WriteTo(os.Stdout)
}
