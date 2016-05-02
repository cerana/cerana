package main

import "github.com/spf13/pflag"

func main() {
	_ = newConfig(nil, nil)
	pflag.Parse()
}
