package Funcs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	jupyter "github.com/Qingluan/jupyter/http"
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

func (self *BaseBrowser) OperPush(id string, args []string, kargs Dict) (res Result) {
	server, ok := kargs["server"]

	_, useIframe := kargs["iframe"]
	attrs, hasattrs := kargs["attrs"]
	if !ok {
		L("must set kargs['server'] ")
		return
	}
	d, ok := kargs["ding"]
	if ok {
		self.PushAPI = d.(string)
	}
	if id != "" {
		var eles []selenium.WebElement
		if useIframe {
			eles, res.Err = self.SmartFindEles(id, "")
		} else {
			eles, res.Err = self.SmartFindEles(id)
		}
		page := ""
		for _, ele := range eles {
			ss := make(map[string]string)

			ss["html"], _ = ele.GetAttribute("outerHTML")
			ss["text"], _ = ele.Text()
			if hasattrs {

				switch attrs.(type) {
				case []string:
					for _, k := range attrs.([]string) {
						if k == "text" {
							ss["text"], _ = ele.Text()
						} else {
							ss[k], _ = ele.GetAttribute(k)
						}
					}
				case string:
					k := attrs.(string)
					if k == "text" {
						ss["text"], _ = ele.Text()
					} else {
						ss[k], _ = ele.GetAttribute(k)
					}
				}

			}
			res, _ := json.Marshal(&ss)
			page += string(res) + "\n"

		}
		push(server.(string), page, "json")
	} else {
		page, _ := self.driver.PageSource()
		push(server.(string), page, "html")
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

func (self *BaseBrowser) EleToJson(ele selenium.WebElement, attrs ...string) string {
	ss := make(map[string]string)
	// subele , _ := self.SmartFindByFromEle(ele,subid)
	ss["html"], _ = ele.GetAttribute("outerHTML")
	ss["text"], _ = ele.Text()
	for _, k := range attrs {
		if k == "text" {
			ss["text"], _ = ele.Text()
		} else {
			ss[k], _ = ele.GetAttribute(k)
		}
	}
	res, _ := json.Marshal(&ss)
	return string(res)
}

func (self *BaseBrowser) OperCollectSingle(id string, args []string, kargs Dict) (res Result) {
	// var filePath string
	filePath, ok := kargs["output"]
	if !ok {
		res.Err = fmt.Errorf("no 'output' to specify a file path !")
		return
	}

	fp, err := os.OpenFile(filePath.(string), os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		res.Err = err
		return
	}
	defer fp.Close()
	var eles []selenium.WebElement
	// var ok bool

	_, useIframe := kargs["iframe"]
	if useIframe {
		eles, res.Err = self.SmartFindEles(id, "")
	} else {
		eles, res.Err = self.SmartFindEles(id)
	}
	index := -1
	if indexstr, ok := kargs["index"]; ok {
		index, _ = strconv.Atoi(indexstr.(string))
	}
	do := func(ele selenium.WebElement, attrs []string, kargs Dict) {

		if multi, ok := kargs["multi"]; ok {
			sss := make(map[string]interface{})
			arrs := []string{}
			switch multi.(type) {
			case []string:

				for _, subid := range multi.([]string) {
					if strings.Contains(subid, ":") {
						fs := strings.SplitN(subid, ":", 2)
						id := strings.TrimSpace(fs[0])
						attr := strings.TrimSpace(fs[1])

						subele, _ := self.SmartFindByFromEle(ele, id)
						arrs = append(arrs, self.EleToJson(subele, attr))
					} else {
						subele, _ := self.SmartFindByFromEle(ele, subid)
						arrs = append(arrs, self.EleToJson(subele))
					}
				}
			case string:
				subid := multi.(string)
				if strings.Contains(subid, ":") {

					fs := strings.SplitN(subid, ":", 2)
					id := strings.TrimSpace(fs[0])
					attr := strings.TrimSpace(fs[1])
					subele, _ := self.SmartFindByFromEle(ele, id)
					arrs = append(arrs, self.EleToJson(subele, attr))
				} else {
					subele, _ := self.SmartFindByFromEle(ele, subid)
					arrs = append(arrs, self.EleToJson(subele))
				}
				// fp.WriteString(self.EleToJson(subele) + "\n")
			}
			sss["multi"] = arrs
			sss["html"], _ = ele.GetAttribute("outerHTML")
			sss["text"], _ = ele.Text()
			resStr, _ := json.Marshal(sss)
			fp.WriteString(string(resStr) + "\n")
		} else {
			ss := make(map[string]string)
			if len(attrs) == 0 {
				ss["html"], _ = ele.GetAttribute("outerHTML")
				ss["text"], _ = ele.Text()
			} else {

				for _, k := range attrs {
					if k == "text" {
						ss["text"], _ = ele.Text()
					} else {
						ss[k], _ = ele.GetAttribute(k)
					}
				}

			}
			res, _ := json.Marshal(&ss)
			// return string(res)
			fp.WriteString(string(res) + "\n")

		}

	}
	for n, ele := range eles {
		if attrs, ok := kargs["attrs"]; ok {
			switch attrs.(type) {
			case []string:
				do(ele, attrs.([]string), kargs)
			case string:
				do(ele, []string{attrs.(string)}, kargs)
			}
		} else {

			do(ele, []string{}, kargs)
			// _, res.Err = fp.WriteString(self.EleToJsonString(sb) + "\n")
		}
		if n == index {
			break
		}
	}
	return

}

func push(server, page, tp string) {
	sess := jupyter.NewSession()
	sess.Json(server, map[string]interface{}{
		"page": page,
		"tp":   tp,
	})
}
