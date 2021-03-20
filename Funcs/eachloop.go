package Funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
)

func (self *BaseBrowser) EleToJsonString(s *goquery.Selection, key ...string) string {
	ss := make(map[string]string)

	if key == nil {
		ss["html"], _ = s.Html()
		ss["text"] = s.Text()
	} else {
		if key[0] == "_" || key[0] == "*" {
			if len(s.Nodes) > 0 {
				node := s.Nodes[0]
				for _, attr := range node.Attr {
					ss[attr.Key] = ss[attr.Val]
				}
			}

		} else {
			for _, k := range key {
				if k == "text" {
					ss["text"] = s.Text()
				} else {
					ss[k] = s.AttrOr(k, "")
				}
			}

		}
	}

	res, _ := json.Marshal(&ss)
	return string(res)
}

func WithEle(args []string, kargs Dict, s *goquery.Selection, do func(sb *goquery.Selection, args []string)) {
	if ifcondition, ok := kargs["contains"]; ok {
		if !strings.Contains(s.Text(), ifcondition.(string)) {
			L("test Failed so jump")
			return
		}
	}
	if key, ok := kargs["find"]; ok {
		s.Find(key.(string)).Each(func(i int, sb *goquery.Selection) {
			if attrs, ok := kargs["attrs"]; ok {
				switch attrs.(type) {
				case []string:
					do(sb, attrs.([]string))
				}
			} else {

				do(sb, []string{})
				// _, res.Err = fp.WriteString(self.EleToJsonString(sb) + "\n")
			}
		})
	} else {
		argc := len(args)
		if argc > 2 {
			do(s, args[2:])
			// _, res.Err = fp.WriteString(self.EleToJsonString(s, args[2:]...) + "\n")
		} else {
			do(s, []string{})
			// _, res.Err = fp.WriteString(self.EleToJsonString(s) + "\n")

		}
	}
}

func (self *BaseBrowser) RunEach(id string, page string, stacks []string) (res Result) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(page)))
	if err != nil {
		res.Err = err
		return
	}
	// L("Page:", page)
	doc.Find(id).Each(func(i int, s *goquery.Selection) {
		// L("Found!", s.Nodes[0].Namespace)
		defer func() {
			if res.Err != nil {
				L("Each Err", res.Err.Error())
			}
		}()
		for _, line := range stacks {
			args, kargs := splitArgsTrim(line)
			// argc := len(args)

			// id := args[1]
			main := args[0]
			// ss := args[1:].(interface{})
			if main == "end" {
				continue
			}

			// if argc < 1 {
			// 	continue

			// }
			L("For-Each:"+main, args[1:], kargs)

			// kargs := parseKargs(args...)
			switch main {
			case "save":

				fp, err := os.OpenFile(args[1], os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
				if err != nil {
					res.Err = err
					return
				}
				defer fp.Close()
				WithEle(args, kargs, s, func(sb *goquery.Selection, args []string) {
					if len(args) > 0 {
						_, res.Err = fp.WriteString(self.EleToJsonString(sb, args...) + "\n")
					} else {
						_, res.Err = fp.WriteString(self.EleToJsonString(sb) + "\n")
					}

				})
			case "wait":
				sleep := 10
				if tk, ok := kargs["timeout"]; ok {
					sleep, res.Err = strconv.Atoi(tk.(string))
					if res.Err != nil {
						return
					}
				}
				timeout := time.Now().Add(time.Second * time.Duration(sleep))
				lastURL, _ := self.driver.CurrentURL()

				if util, ok := kargs["change"]; ok {
					if util.(string) == "url" {
						L("Wait Url Change", "timeout:", sleep)
					}
				} else {
					if id != "" {
						L("Wait Ele appearence", "timeout:", sleep)
					}
				}

				for {
					if id != "" {
						// L("Wait Ele appearence", "timeout:", sleep)
						_, err := self.SmartFindEle(id)
						if err != nil {
							res.Err = err
						} else {
							break
						}

					} else {
						if _, ok := kargs["change"]; ok {
							// if util.(string) == "url" {
							// L("Wait Url Change", "timeout:", sleep)
							thisURL, _ := self.driver.CurrentURL()
							if thisURL != lastURL {
								break
							}
							// }
						}
					}
					if time.Now().After(timeout) {
						break
					}
					time.Sleep(500 * time.Millisecond)
					res.Err = nil
				}
				res.Err = nil
			// case "if":
			// 	Ok := false
			// 	if _, ok := kargs["find"]; ok {
			// 		WithEle(args, kargs, s, func(sb *goquery.Selection, args []string) {
			// 			if strings.Contains(sb.Text(), id) {
			// 				Ok = true
			// 			}
			// 		})
			// 	} else {
			// 		if strings.Contains(s.Text(), id) {
			// 			Ok = true
			// 		}
			// 	}
			// 	res.Bool = Ok
			case "back":
				res.Err = self.driver.Back()
				self.Sleep()
			case "click":
				if len(args) == 1 {
					var eles []selenium.WebElement
					eles, res.Err = self.SmartFindEles(id)
					if res.Err != nil {

						return
					}

					L("Click ele:", len(eles))
					if i >= len(eles) {
						// res.Err = fmt.Errorf("Err: %s", "beyond ele")
						return
					}
					ele := eles[i]
					_, res.Err = ele.LocationInView()
					L("Click Then")
					if res.Err != nil {
						return
					}
					time.Sleep(4 * time.Second)
					res.Err = ele.Click()
					L("Click ok")
					self.Sleep()
					if res.Err != nil {
						return
					}
				} else {
					before, _ := self.driver.PageSource()
					var ele selenium.WebElement
					ele, res.Err = self.SmartFindEle(id)

					if res.Err != nil {
						return
					}
					if ele == nil {

						return
					}
					if ok, _ := ele.IsEnabled(); ok {
						if res.Err = ele.Click(); res.Err != nil {

							L("click error")
							return
						} else {
							defer func() {
								after, _ := self.driver.PageSource()
								if after == before {
									res.Err = fmt.Errorf("not click :%s", id)
								}
							}()
						}
					} else {
						L("Not ok click")
						res.Err = fmt.Errorf("not click :%s", id)
					}
				}

			case "input":
				var eles []selenium.WebElement
				eles, res.Err = self.SmartFindEles(id)
				if res.Err != nil {
					return
				}
				if i >= len(eles) {
					res.Err = fmt.Errorf("Err: %s", "beyond ele")
					return
				}
				ele := eles[i]
				res.Err = ele.SendKeys(args[1])
				self.Sleep()
			case "print":
				WithEle(args, kargs, s, func(sb *goquery.Selection, args []string) {
					L("show", self.EleToJsonString(s, args...))
				})
			}

			// case "find":
			// 	s.Find(args[1]).Each(func(i int, sb *goquery.Selection) {
			// 		L("find", args[1], sb.Text())
			// 	})
			// }
		}
	})

	return
}
