package Funcs

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sort"
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
	Proxy           string
	Ua              string
	driver          selenium.WebDriver
	loger           io.WriteCloser
	log             string
	tmpEles         []selenium.WebElement
	service         CloserService
	OperStack       []Stack
	// OperStackWhere  [][]int
	// OperNow         string
	option_casp   selenium.Capabilities
	option_prefix string
	PushAPI       string
	NextLine      int
}
type Result struct {
	Action string `json:"action"`
	Text   string `json:"text"`
	Err    error  `json:"err"`
	Int    int    `json:"int"`
	Bool   bool   `json:"bool"`
}

type CloserService interface {
	Stop() error
}

var (
	Green  = color.New(color.FgGreen).SprintFunc()
	Cyanb  = color.New(color.FgWhite, color.BgCyan).SprintFunc()
	Cyan   = color.New(color.FgCyan).SprintFunc()
	Greenb = color.New(color.FgBlack, color.BgGreen).SprintFunc()
	Blueb  = color.New(color.FgHiWhite, color.BgBlue).SprintFunc()
	Blue   = color.New(color.FgBlue).SprintFunc()

	Bold            = color.New(color.Bold).SprintFunc()
	Red             = color.New(color.FgRed).SprintFunc()
	Yellow          = color.New(color.FgYellow).SprintFunc()
	Yellowb         = color.New(color.BgYellow).SprintFunc()
	DefaultConfPath = *flag.String("conf", "conf.ini", "default conf path")

	MODE_FLOW     = 0
	MODE_FOR      = 1
	MODE_EACH     = 2
	MODE_IF       = 3
	MODE_FOR_RUN  = 4
	MODE_EACH_RUN = 5
)

const (
	//Set constants separately chromedriver.exe Address and local call port of
	port = 9515
)

func GenerateBaseConf() {
	fmt.Println(`
[base]
browser = chrome
path = chromedriver
#path = chromedriver.exe
timeout = 4
UA = Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36
	
	`)
}

