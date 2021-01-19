package Funcs

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tebeka/selenium"
)

func (self *BaseBrowser) Action(id string, action string, args ...string) (res Result) {
	res = Result{
		Action: action,
	}
	if action == "" {
		res.Action = id
	}
	e := color.New(color.FgGreen, color.Bold).SprintFunc()

	defer log.Println(e("["+action+"]"), id, args)
	defer time.Sleep(2 * time.Second)
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
		} else if strings.TrimSpace(id) != "" {
			_, res.Err = self.driver.ExecuteScript(strings.TrimSpace(id), nil)
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
		idint, iterr := strconv.Atoi(id)
		if iterr == nil {
			// res.Int = idint
			self.loopCount = idint
			return
		}
		var ele selenium.WebElement
		// var err error
		ele, res.Err = self.SmartFindEle(id)
		if res.Err != nil {
			res.Bool = true
			return
		}
		if args != nil {
			res.Text, res.Err = ele.GetAttribute(strings.TrimSpace(args[0]))
		} else {
			res.Text, res.Err = ele.Text()
		}
		if res.Err != nil {
			res.Bool = true
			// return
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
	self.lines = strings.Split(actions, "\n")
	var last Result
	ifjump := false
	// Mode := MODE_FLOW

	// forstack := []string{}
	// forconditionargs := []string{}
	// forloop := false
	for no, action := range self.lines {

		args := splitArgsTrim(action)
		if self.PreTest(args) {
			continue
		}

		switch self.Mode {
		case MODE_FOR_RUN:
			self.RunForStack()
			self.switchMode()
		case MODE_EACH_RUN:
			self.ActionEle(self.TmpID, []byte(self.TmpPage), self.forstack)
			self.switchMode()

		case MODE_FLOW:
			switch args[0] {
			case "for":
				self.AddOper(no, "for")
				self.switchMode(MODE_FOR)
				self.forLoopTestArgs = splitArgsTrim(action)
				log.Println("[FOR]:", strings.Join(self.forLoopTestArgs, " - "))
				// continue
			case "each":

				self.AddOper(no, "each")
				self.switchMode(MODE_EACH)
				// self.forLoopTestArgs = splitArgsTrim(action)
				log.Println("[EACH]:", strings.Join(self.forLoopTestArgs, " - "))
				var err error

				self.TmpPage, err = self.driver.PageSource()
				self.TmpID = args[1]
				if err != nil {
					log.Fatal(Red("[ERR]"), ":", Bold(err.Error()))
					break
				}
			default:
				ifjump, last = self.oneLine(no, ifjump, args)
				self.ConsoleLog(last)
			}
		}

	}
}

func (self *BaseBrowser) AddOper(no int, oper string) {
	self.OperStack = append(self.OperStack, oper)
	self.OperNow = oper
	self.OperStackWhere = append(self.OperStackWhere, []int{no})
}

func (self *BaseBrowser) PreTest(no int, args []string) bool {
	if strings.HasPrefix(args[0], "#") {
		log.Println("[comment]:", strings.Join(args, " "))
		return true
	}
	if len(args) == 0 {
		return true
	}
	if self.OperNow != "" {
		if args[0] == "end" {
			lastStackWhere := self.OperStackWhere[len(self.OperStackWhere)-1]
			lastStackWhere = append(lastStackWhere, no)
			self.RunStack()
			self.OperStack = self.OperStack[:len(self.OperStack)-1]
		} else {
			self.forstack = append(self.forstack, args)
		}
	}

	// switch self.Mode {
	// case MODE_FOR:
	// 	if args[0] == "endfor" {
	// 		self.switchMode(MODE_FOR_RUN)
	// 	} else {
	// 		self.forstack = append(self.forstack, args)
	// 	}
	// 	return true
	// case MODE_EACH:
	// 	if args[0] == "endeach" {
	// 		self.switchMode(MODE_EACH_RUN)
	// 	} else {
	// 		self.forstack = append(self.forstack, args)
	// 	}
	// 	return true
	// }
	return false
}

func (self *BaseBrowser) RunStack() {
	defer func() {
		if len(self.OperStack) != 0 {
			self.OperNow = self.OperStack[len(self.OperStack)-1]
		} else {
			self.OperNow = ""
		}
		self.OperStackWhere = self.OperStackWhere[:len(self.OperStackWhere)-1]
	}()
	wheres := self.OperStackWhere[len(self.OperStackWhere)-1]
	if len(wheres) != 2 {
		log.Fatal("error must where is 2", wheres)
	}
	for i := wheres[0]; i < wheres[1]; i++ {
		// args := splitArgsTrim(self.lines[i])
		self.runLine(i)
	}
}

func (self *BaseBrowser) runLine(no int) {
	args := splitArgsTrim(self.lines[no])
	if len(args) == 0 {
		log.Fatal("no line found!!!")
		return
	}
	switch args[0] {
	case "for":
		self.AddOper(no, "for")
		self.switchMode(MODE_FOR)
		// self.forLoopTestArgs = splitArgsTrim(action)
		log.Println("[FOR]:", strings.Join(self.forLoopTestArgs, " - "))
		// continue
	case "each":

		self.AddOper(no, "each")
		self.switchMode(MODE_EACH)
		// self.forLoopTestArgs = splitArgsTrim(action)
		log.Println("[EACH]:", strings.Join(self.forLoopTestArgs, " - "))
		var err error

		self.TmpPage, err = self.driver.PageSource()
		self.TmpID = args[1]
		if err != nil {
			log.Fatal(Red("[ERR]"), ":", Bold(err.Error()))
			break
		}
	default:
		self.ifjump, last = self.oneLine(no, self.ifjump, args)
		self.ConsoleLog(last)
	}
}
