package main

import (
	"flag"
	"log"

	"github.com/Qingluan/Violet/Funcs"
)

var (
	temp string
)

func main() {
	flag.StringVar(&temp, "action", "run.vio", "actions list file")
	flag.Parse()
	// Violet.New
	browser, err := Funcs.NewBaseBrowser()
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()
	browser.ExecuteByFile(temp)
}
