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
		if res.Err == nil {
			res.Bool = true
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
		if res.Err == nil {
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

func (self *BaseBrowser) RunThenBack(no int) (result Result) {
	tmpNextLine := self.NextLine
	defer func() {
		self.NextLine = tmpNextLine
	}()
	self.NextLine = no
	result = self.StepRun()
	return
}

func (self *BaseBrowser) RunNextLine() (result Result) {
	line := self.lines[self.NextLine]
	args := splitArgsTrim(line)
	argc := len(args)

	defer func() {
		if args[0] == "end" {
			stack := self.GetStack()
			if stack != nil {
				if stack.End == 0 {
					self.PushStack(self.NextLine, "end")
					stack.End = self.NextLine
				}
				if stack.Oper == "for" {
					if tmp := self.RunThenBack(stack.Start); tmp.Bool {
						self.NextLine = stack.Start + 1
					} else {
						self.NextLine = stack.End + 1
					}
				} else if stack.Oper == "if" {
					self.NextLine++
					self.switchMode()
				}
			} else {
				log.Fatal("Ilegal End , end must match \"for\", \"each\". \"if\" ")
			}
		} else {
			self.NextLine++
		}
	}()

	if argc == 0 {
		return
	}

	if strings.HasPrefix(args[0], "#") {
		L("comment", strings.Join(args, " "))
		return
	}
	switch self.Mode {
	case MODE_IF:
	default:
		switch args[0] {
		case "for":
			self.PushStack(self.NextLine, "for")
		case "if":
			self.PushStack(self.NextLine, "if")
			if tmp := self.RunThenBack(self.NextLine); !tmp.Bool {
				self.switchMode(MODE_IF)
			}
		}
		result = self.StepRun()
	}
	return
}

func (self *BaseBrowser) StepRun() (result Result) {
	args := splitArgsTrim(self.lines[self.NextLine])
	argc := len(args)
	if argc == 0 {
		L("Err", "err for empty")
		return
	}
	main := args[0]
	if argc > 2 {
		result = self.Action(args[1], main, args[2:]...)
	} else if argc > 1 {
		result = self.Action(args[1], main)
	} else {
		result = self.Action("", main)
	}
	return
}

func (self *BaseBrowser) Parse(actions string) {
	self.lines = strings.Split(actions, "\n")
	linenum := len(self.lines)
	var last Result
	for {
		if self.NextLine >= linenum {
			break
		}
		last = self.RunNextLine()
		self.ConsoleLog(last)
	}
}

// func (self *BaseBrowser) Parse(actions string) {
// 	self.lines = strings.Split(actions, "\n")
// 	var last Result
// 	ifjump := false
// 	// Mode := MODE_FLOW

// 	// forstack := []string{}
// 	// forconditionargs := []string{}
// 	// forloop := false
// 	for no, action := range self.lines {

// 		args := splitArgsTrim(action)
// 		if self.PreTest(args) {
// 			continue
// 		}

// 		switch self.Mode {
// 		case MODE_FOR_RUN:
// 			self.RunForStack()
// 			self.switchMode()
// 		case MODE_EACH_RUN:
// 			self.ActionEle(self.TmpID, []byte(self.TmpPage), self.forstack)
// 			self.switchMode()

// 		case MODE_FLOW:
// 			switch args[0] {
// 			case "for":
// 				self.AddOper(no, "for")
// 				self.switchMode(MODE_FOR)
// 				self.forLoopTestArgs = splitArgsTrim(action)
// 				log.Println("[FOR]:", strings.Join(self.forLoopTestArgs, " - "))
// 				// continue
// 			case "each":

// 				self.AddOper(no, "each")
// 				self.switchMode(MODE_EACH)
// 				// self.forLoopTestArgs = splitArgsTrim(action)
// 				log.Println("[EACH]:", strings.Join(self.forLoopTestArgs, " - "))
// 				var err error

// 				self.TmpPage, err = self.driver.PageSource()
// 				self.TmpID = args[1]
// 				if err != nil {
// 					log.Fatal(Red("[ERR]"), ":", Bold(err.Error()))
// 					break
// 				}
// 			default:
// 				ifjump, last = self.oneLine(no, ifjump, args)
// 				self.ConsoleLog(last)
// 			}
// 		}

// 	}
// }

// func (self *BaseBrowser) runLine(no int) {
// 	args := splitArgsTrim(self.lines[no])
// 	if len(args) == 0 {
// 		log.Fatal("no line found!!!")
// 		return
// 	}
// 	switch args[0] {
// 	case "for":
// 		self.AddOper(no, "for")
// 		self.switchMode(MODE_FOR)
// 		// self.forLoopTestArgs = splitArgsTrim(action)
// 		log.Println("[FOR]:", strings.Join(self.forLoopTestArgs, " - "))
// 		// continue
// 	case "each":

// 		self.AddOper(no, "each")
// 		self.switchMode(MODE_EACH)
// 		// self.forLoopTestArgs = splitArgsTrim(action)
// 		log.Println("[EACH]:", strings.Join(self.forLoopTestArgs, " - "))
// 		var err error

// 		self.TmpPage, err = self.driver.PageSource()
// 		self.TmpID = args[1]
// 		if err != nil {
// 			log.Fatal(Red("[ERR]"), ":", Bold(err.Error()))
// 			break
// 		}
// 	default:
// 		self.ifjump, last = self.oneLine(no, self.ifjump, args)
// 		self.ConsoleLog(last)
// 	}
// }
