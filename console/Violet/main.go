package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Qingluan/Violet/Funcs"
)

var (
	temp                  string
	showDoc               = false
	IsGenerateDefaultConf = false
)

func main() {
	// docs, _ := json.MarshalIndent(Funcs.Docs, "", "\t")

	flag.StringVar(&temp, "action", "run.vio", "actions list file")
	flag.BoolVar(&IsGenerateDefaultConf, "init", false, "generate default conf path")
	flag.BoolVar(&showDoc, "doc", false, "show help")

	flag.Parse()
	// Violet.New
	if IsGenerateDefaultConf {
		Funcs.GenerateBaseConf()
		return

	}
	if showDoc {
		for k, v := range Funcs.Docs {
			fmt.Println("-------------------------------------------------------------------")
			switch v.(type) {
			case map[string]string:
				fmt.Println(k, ":")
				for subk, subv := range v.(map[string]string) {
					fmt.Println("      ", subk, ":", subv)
				}
			default:
				fmt.Println(k, ":", v)

			}
		}
		return
	}
	browser, err := Funcs.NewBaseBrowser()
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()
	browser.ExecuteByFile(temp)
}
