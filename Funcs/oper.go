package Funcs

import (
	"fmt"
	"log"
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
	defer func() {
		pre := ""
		for range self.OperStack {
			pre += "\t"
		}
		log.Println(Green(pre+"["+action+"]"), id, args)
	}()
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
			ele, res.Err = self.SmartFindEle("//input[@type=\"password\"]/../..//input[@type=\"text\"]")
			if res.Err != nil {
				return
			}
			res.Err = ele.SendKeys(name.(string))
			if res.Err != nil {
				return
			}
			L("User -- >", name.(string))
			if pwd, ok := kargs["password"]; ok {
				ele, res.Err = self.SmartFindEle("//input[@type=\"password\"]")
				if res.Err != nil {
					return
				}
				res.Err = ele.SendKeys(pwd.(string))

				L("Pwd -- >", pwd.(string))
				self.Sleep()
				if end, ok := kargs["end"]; ok {
					switch end.(string) {
					case "\n":
						ele.SendKeys(selenium.ReturnKey)
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
		sleep := 4
		if tk, ok := kargs["sleep"]; ok {
			sleep, res.Err = strconv.Atoi(tk.(string))
			if res.Err != nil {
				return
			}
		}
		timeout := time.Now().Add(time.Second * time.Duration(sleep))
		for {
			if id != "" {
				_, err := self.SmartFindEle(id)
				if err != nil {
					if time.Now().After(timeout) {
						res.Err = err
						break
					}
				} else {
					break
				}

			} else {
				if time.Now().After(timeout) {
					// res.Err = err
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
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
	args, _ := splitArgsTrim(line)
	argc := len(args)

	defer func() {

		if argc > 0 && args[0] == "end" {
			if stack := self.GetStack(); stack != nil {
				L("=============== End " + stack.Oper + " =================")

				if stack.End == 0 {
					self.PushStack(self.NextLine, "end")
					stack.End = self.NextLine
				}
				if stack.Oper == "for" {
					if tmp := self.RunThenBack(stack.Start); tmp.Bool {

						self.NextLine = stack.Start + 1

						L("For", tmp.Bool, "Go-To:", self.NextLine)
					} else {
						self.NextLine = stack.End + 1
						// L("EndFor", tmp.Bool, "Go-To:", self.NextLine)
						self.PopStack()
						// L("testFor", tmp.Bool, "Go-To:", self.NextLine)
					}
				} else if stack.Oper == "if" {
					self.NextLine++
					self.switchMode()
				} else if stack.Oper == "each" {
					if last := self.PopStack(); last != nil {
						L("EndEach", len(last.TmpStacks))
						result = self.RunEach(last.TmpID, last.TmpPage, last.TmpStacks)
					} else {
						L("err", "no stack! for each")
						self.switchMode(MODE_IF)
					}
					self.NextLine++
					self.switchMode()

				}
			} else {
				// log.Fatal("Ilegal End , end must match \"for\", \"each\". \"if\" ")
				self.switchMode()
				self.NextLine++
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
	case MODE_EACH:
		stack := self.GetStack()
		stack.TmpStacks = append(stack.TmpStacks, line)
		self.SetLastStack(stack)

		L("-- EachMode", line, len(stack.TmpStacks))
	default:
		switch args[0] {
		case "for":
			self.PushStack(self.NextLine, "for")
		case "if":
			if tmp := self.RunThenBack(self.NextLine); !tmp.Bool {
				L("Entry If")
				self.switchMode(MODE_IF)
				return
			}
		case "each":
			if tmp := self.RunThenBack(self.NextLine); tmp.Bool {
				L("Entry Each", args[1])
				self.switchMode(MODE_EACH)
				self.PushStack(self.NextLine, "each")
				last := self.GetStack()
				last.TmpID = args[1]
				last.TmpPage, result.Err = self.driver.PageSource()
				if len(last.TmpPage) == 0 {
					L("Err", "empty")
				}
				self.SetLastStack(last)
				return
			} else {
				L("Not found Each")
				self.ConsoleLog(tmp)
			}
		}
		result = self.StepRun()
	}
	return
}

func (self *BaseBrowser) StepRun() (result Result) {
	args, kargs := splitArgsTrim(self.lines[self.NextLine])
	argc := len(args)
	if argc == 0 {
		L("Err", "err for empty")
		return
	}
	main := args[0]
	if argc > 2 {
		result = self.Action(args[1], main, kargs, args[2:]...)
	} else if argc > 1 {
		result = self.Action(args[1], main, kargs)
	} else {
		result = self.Action("", main, kargs)
	}
	if result.Err != nil && strings.Contains(result.Err.Error(), "Timed out receiving message from renderer") {
		result.Err = nil
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
		if last.Err != nil {
			L("Exit", last.Err)
			break
		}

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
