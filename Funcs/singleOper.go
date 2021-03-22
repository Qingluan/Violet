package Funcs

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tebeka/selenium"
)

func (self *BaseBrowser) OperWait(id string, args []string, kargs Dict) (res Result) {
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
			_, err := self.SmartMultiFind(id)
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
	return
}

func (self *BaseBrowser) OperScrollTo(id string, args []string, kargs Dict) (res Result) {
	if id != "" {
		var ele selenium.WebElement
		ele, res.Err = self.SmartMultiFind(id)
		if res.Err != nil {
			L("scroll err", res.Err)
			return
		} else {
			// L("Found :", ele)
		}
		p, err := ele.Location()
		if err != nil {
			res.Err = err
		}
		L("move to :", p.X, p.Y)
		self.driver.ExecuteScript(fmt.Sprintf("window.scrollTo(0, %d);", p.Y), nil)
		self.Sleep()
	} else {
		if f, ok := kargs["offset"]; ok {
			moveOffset, _ := strconv.Atoi(f.(string))
			self.driver.ExecuteScript(fmt.Sprintf("window.scrollTo(0, %d);", moveOffset), nil)
			self.Sleep()

		} else {
			self.driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil)
		}
	}
	if t, ok := kargs["timeout"]; ok {
		if w, err := strconv.Atoi(t.(string)); err == nil {
			time.Sleep(time.Duration(w) * time.Second)
		}
	}
	return
}

func (self *BaseBrowser) OperClickToId(id string, args []string, kargs Dict) (res Result) {
	before, _ := self.driver.PageSource()
	var ele selenium.WebElement
	if i, ok := kargs["index"]; ok {
		if is, err := strconv.Atoi(i.(string)); err == nil {
			es, _ := self.SmartFindEles(id)
			if len(es) > is {
				ele = es[is]
			}
		}
	} else if _, ok := kargs["auto"]; ok {
		es, _ := self.SmartFindEles(id)
		for i, e := range es {
			log.Println(Green("auto Click: try to click no.", i))
			if res.Err = e.Click(); res.Err != nil {
				continue
			} else {
				defer func() {
					after, _ := self.driver.PageSource()
					if after == before {
						res.Err = fmt.Errorf("not click :%s", id)
					}
				}()
			}
		}
	} else {
		ele, res.Err = self.SmartMultiFind(id)

	}
	if res.Err != nil {
		return
	}
	if ele == nil {

		return
	}

	if ok, _ := ele.IsEnabled(); ok {
		// ele.MoveTo(0, 0)
		// self.driver.ExecuteScript("arguments[0].scrollIntoView()", []interface{}{ele})
		if res.Err = ele.Click(); res.Err != nil {

			L("found by click error:", self.GetEleInfo(ele))
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
	return
}

func (self *BaseBrowser) OperClickToIdFromEle(i int, id string, args []string, kargs Dict) (res Result) {
	if len(args) == 2 {

		// before, _ := self.driver.PageSource()
		var ele selenium.WebElement
		var eles []selenium.WebElement
		if _, ok := kargs["global"]; ok {
			var ele selenium.WebElement
			ele, res.Err = self.SmartMultiFind(args[1])
			if ele != nil {
				if ok, _ := ele.IsEnabled(); ok {
					before, _ := self.driver.PageSource()

					if res.Err = ele.Click(); res.Err != nil {

						// L("click error")
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
		} else {
			eles, res.Err = self.SmartFindEles(id)
			if len(eles) <= i {
				log.Println(Red("ranger err :", i))
				return
			}
			if i >= len(eles) {
				// res.Err = fmt.Errorf("Err: %s", "beyond ele")
				return
			}
			ele = eles[i]
			_, res.Err = ele.LocationInView()

			if res.Err != nil {
				return
			}
			L("Click Then by", args[1])
			ele, res.Err = self.SmartFindByFromEle(ele, args[1])
			if ele != nil {
				res.Err = ele.Click()
				// L("Click ok")
				self.Sleep()
				if res.Err != nil {
					return
				}
			}
			// time.Sleep(4 * time.Second)

		}

	} else {
		before, _ := self.driver.PageSource()
		var ele selenium.WebElement
		var eles []selenium.WebElement
		eles, res.Err = self.SmartFindEles(id)
		if len(eles) <= i {
			log.Println(Red("ranger err :", i))
			return
		}
		ele = eles[i]

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
	return
}
