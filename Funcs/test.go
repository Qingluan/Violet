package Funcs

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/c-bata/go-prompt"
)

func (self *BaseBrowser) TestSelect(page string) (res Result) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(page)))
	if err != nil {
		res.Err = err
		return
	}
	for {
		inputArgs := prompt.Input("Test Css Select:", func(d prompt.Document) (s []prompt.Suggest) {
			return
		})
		args, kargs := splitArgsTrim(inputArgs)

		fmt.Println(Blue(args, kargs))
		if len(args) < 1 {
			continue
		}
		tp := args[0]
		if tp == "exit" {
			break
		}
		if len(args) < 2 {
			continue
		}
		id := args[1]
		switch tp {
		case "xpath":
		default:
			doc.Find(id).EachWithBreak(func(i int, sb *goquery.Selection) (isReturn bool) {
				isReturn = true
				if is, ok := kargs["index"]; ok {
					if sss, _ := strconv.Atoi(is.(string)); sss == i {
						isReturn = false
						fmt.Println(Yellow(sb.Html()))
					}

				} else {
					fmt.Println(Yellow(sb.Html()))
				}
				return
			})
		}

	}
	return

}