func NewBaseBrowser() (browser *BaseBrowser, err error) {

	conf, err := configparser.Read(DefaultConfPath)
	if err != nil {
		return
	}
	b, err := conf.Section("base")
	if err != nil {
		return
	}
	browser = new(BaseBrowser)
	browser.Name = b.ValueOf("browser")
	browser.Path = b.ValueOf("path")
	browser.Proxy = b.ValueOf("proxy")
	browser.log = b.ValueOf("log")
	browser.Ua = b.ValueOf("UA")
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
	prefixURL := fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
	// L(self.Name)
	switch self.Name {
	case "phantomjs":

		caps["phantomjs.binary.path"] = self.Path
		if self.Ua != "" {
			caps["phantomjs.page.settings.userAgent"] = self.Ua
		} else {
			caps["phantomjs.page.settings.userAgent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.146 Safari/537.3"
		}
		// selenium.SetDebug(true)
		if self.Proxy != "" {
			L("Use Proxy:", self.Proxy)
		}
		// if sys.
		// fmt.Println("do some")

		prefixURL := fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port+1)
		var err error
		if self.log != "" {
			self.loger, err = os.OpenFile(self.log, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
			if err != nil {
				return err
			}
		}
		// fmt.Println("do some")
		service, err := self.PhantomJSService(port+1, self.Proxy)
		// fmt.Println("do some --")
		if err != nil {
			panic(err)
		}
		self.service = service
		L("connected ", prefixURL)
		driver, err := selenium.NewRemote(caps, prefixURL)
		if err != nil {
			log.Fatal(err)
			// driver.Get("https://www.google.com")
			// return err
		}
		// self.driver.
		self.driver = driver

	default:

		// fmt.Println("do some")
		caps["acceptInsecureCerts"] = true
		caps["args"] = []string{
			//https://peter.sh/experiments/chromium-command-line-switches/
			"--disable-gpu",
			"--disable-web-security",
			"--headless",
			"--ignore-certificate-errors",
			"--log-level=0", // INFO = 0, WARNING = 1, LOG_ERROR = 2, LOG_FATAL = 3
			"--no-sandbox",
			"--window-size=1024x768",
		}
		if self.Proxy != "" {

			if strings.HasPrefix(self.Proxy, "socks5://") {
				p := strings.TrimLeft(self.Proxy, "socks5://")
				hostFields := strings.SplitN(p, ":", 2)
				port2, err := strconv.Atoi(hostFields[1])
				if err != nil {
					log.Fatal("Port Err:", err)
				}
				caps.AddProxy(selenium.Proxy{
					Type:         selenium.Manual,
					SOCKS:        p,
					SOCKSVersion: 5,
				})

				L("Use Proxy:", hostFields[0], port2)
			} else {
				log.Fatal("Only support proxy : socks5://ip:port")
			}
		}
		// caps := chrome.Capabilities{
		// 	Args: []string{ // https://peter.sh/experiments/chromium-command-line-switches/
		// 		"--disable-gpu",
		// 		"--disable-web-security",
		// 		"--headless",
		// 		"--ignore-certificate-errors",
		// 		// "--log-level=2", // INFO = 0, WARNING = 1, LOG_ERROR = 2, LOG_FATAL = 3
		// 		// "--no-sandbox",
		// 		"--window-size=1024x768",
		// 	},
		// }
		// }

		service, err := selenium.NewChromeDriverService(self.Path, port, ops...)
		if err != nil {
			return err
		}
		self.service = service

		// driver, err := selenium.NewRemote(caps, "")
		caps.SetLogLevel("browser", "ALL")
		driver, err := selenium.NewRemote(caps, prefixURL)

		if err != nil {
			return err
		}
		self.driver = driver
	}
	self.option_casp = caps
	self.option_prefix = prefixURL
	return nil
}
func (self *BaseBrowser) ReInit() {
	service, err := selenium.NewChromeDriverService(self.Path, port)
	if err != nil {
		log.Println("re connecting .... failed")
	}
	self.service = service
	driver, err := selenium.NewRemote(self.option_casp, self.option_prefix)
	if err != nil {
		log.Fatal("re connecting ... failed :", err)
	}
	self.driver = driver
}

func (self *BaseBrowser) Close() error {
	if err := self.driver.Quit(); err != nil {
		return err
	}
	if self.service != nil {
		return self.service.Stop()
	}
	return nil

}

func (self *BaseBrowser) CurrentDomain() string {
	u, _ := self.driver.CurrentURL()
	uu, _ := url.Parse(u)
	return uu.Host
}

func (self *BaseBrowser) SmartMultiFind(ids string, useGoQuery ...bool) (ele selenium.WebElement, err error) {
	if strings.Contains(ids, "|") {
		idFields := strings.Split(ids, "|")
		id := strings.TrimSpace(idFields[0])
		if eles, errs := self.SmartFindEles(id); err != nil {
			err = fmt.Errorf("%v:%s", errs, id)
			return
		} else {
			for _, ele := range self.SortEles(eles) {
				for _, id := range idFields[1:] {
					id = strings.TrimSpace(id)
					notfound := true
					if strings.HasPrefix(id, "/") {

						ele, err = ele.FindElement(selenium.ByXPATH, id)
						if err == nil {
							// L("Xpath Match", id[1:])
							err = fmt.Errorf("%v:%s", errs, id)
						}
						notfound = false
					} else if strings.HasPrefix(strings.TrimSpace(id), "*") {
						notfound = false
						ele, err = ele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), \"%s\")]", id[1:]))
						if err == nil {
							// L("Fuzy Match", id[1:])
							err = fmt.Errorf("%v:%s", errs, id)
						}
					} else if strings.HasPrefix(strings.TrimSpace(id), "*") {
						eles, errs := ele.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), \"%s\")]", id[1:]))
						if errs != nil {
							err = errs
							break
						}
						for _, e := range self.SortEles(eles) {
							if ok, errs := e.IsDisplayed(); errs == nil && ok {
								ele = e
								notfound = false
								break
							}
						}
						if notfound {
							err = fmt.Errorf("can not found! by %s", "default")
						} else {
							// L("Xpath Text fuzzy '", id)
						}
					} else if strings.HasPrefix(id, "'") && strings.HasSuffix(id, "'") {
						eles, errs := ele.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id))
						if errs != nil {
							err = errs
							break
						}
						for _, e := range self.SortEles(eles) {
							if ok, errs := e.IsDisplayed(); errs == nil && ok {
								ele = e
								notfound = false
								break
							}
						}
						if notfound {
							err = fmt.Errorf("can not found! by %s", "default")
						} else {
							// L("Xpath Text Match '", id)
						}
					} else if strings.HasPrefix(id, "\"") && strings.HasSuffix(id, "\"") {
						if eles, errs := ele.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id)); errs != nil {
							err = errs
							break
						} else {
							for _, e := range self.SortEles(eles) {
								log.Printf("  Find: %s \n", Green(self.GetEleInfo(e)))
								if ok, errs := e.IsDisplayed(); errs == nil && ok {
									ele = e
									notfound = false
									break
								}
							}
							if notfound {
								err = fmt.Errorf("can not found! by %s", "default")
							} else {
								// L("Xpath Text Match \" ", id)
							}
						}
					} else {
						if eles, errs := ele.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = '%s']", id)); errs != nil {
							err = errs
							break
						} else {
							for _, e := range self.SortEles(eles) {
								log.Printf("  Find: %s\n", Green(self.GetEleInfo(e)))
								if ok, errs := e.IsDisplayed(); errs == nil && ok {
									notfound = false
									ele = e
									break
								}
							}
							if notfound {
								err = fmt.Errorf("can not found! by %s", "default")
							} else {

								// L("Xpath Text Match", id)
							}
						}
					}

					if err != nil {
						// L("Css Text Match", id)
						ele, err = ele.FindElement(selenium.ByCSSSelector, id)
					}
				}
				if ele != nil {
					return ele, nil
				}
			}

		}

	} else {
		ele, err = self.SmartFindEle(ids)
	}
	return

}

