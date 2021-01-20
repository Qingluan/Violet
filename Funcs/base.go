package Funcs

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	Name            string `json:"name"`
	Path            string `json:"path"`
	PageLoadTime    int    `json:"timeout"`
	Mode            int
	lines           []string
	forstack        [][]string
	forLoopTestArgs []string
	loopCount       int
	loopNow         int
	TmpPage         string
	TmpID           string
	driver          selenium.WebDriver
	tmpEles         []selenium.WebElement
	service         *selenium.Service
	OperStack       []Stack
	// OperStackWhere  [][]int
	// OperNow         string
	NextLine int
}
type Result struct {
	Action string `json:"action"`
	Text   string `json:"text"`
	Err    error  `json:"err"`
	Int    int    `json:"int"`
	Bool   bool   `json:"bool"`
}

var (
	Green           = color.New(color.FgGreen).SprintFunc()
	Bold            = color.New(color.Bold).SprintFunc()
	Red             = color.New(color.FgRed).SprintFunc()
	DefaultConfPath = *flag.String("conf", "conf.ini", "default conf path")
	MODE_FLOW       = 0
	MODE_FOR        = 1
	MODE_EACH       = 2
	MODE_IF         = 3
	MODE_FOR_RUN    = 4
	MODE_EACH_RUN   = 5
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

func (self *BaseBrowser) SmartFindEles(id string) (ele []selenium.WebElement, err error) {
	if strings.HasPrefix(id, "/") {
		ele, err = self.driver.FindElements(selenium.ByXPATH, id)
	} else {
		ele, err = self.driver.FindElements(selenium.ByCSSSelector, id)
	}
	return
}

func L(action string, args ...interface{}) {
	e := color.New(color.FgGreen, color.Bold).SprintFunc()
	pre := e("[" + action + "]")
	// log.Println(args...)

	log.Printf("[%s] : %v", pre, args)
}

func (self *BaseBrowser) EleToJsonString(s *goquery.Selection, key ...string) string {
	ss := make(map[string]string)
	ss["text"] = s.Text()
	ss["html"], _ = s.Html()
	for _, k := range key {
		ss[k] = s.AttrOr(k, "")
	}
	res, _ := json.Marshal(&ss)
	return string(res)
}

func (self *BaseBrowser) ActionEle(id string, page []byte, stacks [][]string) (res Result) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(page))
	if err != nil {
		res.Err = err
		return
	}
	doc.Find(id).Each(func(i int, s *goquery.Selection) {
		for _, args := range stacks {
			argc := len(args)
			if argc < 2 {
				continue
			}
			// id := args[1]
			main := args[0]

			switch main {
			case "save":
				switch argc {
				case 1:
					fp, err := os.OpenFile(args[1], os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
					if err != nil {
						res.Err = err
						return
					}
					defer fp.Close()
					_, res.Err = fp.WriteString(self.EleToJsonString(s, args[1:]...) + "\n")
					// for _, ele := range eles {
					// 	_, res.Err = fp.WriteString(ele.Text())
					// }
				}
			case "print":
				L("show", self.EleToJsonString(s, args[1:]...))
			case "find":
				s.Find(args[1]).Each(func(i int, s *goquery.Selection) {
					L("find", args[1], s.Text())
				})
			}
		}
	})

	return
}

func (self *BaseBrowser) oneLine(no int, ifInCondition bool, args []string) (ifJUmp bool, result Result) {

	argc := len(args)
	if argc == 0 {
		return
	}
	main := args[0]
	if ifInCondition {
		if main == "end" {
			return
		}
	}
	if main == "" {
		L("All is Empty")
		return
	}
	if argc > 2 {
		result = self.Action(args[1], main, args[2:]...)
	} else if argc > 1 {
		result = self.Action(args[1], main)
	} else {
		result = self.Action("", main)
	}
	if main == "if" {
		if result.Err != nil {
			ifJUmp = true
		}
	}

	return
}

func (self *BaseBrowser) ConsoleLog(result Result) {
	if result.Err != nil {
		if result.Err != nil && !strings.Contains(result.Err.Error(), "Timed out receiving message from renderer") {
			// log.Println(no, action, result.Err)
			log.Println(color.New(color.FgHiRed).SprintFunc()("[ENDFOR]"), ":", result.Err)
			self.switchMode()
		}

		// break
	} else {
		log.Println(Green("\t[result:", result.Action, "]"), ":\n", Bold(result.Text, "---------------------------------------"))
	}
}

func (self *BaseBrowser) RunForStack() {
	self.loopNow = 0
	// var last Result
	for {
		subjump := false
		for subno, subargs := range self.forstack {
			subjump, _ = self.oneLine(subno, subjump, subargs)
		}
		_, last := self.oneLine(-1, false, self.forLoopTestArgs)
		if self.loopCount != 0 && self.loopNow >= self.loopCount {
			log.Println("[ENDFOR]:", last.Err)
			self.switchMode()
			break
		} else {
			if last.Bool {
				self.switchMode()
				break
			}
		}
		self.ConsoleLog(last)
		self.loopNow++
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

func splitArgsTrim(raw string) (as []string) {
	for _, w := range strings.SplitN(raw, ":", 2) {
		as = append(as, strings.TrimSpace(w))
	}
	if len(as) > 1 {
		if strings.Contains(as[1], ":") {
			argsStr := as[1]
			as = []string{as[0]}
			for _, w2 := range strings.Split(argsStr, ",") {
				as = append(as, strings.TrimSpace(w2))
			}
		}
	}
	return
}

func (self *BaseBrowser) switchMode(modes ...int) {
	ac := MODE_FLOW
	if modes != nil {
		ac = modes[0]
	}
	switch ac {
	case MODE_FLOW:
		self.forstack = [][]string{}
		self.forLoopTestArgs = []string{}
		self.TmpID = ""
		self.TmpPage = ""
	}
	self.Mode = ac

}
