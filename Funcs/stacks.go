package Funcs

import "reflect"

type Stack struct {
	Start     int
	End       int
	Oper      string
	NowLoop   int
	TmpID     string
	TmpPage   string
	TmpStacks []string
	TmpKargs  Dict
}

func (self *BaseBrowser) PushStack(no int, oper string) {
	if oper == "end" {
		last := self.OperStack[len(self.OperStack)-1]
		last.End = no
		self.OperStack[len(self.OperStack)-1] = last
	} else {
		newstack := Stack{
			Start: no,
			Oper:  oper,
		}
		self.OperStack = append(self.OperStack, newstack)
	}
}

func (self *BaseBrowser) GetStack() (s *Stack) {
	if len(self.OperStack) != 0 {
		last := self.OperStack[len(self.OperStack)-1]
		s = new(Stack)
		s.Start = last.Start
		s.End = last.End
		s.Oper = last.Oper
		s.NowLoop = last.NowLoop
		s.TmpStacks = last.TmpStacks
		s.TmpID = last.TmpID
		s.TmpPage = last.TmpPage
		s.TmpKargs = last.TmpKargs
		// CopyStruct(s, &last)
		// copy(s, last)
	} else {
		return nil
	}
	return
}

func (self *BaseBrowser) SetLastStack(stack *Stack) {
	self.OperStack[len(self.OperStack)-1] = *stack
}

// func (self *)

func (self *BaseBrowser) PopStack() *Stack {
	if len(self.OperStack) != 0 {
		var last Stack
		last, self.OperStack = self.OperStack[len(self.OperStack)-1], self.OperStack[:len(self.OperStack)-1]
		return &last
	} else {
		return nil
	}
}

// func (self *)
func CopyStruct(src, dst interface{}) {
	sval := reflect.ValueOf(src).Elem()
	dval := reflect.ValueOf(dst).Elem()

	for i := 0; i < sval.NumField(); i++ {
		value := sval.Field(i)
		name := sval.Type().Field(i).Name

		dvalue := dval.FieldByName(name)
		if dvalue.IsValid() == false {
			continue
		}
		dvalue.Set(value) //这里默认共同成员的类型一样，否则这个地方可能导致 panic，需要简单修改一下。
	}
}
