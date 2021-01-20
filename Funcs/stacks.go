package Funcs

type Stack struct {
	Start int
	End   int
	Oper  string
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
	} else {
		return nil
	}
	return
}
