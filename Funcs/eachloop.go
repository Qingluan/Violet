package Funcs

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (self *BaseBrowser) EleToJsonString(s *goquery.Selection, key ...string) string {
	ss := make(map[string]string)

	if key == nil {
		ss["html"], _ = s.Html()
		ss["text"] = s.Text()
	} else {
		if key[0] == "_" || key[0] == "*" {
			if len(s.Nodes) > 0 {
				node := s.Nodes[0]
				for _, attr := range node.Attr {
					ss[attr.Key] = ss[attr.Val]
				}
			}

		} else {
			for _, k := range key {
				if k == "text" {
					ss["text"] = s.Text()
				} else {
					ss[k] = s.AttrOr(k, "")
				}
			}

		}
	}

	res, _ := json.Marshal(&ss)
	return string(res)
}

func WithEle(id string, args []string, kargs Dict, doc *goquery.Document, s *goquery.Selection, do func(sb *goquery.Selection, args []string), multiDo func(sss map[string][]string)) {
	if ifcondition, ok := kargs["contains"]; ok {
		if !strings.Contains(s.Text(), ifcondition.(string)) {
			L("test Failed so jump")
			return
		}
	}
	if _, ok := kargs["global"]; ok {
		doc.Find(id).Each(func(i int, sb *goquery.Selection) {
			if attrs, ok := kargs["attrs"]; ok {
				switch attrs.(type) {
				case []string:
					do(sb, attrs.([]string))
				}
			} else {

				do(sb, []string{})
				// _, res.Err = fp.WriteString(self.EleToJsonString(sb) + "\n")
			}
		})
		return
	}

	if key, ok := kargs["find"]; ok {
		s.Find(key.(string)).Each(func(i int, sb *goquery.Selection) {
			if attrs, ok := kargs["attrs"]; ok {
				switch attrs.(type) {
				case []string:
					do(sb, attrs.([]string))
				}
			} else {

				do(sb, []string{})
				// _, res.Err = fp.WriteString(self.EleToJsonString(sb) + "\n")
			}
		})
	} else if multi, ok := kargs["multi"]; ok {
		sss := make(map[string][]string)
		sss["multi"] = []string{}

		switch multi.(type) {
		case []string:
			log.Println(Green("Muti:"), multi.([]string))
			for _, subid := range multi.([]string) {
				if strings.HasPrefix(subid, "^") {
					doc.Find(subid[1:]).EachWithBreak(func(i int, sb *goquery.Selection) bool {
						log.Println(Cyan("found:", subid))
						ss := make(map[string]string)
						ss["text"] = sb.Text()
						// ss["html"], _ = sb.Html()
						// fp.WriteString(self.EleToJson(subele) + "\n")
						ssRes, _ := json.Marshal(ss)
						sss["multi"] = append(sss["multi"], string(ssRes))

						return false
					})
				} else {
					s.Find(subid).EachWithBreak(func(i int, sb *goquery.Selection) bool {
						log.Println(Cyan("found:", subid))
						ss := make(map[string]string)
						ss["text"] = sb.Text()
						// ss["html"], _ = sb.Html()
						// fp.WriteString(self.EleToJson(subele) + "\n")
						ssRes, _ := json.Marshal(ss)
						sss["multi"] = append(sss["multi"], string(ssRes))

						return false
					})
				}

			}
		case string:
			s.Find(multi.(string)).Each(func(i int, sb *goquery.Selection) {
				ss := make(map[string]string)
				ss["text"] = sb.Text()
				// ss["html"], _ = sb.Html()
				// fp.WriteString(self.EleToJson(subele) + "\n")
				ssRes, _ := json.Marshal(ss)
				sss["multi"] = append(sss["multi"], string(ssRes))

			})
		}

		multiDo(sss)

	} else {
		argc := len(args)
		if argc > 2 {
			do(s, args[2:])
			// _, res.Err = fp.WriteString(self.EleToJsonString(s, args[2:]...) + "\n")
		} else {
			do(s, []string{})
			// _, res.Err = fp.WriteString(self.EleToJsonString(s) + "\n")

		}
	}
}
func (self *BaseBrowser) Page() string {
	p, _ := self.driver.PageSource()
	return p
}

func (self *BaseBrowser) RunEach(id string, page string, stacks []string, kargs Dict) (res Result) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(page)))
	if err != nil {
		res.Err = err
		return
	}
	counter := func(page, id string) int {
		newdoc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(self.Page())))
		e := 0
		newdoc.Find(id).Each(func(i int, _ *goquery.Selection) {
			e++
		})
		return e
	}

	docSize := counter(page, id)
	L("kargs:", kargs)
	// Mainargs, Mainkargs := splitArgsTrim(id)
	baseLoopNum := 0
	if mode, ok := kargs["auto"]; ok {
		log.Println("each mode", Yellow(mode), "size:", Yellow(docSize))
		thisFinishedNum, _ := self.EachOneBatch(id, doc, baseLoopNum, stacks)
		baseLoopNum += thisFinishedNum
		switch mode {
		case "scroll":
			batchnum := 0
			for {
				batchnum++
				log.Println(Cyan("batch:", batchnum))
				self.OperScrollTo("", []string{}, kargs)
				finishNum := 0
				newdoc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(self.Page())))
				x := counter(self.Page(), id)

				if x == docSize {
					break
				} else {
					docSize = x
				}
				finishNum, res = self.EachOneBatch(id, newdoc, baseLoopNum, stacks)
				baseLoopNum += finishNum
			}
		case "next":

		}
	} else {
		_, res = self.EachOneBatch(id, doc, baseLoopNum, stacks)
		return
	}

	return
}

func (self *BaseBrowser) EachOneBatch(id string, doc *goquery.Document, baseLoopNum int, stacks []string) (loopNum int, res Result) {
	doc.Find(id).Each(func(i int, s *goquery.Selection) {
		if i < baseLoopNum {
			return
		}
		loopNum++
		// L("Found!", s.Nodes[0].Namespace)
		defer func() {
			if res.Err != nil {
				L("Each Err", res.Err.Error())
			}
		}()

		for _, line := range stacks {

			args, kargs := splitArgsTrim(line)
			// argc := len(args)

			// id := args[1]
			main := args[0]
			// ss := args[1:].(interface{})
			if main == "end" {
				continue
			}

			// if argc < 1 {
			// 	continue

			// }
			if len(args) > 1 {

				L("For-Each", main, args[1])
			} else {
				L("For-Each", main)
			}
			// kargs := parseKargs(args...)
			switch main {
			case "collect":
				self.OperCollect(id, doc, s, args, kargs)
			case "wait":
				res = self.OperWait(id, args, kargs)
			case "back":
				res.Err = self.driver.Back()
				self.Sleep()
			case "click":

				res = self.OperClickToIdFromEle(i, id, args, kargs)

			case "input":
				res = self.OperInputFromEle(i, id, args, kargs)
				self.Sleep()
			case "print":
				WithEle(id, args, kargs, doc, s, func(sb *goquery.Selection, args []string) {
					L("show", self.EleToJsonString(s, args...))
				}, func(sss map[string][]string) {
					f, _ := json.MarshalIndent(sss, "", "  ")
					L("show:", Yellow(string(f)))

				})
			}

			// case "find":
			// 	s.Find(args[1]).Each(func(i int, sb *goquery.Selection) {
			// 		L("find", args[1], sb.Text())
			// 	})
			// }
		}
	})
	return
}