// SmartFindEle, can use xpath, cssselector
func (self *BaseBrowser) SmartFindEle(id string, useGoQuery ...bool) (ele selenium.WebElement, err error) {
	notfound := true
	if strings.HasPrefix(id, "/") {
		ele, err = self.driver.FindElement(selenium.ByXPATH, id)
		if err == nil {
			// L("Xpath Match", id[1:])
		}
		notfound = false
	} else if strings.HasPrefix(strings.TrimSpace(id), "*") {
		notfound = false
		ele, err = self.driver.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), \"%s\")]", id[1:]))
		if err == nil {
			// L("Fuzy Match", id[1:])

		}
		// }
	} else if strings.HasPrefix(id, "'") && strings.HasSuffix(id, "'") {

		// text := strings.ReplaceAll(id, " ", "&nsp;")
		eles, errs := self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id))
		if errs != nil {
			err = errs
			return
		}
		for _, e := range self.SortEles(eles) {
			if ok, errs := e.IsDisplayed(); errs == nil && ok {
				ele = e
				notfound = false
				// L("Xpath Text Match '", id)

				break
			}
		}

		if notfound {
			err = fmt.Errorf("can not found! by %s", "default")
		}
	} else if strings.HasPrefix(id, "\"") && strings.HasSuffix(id, "\"") {

		// text := strings.ReplaceAll(id, " ", "&nsp;")
		if eles, errs := self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id)); errs != nil {
			err = errs
			return
		} else {
			for _, e := range self.SortEles(eles) {

				log.Printf("  Find: %s\n", Green(self.GetEleInfo(e)))
				if ok, errs := e.IsDisplayed(); errs == nil && ok {
					ele = e
					// L("Xpath Text Match \" ", id)

					notfound = false
					break
				}
			}

			if notfound {
				err = fmt.Errorf("can not found! by %s", "default")
			}
		}
	} else {
		// text := strings.ReplaceAll(id, " ", "&nsp;")
		if eles, errs := self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = '%s']", id)); errs != nil {
			err = errs
			return
		} else {
			for _, e := range self.SortEles(eles) {
				log.Printf("  Find: %s\n", Green(self.GetEleInfo(e)))

				if ok, errs := e.IsDisplayed(); errs == nil && ok {
					notfound = false
					ele = e
					// L("Xpath Text Match", id)

					break
				}
			}
			if notfound {
				err = fmt.Errorf("can not found! by %s", "default")
			}
		}
	}

	if err != nil {
		// L("Css Text Match", id)

		// L("warn try cssselector", err.Error())
		ele, err = self.driver.FindElement(selenium.ByCSSSelector, id)
	}
	return
}

