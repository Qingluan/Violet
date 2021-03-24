package Funcs

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/manifoldco/promptui"
)

type Datas map[string]string

var (
	Docs = map[string]interface{}{
		"print:":      `text="true", cookie="true", all="true", in="true"`,
		"load":        `cookie=,header=`,
		"get:":        `https://www.,timeout=`,
		"wait:":       `timeout=7,change="true"`,
		"scroll:":     `offset=100`,
		"save:":       `cookie="True"`,
		"savescreen:": ``,
		"collecct:":   "",
		"search:":     "",
		"click:":      "index=",
		"if:":         `attrStr,compareStr`,
		"for:":        `selector_or_int`,
		"input:":      `name=,password=,end=, smart input`,
		"each:": map[string]string{
			"save": `[,contains=, find=, attrs= [...]] save node's data to file, in json format`,
		},
		"js:":     `*jscode in one line`,
		"export:": ``,
		"index=":  "",
		"global=": "",
		"output=": "",
		"mulit=":  "",
	}

	Tui = struct {
		Select         func(label string, options ...string) string
		Input          func(label string, suggest Datas, dynamic ...func(nowDesc string) Datas) string
		InputSmartPath func(root ...string) string
	}{
		Select: func(label string, options ...string) string {
			prompt := promptui.Select{
				Label:        label,
				Items:        options,
				HideSelected: true,
				Size:         10,
				Searcher: func(s string, ix int) bool {
					return strings.Contains(options[ix], s)
				},
			}
			_, result, err := prompt.Run()
			if err != nil {
				log.Println(err)
				return ""
			}
			return result
		},
		Input: func(label string, suggest Datas, dyname ...func(nowDesc string) Datas) string {
			return prompt.Input(label, func(d prompt.Document) (s []prompt.Suggest) {
				if dyname != nil {

					if now, ok := suggest[strings.TrimSpace(d.CurrentLineBeforeCursor())]; ok {

						if ss := dyname[0](now); len(ss) > 0 {
							// L("now:", ss)
							for k, v := range ss {
								s = append(s, prompt.Suggest{
									Text:        k,
									Description: v,
								})
							}
							return prompt.FilterFuzzy(s, d.GetWordBeforeCursor(), true)
							// fmt.Println(ss)
						}
					}

				}
				for k, v := range suggest {
					s = append(s, prompt.Suggest{
						Text:        k,
						Description: v,
					})
				}
				return prompt.FilterFuzzy(s, d.GetWordBeforeCursor(), true)
			})
		},
		InputSmartPath: func(root ...string) string {

			completer := func(d prompt.Document) (s []prompt.Suggest) {
				t := d.GetWordBeforeCursor()
				if root != nil {
					t = filepath.Join(root[0], t)
				}
				dir := filepath.Dir(t)
				if dir == "" {
					dir = "."
				}
				fs, err := ioutil.ReadDir(dir)
				if err != nil {
					log.Println("ioutil err:", err)
					return
				}
				for _, f := range fs {
					desc := "file"
					if f.IsDir() {
						desc = "dir"
					}
					s = append(s, prompt.Suggest{
						Text:        filepath.Join(dir, f.Name()),
						Description: desc,
					})
				}
				return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
			}
			return prompt.Input("Input a file path:", completer)
		},
	}
)

func SubCompleteByDoc(now string) Datas {
	if now != "" {
		newD := make(Datas)
		now = strings.ReplaceAll(now, "[", "")
		now = strings.ReplaceAll(now, "]", "")
		now = strings.ReplaceAll(now, "*", "")
		for _, field := range strings.Split(now, ",") {
			newD[field] = "arg complete"
		}
		// L("Now:", newD)
		return newD
	} else {
		return nil
	}
}
