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
		"get":    `*url timeout:int`,
		"wait":   `[*selector, timeout:int,change=]`,
		"scroll": `[selector,offset=int] if null will scroll download to bottom `,
		"save":   `[*filename,cookie=]`,
		"if":     `[*selector,attr:str,compareStr:str]  test @selector.attr == comparestr `,
		"for":    `[selector_or_int:str|int ] if selecotr exists or in loop limit .`,
		"input":  `[*selector: str, name=,password=,end=] , smart input`,
		"each": map[string]string{
			"save": `[*filename,contains=, find=, attrs= [...]] save node's data to file, in json format`,
		},
	}

	Tui = struct {
		Select         func(label string, options ...string) string
		Input          func(label string, suggest Datas) string
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
		Input: func(label string, suggest Datas) string {
			return prompt.Input(label, func(d prompt.Document) (s []prompt.Suggest) {
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