func (self *BaseBrowser) SortEles(eles []selenium.WebElement) []selenium.WebElement {
	sort.Slice(eles, func(i, j int) bool {
		l, _ := eles[i].Location()
		r, _ := eles[j].Location()
		// if l.Y < r.Y {
		// 	log.Println("less:", self.GetEleInfo(eles[i]), self.GetEleInfo(eles[j]))
		// } else {
		// 	log.Println("more:", self.GetEleInfo(eles[i]), self.GetEleInfo(eles[j]))

		// }
		return l.Y < r.Y

	})
	return eles
}

// SmartFindEle, can use xpath, cssselector
func (self *BaseBrowser) SmartFindEleSource(id string) (html string, err error) {
	ele, eerr := self.SmartFindEle(id)
	if eerr != nil {
		err = eerr
		return
	}
	html, err = ele.GetAttribute("innerHTML")
	return
}

func (self *BaseBrowser) SmartFindByFromEle(ele selenium.WebElement, id string) (outele selenium.WebElement, err error) {
	id = strings.TrimSpace(id)
	outele = ele
	last := ele
	for _, ii := range strings.Split(id, "|") {
		pipe := strings.TrimSpace(ii)
		if strings.HasPrefix(pipe, "/") {
			outele, err = outele.FindElement(selenium.ByXPATH, pipe)
		} else if strings.HasPrefix(pipe, "'") && strings.HasSuffix(pipe, "'") {
			outele, err = outele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", pipe))
		} else if strings.HasPrefix(strings.TrimSpace(pipe), "*") {
			outele, err = outele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), \"%s\")]", pipe[1:]))
		} else if strings.HasPrefix(pipe, "\"") && strings.HasSuffix(pipe, "\"") {
			outele, err = outele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", pipe))
		} else {
			outele, err = outele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[text() = '%s']", pipe))
		}

		if err != nil || outele == nil {
			// log.Println("Warrning try css select", self.GetEleInfo(last), pipe)
			outele, err = last.FindElement(selenium.ByCSSSelector, pipe)
			if err != nil {
				log.Println("err:", err)
			}
		}
		if outele != nil {
			last = outele
		} else {
			log.Println("Warrning", self.GetEleInfo(last), pipe)
		}
	}
	return
}

