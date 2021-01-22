package main

import (
	"flag"
	"log"

	"github.com/Qingluan/Violet/Funcs"
)

var (
	temp                  string
	IsGenerateDefaultConf = false
)

func main() {
	flag.StringVar(&temp, "action", "run.vio", "actions list file")
	flag.BoolVar(&IsGenerateDefaultConf, "init", false, "generate default conf path")

	flag.Parse()
	// Violet.New
	if IsGenerateDefaultConf {
		Funcs.GenerateBaseConf()
		return

	}
	browser, err := Funcs.NewBaseBrowser()
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()
	browser.ExecuteByFile(temp)
}
