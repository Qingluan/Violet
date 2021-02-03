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
	cli                   = false
)

func main() {
	// docs, _ := json.MarshalIndent(Funcs.Docs, "", "\t")

	flag.StringVar(&temp, "action", "run.vio", "actions list file")
	flag.BoolVar(&IsGenerateDefaultConf, "init", false, "generate default conf path")
	flag.BoolVar(&showDoc, "doc", false, "show help")
	flag.BoolVar(&cli, "cli", false, "show help")

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
	if cli {
		mode := ""
		pre := " >"
		Docss := Funcs.Datas{
			"repl": "this mode can repl mode",
			"load": "run a bundle actinos in a file.",
			"exit": "exit ",
		}
		choices := ""
		for {
			browser.Clear()
			choices = Funcs.Tui.Input(mode+pre, Docss)
			if choices == "exit" {
				break
			}
			if choices == "repl" {
				mode = "repl"
				Docss = Funcs.Datas{
					"exit": "exit this mode to back mode",
				}
				for k, v := range Funcs.Docs {
					switch v.(type) {
					case map[string]string:
						ss := ""
						for subk, subv := range v.(map[string]string) {
							ss += fmt.Sprintln(subk, ":", subv)
						}
						Docss[k] = ss
					default:
						Docss[k] = fmt.Sprintf("%v", v)
					}
				}
				for {
					browser.Clear()
					choices = Funcs.Tui.Input(mode+pre, Docss)
					if choices == "exit" {
						break
					}
					browser.Parse(choices)

				}

			} else {
				file := Funcs.Tui.InputSmartPath(".")
				browser.ExecuteByFile(file)

			}
		}
	} else {
		browser.ExecuteByFile(temp)

	}
}
