package Funcs

import (
	"fmt"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
)

func (self *BaseBrowser) OperInput(id string, args []string, kargs Dict) (res Result) {
	var ele selenium.WebElement

	if name, ok := kargs["name"]; ok {
		// var ele selenium.WebElement
		var pwdele selenium.WebElement

		pwdele, res.Err = self.SmartMultiFind("//input[@type=\"password\"]")
		if res.Err != nil {
			return
		}
		pa := "../"
		limit := 20
		i := 0
		for {
			if i > limit {
				break
			}
			pa += "../"
			form, err := self.SmartMultiFind("//input[@type=\"password\"]/" + pa + "/form")
			if err == nil && form != nil {
				break
			}
			i++
		}
		ele, res.Err = self.SmartMultiFind("//input[@type=\"password\"]/" + pa + "/input[@type=\"text\"]")
		if res.Err != nil {
			return
		}
		res.Err = ele.SendKeys(name.(string))
		if res.Err != nil {
			return
		}
		L("User -- >", name.(string))
		if pwd, ok := kargs["password"]; ok {
			if res.Err != nil {
				return
			}
			res.Err = pwdele.SendKeys(pwd.(string))

			L("Pwd -- >", pwd.(string))
			self.Sleep()
			if end, ok := kargs["end"]; ok {
				switch end.(string) {
				case "\t":
					ele.SendKeys(selenium.TabKey)

				default:
					ele, res.Err = self.SmartMultiFind("//input[@type=\"password\"]/" + pa + "/*[@type=\"submit\"]")
					if res.Err != nil {
						return
					}

					if ele != nil {

						L("Submit -- >", pwd.(string))
						ele.Click()
					}
				}
			}
		}

		// ele = ele.FindElement(selenium.By)
	} else {
		ele, res.Err = self.SmartMultiFind(id)
		if res.Err != nil {
			return
		}
		if ele == nil {
			res.Err = fmt.Errorf("find err: can not found by id \"%s\"", id)
		}
		if res.Err = ele.Clear(); res.Err != nil {
			return
		}

		// Enter some new code in text box.
		if args == nil {
			res.Text, res.Err = self.driver.PageSource()
			return
		}
		res.Err = ele.SendKeys(args[0])
	}

	return
}

func (self *BaseBrowser) OperInputFromEle(i int, mainId string, args []string, kargs Dict) (res Result) {
	var ele selenium.WebElement
	var eles []selenium.WebElement
	eles, res.Err = self.SmartFindEles(mainId)
	if res.Err != nil {
		return
	}
	if i >= len(eles) {
		res.Err = fmt.Errorf("Err: %s", "beyond ele")
		return
	}
	ele = eles[i]
	var id string
	if len(args) > 1 {
		id = args[1]

		if name, ok := kargs["name"]; ok {
			// var ele selenium.WebElement
			var pwdele selenium.WebElement

			pwdele, res.Err = self.SmartFindByFromEle(ele, "//input[@type=\"password\"]")
			if res.Err != nil {
				return
			}
			pa := "../"
			limit := 20
			i := 0
			for {
				if i > limit {
					break
				}
				pa += "../"
				form, err := self.SmartFindByFromEle(ele, "//input[@type=\"password\"]/"+pa+"/form")
				if err == nil && form != nil {
					break
				}
				i++
			}
			ele, res.Err = self.SmartFindByFromEle(ele, "//input[@type=\"password\"]/"+pa+"/input[@type=\"text\"]")
			if res.Err != nil {
				return
			}
			res.Err = ele.SendKeys(name.(string))
			if res.Err != nil {
				return
			}
			L("User -- >", name.(string))
			if pwd, ok := kargs["password"]; ok {
				if res.Err != nil {
					return
				}
				res.Err = pwdele.SendKeys(pwd.(string))

				L("Pwd -- >", pwd.(string))
				self.Sleep()
				if end, ok := kargs["end"]; ok {
					switch end.(string) {
					case "\t":
						ele.SendKeys(selenium.TabKey)

					default:
						ele, res.Err = self.SmartFindByFromEle(ele, "//input[@type=\"password\"]/"+pa+"/*[@type=\"submit\"]")
						if res.Err != nil {
							return
						}

						if ele != nil {

							L("Submit -- >", pwd.(string))
							ele.Click()
						}
					}
				}
			}

			// ele = ele.FindElement(selenium.By)
		} else {
			if len(args) < 3 {
				return
			}
			ele, res.Err = self.SmartFindByFromEle(ele, id)
			if res.Err != nil {
				return
			}
			if ele == nil {
				res.Err = fmt.Errorf("find err: can not found by id \"%s\"", id)
			}
			if res.Err = ele.Clear(); res.Err != nil {
				return
			}

			// Enter some new code in text box.
			if args == nil {
				res.Text, res.Err = self.driver.PageSource()
				return
			}
			res.Err = ele.SendKeys(args[2])
		}
	} else {
		if len(args) < 2 {
			return
		}
		// ele, res.Err = self.SmartFindByFromEle(ele, id)
		// if res.Err != nil {
		// 	return
		// }
		// if ele == nil {
		// 	res.Err = fmt.Errorf("find err: can not found by id \"%s\"", id)
		// }
		if res.Err = ele.Clear(); res.Err != nil {
			return
		}

		// Enter some new code in text box.
		if args == nil {
			res.Text, res.Err = self.driver.PageSource()
			return
		}
		res.Err = ele.SendKeys(args[1])
	}

	return
}

func (self *BaseBrowser) OperCollect(id string, doc *goquery.Document, s *goquery.Selection, args []string, kargs Dict) (res Result) {
	filePath, _ := kargs["output"]
	if filePath == "" {
		res.Err = fmt.Errorf("must specify 'output' as file path")
	}
	fp, err := os.OpenFile(args[1], os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		res.Err = err
		return
	}
	defer fp.Close()
	WithEle(id, args, kargs, doc, s, func(sb *goquery.Selection, args []string) {
		if len(args) > 0 {
			_, res.Err = fp.WriteString(self.EleToJsonString(sb, args...) + "\n")
		} else {
			_, res.Err = fp.WriteString(self.EleToJsonString(sb) + "\n")
		}

	})
	return
}
