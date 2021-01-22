package Funcs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

func (self *BaseBrowser) Action(id string, action string, kargs Dict, args ...string) (res Result) {
	res = Result{
		Action: action,
	}
	if action == "" {
		res.Action = id
	}
	// defer func() {
	pre := ""
	for range self.OperStack {
		pre += "\t"
	}
	log.Println(Green(pre+"["+action+"]"), id, args, kargs)
	// }()
	if !strings.HasPrefix(action, "#") {
		defer time.Sleep(2 * time.Second)
	}

	switch action {
	case "get":
		if args != nil {
			self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
			self.driver.Get(strings.TrimSpace(args[0]))
			// res.Text, res.Err = self.driver.PageSource()
			// if res.Err != nil {
			// 	return
			// }
		} else if strings.HasPrefix(id, "http") {
			self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
			self.driver.Get(id)
			// res.Text, res.Err = self.driver.PageSource()
			// if res.Err != nil {
			// 	return
			// }
		} else {
			res.Err = fmt.Errorf("need url !!! %s", id)
		}
	case "click":
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
		// res.Text, res.Err = self.driver.PageSource()
	case "back":
		res.Err = self.driver.Back()

	case "input":
		var ele selenium.WebElement

		if name, ok := kargs["name"]; ok {
			// var ele selenium.WebElement
			var pwdele selenium.WebElement

			pwdele, res.Err = self.SmartFindEle("//input[@type=\"password\"]")
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
				form, err := self.SmartFindEle("//input[@type=\"password\"]/" + pa + "/form")
				if err == nil && form != nil {
					break
				}
				i++
			}
			ele, res.Err = self.SmartFindEle("//input[@type=\"password\"]/" + pa + "/input[@type=\"text\"]")
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
					case "\n":
						ele, res.Err = self.SmartFindEle("//input[@type=\"password\"]/" + pa + "/*[@type=\"submit\"]")
						if res.Err != nil {
							return
						}

						if ele != nil {

							L("Submit -- >", pwd.(string))
							ele.Click()
						}
					case "\t":
						ele.SendKeys(selenium.TabKey)
					}
				}
			}

			// ele = ele.FindElement(selenium.By)
		} else {
			ele, res.Err = self.SmartFindEle(id)
			if res.Err != nil {
				return
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

	case "wait":
		sleep := 10
		if tk, ok := kargs["sleep"]; ok {
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
				if util, ok := kargs["change"]; ok {
					if util.(string) == "url" {
						// L("Wait Url Change", "timeout:", sleep)
						thisURL, _ := self.driver.CurrentURL()
						if thisURL != lastURL {
							break
						}
					}
				}
			}
			if time.Now().After(timeout) {
				break
			}
			time.Sleep(500 * time.Millisecond)
			res.Err = nil
		}
		res.Err = nil
	case "refresh":
		self.driver.Refresh()
	case "load":
		if id != "" {
			fs, err := ioutil.ReadFile(id)
			if err == nil {
				d := make(Dict)
				if err := json.Unmarshal(fs, &d); err == nil {
					kargs = d
				}
			}
		}
		if newurl, ok := kargs["url"]; ok {
			self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
			res.Err = self.driver.Get(strings.TrimSpace(newurl.(string)))

		}

		if cookie, ok := kargs["cookie"]; ok {
			u, err := self.driver.CurrentURL()
			if err != nil {
				res.Err = err
				return
			}
			urlObj, err := url.Parse(u)
			if err != nil {
				res.Err = err
				return
			}
			cookies, err := parseCookie(cookie.(string), urlObj.Host)
			self.driver.DeleteAllCookies()
			for _, c := range cookies {

				err := self.driver.AddCookie(c)
				if err != nil {
					L("----- Add Cookie err ----", Yellow(err))
				}
				L("---- Add Cookie ----", c.Name, " : ", c.Value, " in :", c.Domain)
			}
			self.driver.ExecuteScriptRaw("alert(document.cookie)", nil)
			self.driver.Refresh()
		}
		// if header, ok := kargs[""]

	case "scroll":
		if id != "" {
			var ele selenium.WebElement
			ele, res.Err = self.SmartFindEle(id)
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
			self.driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil)
		}
	case "js":
		if args != nil {
			_, res.Err = self.driver.ExecuteScript(strings.TrimSpace(args[0]), nil)
		} else if strings.TrimSpace(id) != "" {
			_, res.Err = self.driver.ExecuteScript(strings.TrimSpace(id), nil)
		} else {
			res.Err = fmt.Errorf("no args to execute js")
		}
	case "save":
		if id == "" {
			return
		}
		// fp, err := os.OpenFile(id, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		fp, err := os.Create(id)

		if err != nil {
			res.Err = err
			return
		}
		defer fp.Close()
		if _, ok := kargs["cookie"]; ok {
			if cs, err := self.driver.GetCookies(); err != nil {
				res.Err = err
			} else {
				msg := ""
				for _, c := range cs {
					msg += fmt.Sprintf("%s=%s; ", c.Name, c.Value)
				}
				d := make(Dict)
				d["cookie"] = url.QueryEscape(strings.TrimSpace(msg))
				d["url"], _ = self.driver.CurrentURL()
				buf, _ := json.MarshalIndent(&d, "", "    ")
				fp.Write(buf)
			}
		}
	case "savescreen":
		if id != "" {
			buf, err := self.driver.Screenshot()
			if err != nil {
				res.Err = err
				return
			}
			fp, err := os.Create(id)
			if err != nil {
				res.Err = err
				return
			}
			defer fp.Close()
			fp.Write(buf)
		}

	case "each":
		fmt.Println(id)
		if id == "" {

			return
		}
		// time.Sleep(1 * time.Second)
		_, res.Err = self.SmartFindEle(id)
		if res.Err == nil {
			res.Bool = true
		}
		// res.Bool = true
	case "if":
		var ele selenium.WebElement
		ele, res.Err = self.SmartFindEle(id)
		if res.Err != nil {

			return
		}
		if args != nil {
			if len(args) == 2 {
				res.Text, res.Err = ele.GetAttribute(strings.TrimSpace(args[0]))
				if res.Text != strings.TrimSpace(args[1]) {
					res.Err = fmt.Errorf("not eq")
				}
			} else {
				res.Text, res.Err = ele.Text()
				if res.Text != strings.TrimSpace(args[0]) {
					res.Err = fmt.Errorf("not eq")
				}
			}
		}
		if res.Err == nil {
			res.Err = nil
			res.Bool = true
		}
	case "end":
	case "for":
		idint, iterr := strconv.Atoi(id)
		stack := self.GetStack()
		stack.NowLoop++
		nowloop := stack.NowLoop
		self.OperStack[len(self.OperStack)-1] = Stack{
			NowLoop: nowloop,
			Start:   stack.Start,
			End:     stack.End,
			Oper:    stack.Oper,
		}
		// nowloop := self.OperStack[len(self.OperStack)-1].NowLoop
		if iterr == nil {
			// res.Int = idint
			if nowloop <= idint {
				L("Loop", "Total:", nowloop, "/", idint)
				res.Bool = true
				return
			} else {

			}
			// self.loopCount = idint

			return
		}
		// var ele selenium.WebElement
		// var err error
		_, res.Err = self.SmartFindEle(id)
		if res.Err == nil {
			res.Bool = true
			return
		} else {
			return
		}
		// if args != nil {
		// 	res.Text, res.Err = ele.GetAttribute(strings.TrimSpace(args[0]))
		// } else {
		// 	res.Text, res.Err = ele.Text()
		// }
		// if res.Err == nil {
		// 	res.Err = nil
		// 	res.Bool = true
		// 	// return
		// }
	default:
		res.Err = fmt.Errorf("illegal code action:%s  args:%v", action, args)
		// var ele selenium.WebElement
		// ele, res.Err = self.SmartFindEle(id)
		// if res.Err != nil {
		// 	return
		// }
		// if args != nil {
		// 	res.Text, res.Err = ele.GetAttribute(strings.TrimSpace(args[0]))
		// } else {
		// 	res.Text, res.Err = ele.Text()
		// }
	}
	return
}