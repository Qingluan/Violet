package Funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
)

func (self *BaseBrowser) GetEleInfo(ele selenium.WebElement) string {
	name, _ := ele.TagName()
	if a, _ := ele.GetAttribute("id"); strings.TrimSpace(a) != "" {
		name += fmt.Sprintf("#%s", strings.TrimSpace(a))
	}

	if a, _ := ele.GetAttribute("class"); strings.TrimSpace(a) != "" {
		name += fmt.Sprintf(".%s", strings.TrimSpace(a))
	}
	bd, _ := ele.IsDisplayed()
	be, _ := ele.IsEnabled()
	bs, _ := ele.IsSelected()
	name += fmt.Sprintf(" ( display:%v enable:%v select:%v", bd, be, bs)
	l, _ := ele.Location()
	name += fmt.Sprintf(" x:%d y:%d)", l.X, l.Y)
	return name
}

func (self *BaseBrowser) GetParent(ele selenium.WebElement) selenium.WebElement {
	e, _ := ele.FindElement(selenium.ByXPATH, "./..")
	return e
}

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
	printmsg := Greenb(pre+"["+action+"]") + Blueb(fmt.Sprintf(" (id: %s)", id))
	if len(args) != 0 {
		printmsg += Cyanb(strings.ReplaceAll(fmt.Sprintf("Args: %v ", args), "\n", ""))
	}
	if len(kargs) != 0 {
		printmsg += "\n"
		for k, v := range kargs {
			printmsg += Yellow(fmt.Sprintf("  %s: %v", k, v))
		}
	}

	if self.PageLoadTime == 0 {
		self.PageLoadTime = 5
	}
	fmt.Println(printmsg)
	// }()
	if !strings.HasPrefix(action, "#") {
		defer time.Sleep(2 * time.Second)
	}

	switch action {
	case "get":
		if args != nil {
			if timeout, ok := kargs["timeout"]; ok {
				t, _ := strconv.Atoi(timeout.(string))
				L("set timeout =>", t)
				self.driver.SetPageLoadTimeout(time.Second * time.Duration(t))
			} else {
				if self.PageLoadTime > 0 {
					self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
				}
			}
			fmt.Println("Dos", args)
			res.Err = self.driver.Get(strings.TrimSpace(args[0]))
			// res.Text, res.Err = self.driver.PageSource()
			// if res.Err != nil {
			// 	return
			// }
		} else if strings.HasPrefix(id, "http") {

			// fmt.Println("Do", id, self.PageLoadTime)
			if self.PageLoadTime > 0 {
				res.Err = self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
			}
			res.Err = self.driver.Get(id)
			// res.Text, res.Err = self.driver.PageSource()
			// if res.Err != nil {
			// 	return
			// }
		} else {
			res.Err = fmt.Errorf("need url !!! %s", id)
		}

	case "click":
		res = self.OperClickToId(id, args, kargs)
		// res.Text, res.Err = self.driver.PageSource()
	case "back":
		res.Err = self.driver.Back()

	case "input":
		res = self.OperInput(id, args, kargs)
	case "wait":
		res = self.OperWait(id, args, kargs)
	case "refresh":
		self.driver.Refresh()
	case "export":
		spaceA := regexp.MustCompile(`\s+`)
		html, _ := self.driver.PageSource()
		// fmt.Println(html)
		doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(html)))
		if err != nil {
			res.Err = err
			return
		}
		columns := [][]string{}
		mins := 9999999
		for _, arg := range args {
			column := []string{}
			doc.Find(strings.TrimSpace(arg)).Each(func(i int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				text = strings.ReplaceAll(text, "\n", "")
				text = strings.ReplaceAll(text, "\t", " ")
				text = spaceA.ReplaceAllString(text, " ")
				column = append(column, text)
			})
			if len(column) < mins {
				mins = len(column)
			}
			columns = append(columns, column)
		}
		fb, err := os.OpenFile(strings.TrimSpace(id), os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		defer fb.Close()
		for i := 0; i < mins; i++ {
			line := []string{}
			for _, column := range columns {
				line = append(line, column[i])
			}
			if i < 10 {
				fmt.Println(Yellow(strings.Join(line, ",")) + "\n")
			} else if i == 10 {
				fmt.Println("        ....    ")
			}
			fb.WriteString(strings.Join(line, ",") + "\n")
		}

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
		if header, ok := kargs["header"]; ok {
			if strings.Contains(header.(string), "=") {
				// self.driver.AddCookie()
			}
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
			// self.driver.DeleteAllCookies()
			// cookieStr := ""
			for _, c := range cookies {

				// err := self.driver.AddCookie(c)
				// cookieStr += fmt.Sprintf("%s=%s; ", c.Name, c.Value)
				// if err != nil {
				// 	L("----- Add Cookie err ----", Yellow(err))
				// }
				L("---- Add Cookie ----", c.Name, " : ", c.Value, " in :", c.Domain)
			}
			self.driver.ExecuteScriptRaw(fmt.Sprintf("document.cookie=\"%s\"", cookie.(string)), nil)
			self.driver.Refresh()
		}
		// if header, ok := kargs[""]

	case "scroll":
		res = self.OperScrollTo(id, args, kargs)
	case "js":
		var resOut interface{}
		if args != nil {
			resOut, res.Err = self.driver.ExecuteScript(strings.TrimSpace(args[0]), nil)
		} else if strings.TrimSpace(id) != "" {
			resOut, res.Err = self.driver.ExecuteScript(strings.TrimSpace(id), nil)
		} else {
			res.Err = fmt.Errorf("no args to execute js")
		}
		if resOut != nil {
			L("Js Out:", resOut)
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
		_, res.Err = self.SmartMultiFind(id)
		if res.Err == nil {
			res.Bool = true
		}
		// res.Bool = true
	case "if":
		var ele selenium.WebElement
		ele, res.Err = self.SmartMultiFind(id)
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
		operReverse := false
		if _, ok := kargs["not"]; ok {
			operReverse = true
		}

		_, res.Err = self.SmartMultiFind(id)
		if operReverse && res.Err != nil {
			res.Err = nil
			res.Bool = true
		} else if !operReverse && res.Err == nil {
			res.Bool = true
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

	case "search":
		var eles []selenium.WebElement
		if _, ok := kargs["pipe"]; ok {
			var ele selenium.WebElement
			ele, res.Err = self.SmartMultiFind(id)
			if ele != nil {
				res.Text = self.GetEleInfo(ele)
			}
			return
		}
		eles, res.Err = self.SmartFindEles(id)
		for _, e := range eles {
			var t string
			if _, ok := kargs["text"]; ok {
				t, res.Err = e.Text()
				res.Text += "\n" + t
				continue
			}

			if _, ok := kargs["in"]; ok {

				t, res.Err = e.GetAttribute("innerHTML")
				if res.Err != nil {
					break
				} else {
					res.Text += "\n" + t
				}
				continue
			}
			if _, ok := kargs["out"]; ok {
				t, res.Err = e.GetAttribute("outerHTML")
				if res.Err != nil {
					break
				} else {
					res.Text += "\n" + t
				}
				continue
			}

			// tagName, _ := e.TagName()
			res.Text += "\n" + self.GetEleInfo(e)
			if text, _ := e.Text(); strings.TrimSpace(text) != "" {
				res.Text += "\n\t text:" + strings.TrimSpace(text)
			}
		}
	case "print":

		if _, ok := kargs["all"]; ok {
			res.Text, _ = self.driver.PageSource()
			res.Text = removeScriptAndCss(res.Text)
			return
		}
		if _, ok := kargs["cookie"]; ok {
			cookies, err := self.driver.GetCookies()
			if err != nil {
				res.Err = err
				return
			}
			for _, cookie := range cookies {
				if cookie.Domain == self.CurrentDomain() {
					res.Text += fmt.Sprintf("\n%s:%s", cookie.Name, cookie.Value)
				}
			}
			// for _, header := range self.driver.
			return
		}
		var ele selenium.WebElement
		ele, res.Err = self.SmartMultiFind(id, true)
		if res.Err != nil {
			L("err", res.Err)
			return
		}
		if ele == nil {
			return
		}
		// L("Found Ele:", ele)
		if _, ok := kargs["text"]; ok {
			res.Text, res.Err = ele.Text()
			return
		}

		if _, ok := kargs["in"]; ok {

			res.Text, res.Err = ele.GetAttribute("innerHTML")

		} else {
			res.Text, res.Err = ele.GetAttribute("outerHTML")

		}
		// var tag string
		// var text string
		// tag, res.Err = ele.TagName()
		// text, res.Err = ele.Text()
		// L("Show", attr)

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
