package Funcs

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/alyu/configparser"
	"github.com/fatih/color"
	"github.com/tebeka/selenium"
)

/*BaseBrowser help
format:

	$oper : $id : $arg, $arg2, $arg3

actions:
	if:
		# test id's attr value eq value
		if : id , attr, value
		# test id's text value eq testtext
		if : id , testtext

	for:
		# for loop if  id exists
		for: id
			....
		endfor
	click:
		# click id's ele
		click : id
	input:
		# input id's value
		input : id , some text...
	sleep:
		# sleep some second
		sleep: 4
	scroll:
		# scroll to id
		scroll : id
		# scroll to bottom
		scroll
	js :
		# execute js
		js:  alert(some)

*/
type BaseBrowser struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	PageLoadTime int    `json:"timeout"`
	driver       selenium.WebDriver
	service      *selenium.Service
}
type Result struct {
	Text string `json:"text"`
	Err  error  `json:"err"`
}

var (
	DefaultConfPath = *flag.String("conf", "conf.ini", "default conf path")
)

const (
	//Set constants separately chromedriver.exe Address and local call port of
	port = 9515
)

func GenerateBaseConf() {
	fmt.Println(`
[base]
browser = chrome
path = chromedriver.exe
timeout = 4
	`)
}

func NewBaseBrowser() (browser *BaseBrowser, err error) {

	conf, err := configparser.Read(DefaultConfPath)
	if err != nil {
		return
	}
	b, err := conf.Section("base")
	browser = new(BaseBrowser)
	browser.Name = b.ValueOf("browser")
	browser.Path = b.ValueOf("path")
	ti := b.ValueOf("timeout")
	i, _ := strconv.Atoi(ti)
	browser.PageLoadTime = i
	err = browser.Init()
	return

}

func (self *BaseBrowser) Init() error {
	ops := []selenium.ServiceOption{}
	caps := selenium.Capabilities{
		"browserName": self.Name, // "chrome", or any other
		// "phantomjs.binary.path": self.Path, // path to binary from http://phantomjs.org/
	}

	switch self.Name {
	case "phantomjs":
		caps["phantomjs.binary.path"] = self.Path
		// path to binary from http://phantomjs.org/
		driver, err := selenium.NewRemote(caps, "")
		if err != nil {
			return err
		}
		self.driver = driver
	default:
		service, err := selenium.NewChromeDriverService(self.Path, port, ops...)
		if err != nil {
			return err
		}
		self.service = service

		// driver, err := selenium.NewRemote(caps, "")

		driver, err := selenium.NewRemote(caps, "http://127.0.0.1:9515/wd/hub")
		if err != nil {
			return err
		}
		self.driver = driver
	}
	return nil
}

func (self *BaseBrowser) Close() error {
	if err := self.driver.Quit(); err != nil {
		return err
	}
	return self.service.Stop()
}

func (self *BaseBrowser) SmartFindEle(id string) (ele selenium.WebElement, err error) {
	if strings.HasPrefix(id, "/") {
		ele, err = self.driver.FindElement(selenium.ByXPATH, id)
	} else {
		ele, err = self.driver.FindElement(selenium.ByCSSSelector, id)
	}
	return
}

func (self *BaseBrowser) Action(id string, action string, args ...string) (res Result) {
	res = Result{}
	e := color.New(color.FgGreen, color.Bold).SprintFunc()

	defer log.Println(e("["+action+"]"), id, args)
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
		ele, res.Err = self.SmartFindEle(id)

		if res.Err != nil {
			return
		}
		if res.Err = ele.Click(); res.Err != nil {
			return
		}
		res.Text, res.Err = self.driver.PageSource()
	case "input":
		var ele selenium.WebElement
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
	case "sleep":
		if sleepI, err := strconv.Atoi(id); err != nil {
			res.Err = err
		} else {
			time.Sleep(time.Second * time.Duration(sleepI))
		}
	case "scroll":
		if id != "" {
			var ele selenium.WebElement
			ele, res.Err = self.SmartFindEle(id)
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
	case "end":
	case "for":
		var ele selenium.WebElement
		ele, res.Err = self.SmartFindEle(id)
		if res.Err != nil {
			return
		}
		if args != nil {
			res.Text, res.Err = ele.GetAttribute(strings.TrimSpace(args[0]))
		} else {
			res.Text, res.Err = ele.Text()
		}
	default:
		var ele selenium.WebElement
		ele, res.Err = self.SmartFindEle(id)
		if res.Err != nil {
			return
		}
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
	ifjump := false

	oneLine := func(no int, condition bool, action string) (ifJUmp bool, result Result) {
		var oper string
		// defer logrus.Info("Run:")
		action = strings.TrimSpace(action)
		if action == "" {
			return
		}
		if condition {
			if action == "end" {
				return
			}
		}
		if strings.Contains(action, ":") {
			fs := strings.SplitN(strings.TrimSpace(action), ":", 2)
			if strings.Contains(fs[1], ",") {
				args := strings.Split(fs[1], ",")
				oper = strings.TrimSpace(args[0])
				result = self.Action(strings.TrimSpace(args[0]), oper, args[1:]...)
			} else {
				oper = strings.TrimSpace(fs[0])
				result = self.Action(strings.TrimSpace(fs[1]), strings.TrimSpace(fs[0]))
			}
		} else {
			oper = strings.TrimSpace(action)
			result = self.Action("", oper)
		}
		if oper == "if" {
			if result.Err != nil {
				ifJUmp = true
			}
		}

		return
	}
	forstack := []string{}
	forcondition := ""
	forloop := false
	for no, action := range lines {
		if forloop {
			if strings.TrimSpace(action) == "endfor" {
				forloop = false
				continue
			} else {
				forstack = append(forstack, action)
				continue
			}
		}
		if len(forstack) > 0 {
			for {
				subjump := false
				for subno, subaction := range forstack {
					subjump, last = oneLine(subno, subjump, subaction)
				}
				_, last := oneLine(-1, false, forcondition)
				if last.Err != nil {
					log.Println("[ENDFOR]:", last.Err)
					forstack = []string{}
					break
				}
			}
		} else {
			if strings.HasPrefix(strings.TrimSpace(action), "for:") {
				forloop = true
				forcondition = strings.TrimSpace(action)
				log.Println("[FOR]:", forcondition)
				continue
			} else {
				ifjump, last = oneLine(no, ifjump, action)
				if last.Err != nil && !strings.Contains(last.Err.Error(), "Timed out receiving message from renderer") {
					log.Println(no, action, last.Err)
				}

			}
		}

	}
}

func (self *BaseBrowser) ExecuteByFile(template string) (err error) {
	buf, err := ioutil.ReadFile(template)
	if err != nil {
		return
	}
	self.Parse(string(buf))
	// defer fp.Close()
	return
}
