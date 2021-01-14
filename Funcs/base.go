package Funcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

type BaseBrowser struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	PageLoadTime int    `json:"timeout"`
	driver       selenium.WebDriver
}
type Result struct {
	Text string `json:"text"`
	Err  error  `json:"err"`
}

func (self *BaseBrowser) Init() error {
	switch self.Name {
	case "phantomjs":
		caps := selenium.Capabilities{
			"browserName":           self.Name, // "chrome", or any other
			"phantomjs.binary.path": self.Path, // path to binary from http://phantomjs.org/
		}
		driver, err := selenium.NewRemote(caps, "")
		if err != nil {
			return err
		}
		self.driver = driver
	}
	return nil
}

func (self *BaseBrowser) Action(id string, action string, args ...string) (res Result) {
	res = Result{}
	switch action {
	case "get":
		if args != nil {
			self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
			self.driver.Get(strings.TrimSpace(args[0]))
			res.Text, res.Err = self.driver.PageSource()
			if res.Err != nil {
				return
			}
		} else if strings.HasPrefix(id, "http") {
			self.driver.SetPageLoadTimeout(time.Second * time.Duration(self.PageLoadTime))
			self.driver.Get(id)
			res.Text, res.Err = self.driver.PageSource()
			if res.Err != nil {
				return
			}
		} else {
			res.Err = fmt.Errorf("need url !!! %s", id)
		}
	case "click":
		var ele selenium.WebElement
		ele, res.Err = self.driver.FindElement(selenium.ByCSSSelector, id)

		if res.Err != nil {
			return
		}
		if res.Err = ele.Click(); res.Err != nil {
			return
		}
		res.Text, res.Err = self.driver.PageSource()
	case "input":
		var ele selenium.WebElement
		ele, res.Err = self.driver.FindElement(selenium.ByCSSSelector, id)

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
	case "scroll":
		if id != "" {
			var ele selenium.WebElement
			ele, res.Err = self.driver.FindElement(selenium.ByCSSSelector, id)
			if res.Err != nil {
				return
			}
			_, res.Err = ele.LocationInView()
			// self.driver.ExecuteScript("arguments[0].scrollIntoView(true);", ele)
		} else {
			self.driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil)
		}
	case "js":
		if args != nil {
			_, res.Err = self.driver.ExecuteScript(strings.TrimSpace(args[0]), nil)
		} else {
			res.Err = fmt.Errorf("no args to execute js")
		}
	case "savescreen":
	case "if":
		var ele selenium.WebElement
		ele, res.Err = self.driver.FindElement(selenium.ByCSSSelector, id)
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
	case "end":
	default:
		var ele selenium.WebElement
		ele, res.Err = self.driver.FindElement(selenium.ByCSSSelector, id)
		if args != nil {
			res.Text, res.Err = ele.GetAttribute(strings.TrimSpace(args[0]))
		} else {
			res.Text, res.Err = ele.Text()
		}
	}
	return Result{}
}

func (self *BaseBrowser) Parse(actions string) {
	lines := strings.Split(actions, "\n")
	var last Result
	for no, action := range lines {
		var result Result
		if strings.Contains(action, ":") {
			fs := strings.SplitN(strings.TrimSpace(action), ":", 2)
			if strings.Contains(fs[1], ",") {
				args := strings.Split(fs[1], ",")
				result = self.Action(strings.TrimSpace(args[0]), strings.TrimSpace(fs[0]), args[1:]...)
			} else {
				result = self.Action(strings.TrimSpace(fs[1]), strings.TrimSpace(fs[0]))
			}
		} else {
			result = self.Action("", strings.TrimSpace(action))
		}

	}
}
