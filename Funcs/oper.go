package Funcs

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
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

func (self *BaseBrowser) ScrollTo(id string, args []string, kargs Dict) (res Result) {
	if id != "" {
		var ele selenium.WebElement
		ele, res.Err = self.SmartMultiFind(id)
		if res.Err != nil {
			L("scroll err", res.Err)
			return
		} else {
			// L("Found :", ele)
		}
		p, err := ele.Location()
		if err != nil {
			res.Err = err
		}
		L("move to :", p.X, p.Y)
		self.driver.ExecuteScript(fmt.Sprintf("window.scrollTo(0, %d);", p.Y), nil)
		self.Sleep()
	} else {
		if f, ok := kargs["offset"]; ok {
			moveOffset, _ := strconv.Atoi(f.(string))
			self.driver.ExecuteScript(fmt.Sprintf("window.scrollTo(0, %d);", moveOffset), nil)
			self.Sleep()

		} else {
			self.driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil)
		}
	}
	if t, ok := kargs["timeout"]; ok {
		if w, err := strconv.Atoi(t.(string)); err == nil {
			time.Sleep(time.Duration(w) * time.Second)
		}
	}
	return
}

func (self *BaseBrowser) ClickToIdFromEle(i int, id string, args []string, kargs Dict) (res Result) {
	if len(args) == 2 {

		// before, _ := self.driver.PageSource()
		var ele selenium.WebElement
		var eles []selenium.WebElement
		if _, ok := kargs["global"]; ok {
			var ele selenium.WebElement
			ele, res.Err = self.SmartMultiFind(args[1])
			if ele != nil {
				if ok, _ := ele.IsEnabled(); ok {
					before, _ := self.driver.PageSource()

					if res.Err = ele.Click(); res.Err != nil {

						// L("click error")
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
			}
		} else {
			eles, res.Err = self.SmartFindEles(id)
			if len(eles) <= i {
				log.Println(Red("ranger err :", i))
				return
			}
			if i >= len(eles) {
				// res.Err = fmt.Errorf("Err: %s", "beyond ele")
				return
			}
			ele = eles[i]
			_, res.Err = ele.LocationInView()

			if res.Err != nil {
				return
			}
			L("Click Then by", args[1])
			ele, res.Err = self.SmartFindByFromEle(ele, args[1])
			if ele != nil {
				res.Err = ele.Click()
				// L("Click ok")
				self.Sleep()
				if res.Err != nil {
					return
				}
			}
			// time.Sleep(4 * time.Second)

		}

	} else {
		before, _ := self.driver.PageSource()
		var ele selenium.WebElement
		var eles []selenium.WebElement
		eles, res.Err = self.SmartFindEles(id)
		if len(eles) <= i {
			log.Println(Red("ranger err :", i))
			return
		}
		ele = eles[i]

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
	}
	return
}

func (self *BaseBrowser) RunNextLine() (result Result) {
	line := self.lines[self.NextLine]
	args, kargs := splitArgsTrim(line)
	argc := len(args)

	defer func() {

		if argc > 0 && args[0] == "end" {
			if stack := self.GetStack(); stack != nil {
				// L("=============== End " + stack.Oper + " =================")

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
						result = self.RunEach(last.TmpID, last.TmpPage, last.TmpStacks, last.TmpKargs)
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

				self.switchMode(MODE_EACH)
				self.PushStack(self.NextLine, "each")
				last := self.GetStack()
				last.TmpID = args[1]
				last.TmpPage, result.Err = self.driver.PageSource()
				last.TmpKargs = make(Dict)
				for k, v := range kargs {
					last.TmpKargs[k] = v
				}
				if len(last.TmpPage) == 0 {
					L("Err", "empty")
				}
				self.SetLastStack(last)
				L("Entry Each", args[1], last.TmpKargs)
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
			// L("Exit", last.Err)
			if strings.Contains(last.Err.Error(), "dial tcp 127.0.0.1:9515: connect: connection refused") {
				self.ReInit()
			}
			break
		}

	}
}
