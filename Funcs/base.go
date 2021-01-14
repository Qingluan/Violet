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
			self.driver.Get(args[0])
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
	case "":
		var ele selenium.WebElement
		ele, res.Err = self.driver.FindElement(selenium.ByCSSSelector, id)
		if args != nil {
			res.Text, res.Err = ele.GetAttribute(args[0])
		} else {
			res.Text, res.Err = ele.Text()
		}
	case "js":
		if args != nil {
			_, res.Err = self.driver.ExecuteScript(args[0], nil)
		} else {
			res.Err = fmt.Errorf("no args to execute js")
		}
	case "savescreen":
	case "if":

	}
	return Result{}
}
