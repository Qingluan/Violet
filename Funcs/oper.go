package Funcs

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
)

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

func (self *BaseBrowser) Clear() {
	self.lines = []string{}
	// linenum := 0
	// var last Result
	self.NextLine = 0
}

func (self *BaseBrowser) Parse(actions string) {
	self.lines = strings.Split(actions, "\n")
	linenum := len(self.lines)
	var last Result

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ifbreak := false
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			if sig.String() == os.Interrupt.String() {
				ifbreak = true
			}

		}
	}()
	for {
		if ifbreak {
			fmt.Println(Yellow("Ctrl-c"))
			break
		}
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
