package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mistifyio/acomm"
)

func main() {
	l := acomm.NewUnixListener("foo")
	fmt.Println("URL:", l.URL())
	if err := l.Start(); err != nil {
		fmt.Println(err)
		return
	}
	defer l.Stop()

	conn := l.NextConn()
	if conn == nil {
		fmt.Println("empty conn")
		return
	}
	defer l.DoneConn(conn)

	data, err := ioutil.ReadAll(conn)
	fmt.Println("data:", string(data))
	fmt.Println("err:", err)
}