func (self *BaseBrowser) SmartFindEles(id string, frameID ...string) (eles []selenium.WebElement, err error) {
	id = strings.TrimSpace(id)
	var pipe []string
	if frameID != nil {
		element, err := self.driver.FindElement(selenium.ByCSSSelector, "iframe"+frameID[0])
		if err != nil {
			log.Println("can not found iframe :", frameID[0])
		}
		//切换到iframe中
		err = self.driver.SwitchFrame(element)
		if err != nil {
			log.Println("switch frame:", err)
		}
	}
	if strings.Contains(id, "|") {
		for i, ii := range strings.Split(id, "|") {
			if i == 0 {
				id = strings.TrimSpace(ii)
			} else {
				pipe = append(pipe, strings.TrimSpace(ii))
			}
		}
	}
	if strings.HasPrefix(id, "/") {
		eles, err = self.driver.FindElements(selenium.ByXPATH, id)
	} else if strings.HasPrefix(id, "'") && strings.HasSuffix(id, "'") {
		eles, err = self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id))
	} else if strings.HasPrefix(strings.TrimSpace(id), "*") {
		eles, err = self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), \"%s\")]", id[1:]))
	} else if strings.HasPrefix(id, "\"") && strings.HasSuffix(id, "\"") {
		eles, err = self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id))
	} else {
		eles, err = self.driver.FindElements(selenium.ByXPATH, fmt.Sprintf("//*[text() = '%s']", id))
	}
	if err != nil || len(eles) == 0 {
		eles, err = self.driver.FindElements(selenium.ByCSSSelector, id)
	}

	for _, id := range pipe {
		id = strings.TrimSpace(id)
		tmpeles := []selenium.WebElement{}
		for _, ele := range self.SortEles(eles) {
			var tmpele selenium.WebElement
			if strings.HasPrefix(id, "/") {
				tmpele, err = ele.FindElement(selenium.ByXPATH, id)
			} else if strings.HasPrefix(id, "'") && strings.HasSuffix(id, "'") {
				tmpele, err = ele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id))
			} else if strings.HasPrefix(strings.TrimSpace(id), "*") {
				tmpele, err = ele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), \"%s\")]", id[1:]))
			} else if strings.HasPrefix(id, "\"") && strings.HasSuffix(id, "\"") {
				tmpele, err = ele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[text() = %s]", id))
			} else {
				tmpele, err = ele.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[text() = '%s']", id))
			}

			if err != nil || tmpele == nil {
				tmpele, err = ele.FindElement(selenium.ByCSSSelector, id)
			}
			if tmpele != nil {
				tmpeles = append(tmpeles, tmpele)
			}
		}
		if len(tmpeles) > 0 {
			eles = tmpeles
		}
	}
	eles = self.SortEles(eles)
	return
}

func L(action string, args ...interface{}) {
	// e := color.New(color.FgGreen, color.Bold).SprintFunc()
	pre := Green("[" + action + "]")
	// log.Println(args...)
	msg := ""
	for _, v := range args {
		switch v.(type) {
		case []string:
			msg += Yellow("list:\n") + "[\n\t" + strings.Join(v.([]string), ",\n\t") + "\n]"
		case Dict:
			// e, _ := json.Marshal(v.(Dict))
			msg += Yellow("\ndict:\n")
			for k, v := range v.(Dict) {
				msg += fmt.Sprintf("%s : %v\n", k, v)
			}
			// msg += string(e)
		default:
			msg += fmt.Sprint(v) + " "

		}

	}
	log.Printf("%s : %v\n", pre, msg)
}

func (self *BaseBrowser) ConsoleLog(result Result) {
	if result.Err != nil {
		if result.Err != nil && !strings.Contains(result.Err.Error(), "Timed out receiving message from renderer") {
			// log.Println(no, action, result.Err)
			log.Println(color.New(color.FgHiRed).SprintFunc()("[Err]"), ":", result.Err)
			self.switchMode()
		}

		// break
	} else {
		fmt.Println(Cyan("---------------", Bold(" Bool:", result.Bool, " Int:", result.Int, "---------------------------------------")))
		if result.Text != "" {
			fmt.Println(Green(result.Text))
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
		self.loopNow = 0

	}
	self.Mode = ac

}

func (self *BaseBrowser) Sleep() {
	time.Sleep(time.Duration(self.PageLoadTime) * time.Second)
}
