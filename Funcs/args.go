package Funcs

import (
	"encoding/json"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/tebeka/selenium"
	"golang.org/x/net/html"
)

var (
	removeJSRE = regexp.MustCompile(`<script>.+?</script>`)
)

type Dict map[string]interface{}

func splitArgsTrim(raw string) (as []string, kargs Dict) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	if strings.HasPrefix(raw, ":") {

		fs := strings.SplitN(raw, ":", 2)
		as = append(as, strings.TrimSpace(fs[0]))
		args, wkargs := utils.DecodeToOptions(fs[1])
		as = append(as, args...)
		for k, v := range wkargs {
			kargs[k] = v
		}
		// kargs = wkargs
	} else {
		as = append(as, strings.TrimSpace(raw))
	}
	return
	// kas := []string{}
	// // L("argc :", len(as))
	// if len(as) > 1 {
	// 	if strings.Contains(as[1], ",") {
	// 		argsStr := as[1]
	// 		as = []string{as[0]}
	// 		for _, w2 := range splitargs(argsStr) {
	// 			if isKargs(w2) {
	// 				kas = append(kas, parseArg(w2))
	// 			} else {
	// 				as = append(as, parseArg(w2))

	// 			}

	// 		}
	// 	} else {
	// 		needremove := []string{}
	// 		for i := range as {
	// 			w2 := as[i]
	// 			if isKargs(w2) {
	// 				kas = append(kas, parseArg(w2))
	// 				needremove = append(needremove, w2)
	// 			}
	// 		}
	// 		for _, is := range needremove {
	// 			as = remove(as, is)
	// 		}
	// 	}

	// }
	// // else {
	// // 	if isKargs(as[1]) {
	// // 		as = []string{as[0]}
	// // 		kas = append(kas, parseArg(as[1]))
	// // 	}
	// // }
	// kargs = parseKargs(kas...)
	// return
}

func isKargs(raw string) (ok bool) {
	w2 := strings.TrimSpace(raw)

	if strings.Contains(w2, "=") {

		if !strings.HasPrefix(w2, "\"") && !strings.HasPrefix(w2, "'") {
			// L("l1", w2, raw)
			if strings.HasSuffix(w2, "\"") || strings.HasSuffix(w2, "\"") {
				// L("l2", w2, raw)

				if strings.Count(w2, "'")%2 == 0 || strings.Count(w2, "\"")%2 == 0 {

					// L("isKargs", w2)
					ok = true
				}
			} else if strings.Count(w2, "[") == 1 || strings.Count(w2, "]") == 1 {

				if strings.Count(w2, "'")%2 == 0 || strings.Count(w2, "\"")%2 == 0 {

					// L("isKargs", w2)
					ok = true
				}
				// L("isKargs", w2)
				// ok = true
			} else {
				if _, err := strconv.Atoi(strings.TrimSpace(strings.SplitN(w2, "=", 2)[1])); err == nil {
					ok = true
				}

			}
		}
	}
	if (strings.HasPrefix(w2, "'") && strings.HasSuffix(w2, "'")) || (strings.HasPrefix(w2, "\"") && strings.HasSuffix(w2, "\"")) {
		return
	}

	return
}
func parseArg(arg string) string {
	p := strings.TrimSpace(arg)
	if (strings.HasPrefix(p, "'") && strings.HasSuffix(p, "'")) || (strings.HasPrefix(p, "\"") && strings.HasSuffix(p, "\"")) {
		return p[1 : len(p)-1]
	} else {
		// } else if p == "true" {
		// 	return true
		// } else if p == "true" {
		// 	return false
		// } else {
		return p
	}
}

func parseArgs(arg string) (args []string) {
	ps := strings.TrimSpace(arg)
	for _, p := range splitargs(ps) {
		args = append(args, parseArg(p))
	}
	return
}

func parseKargs(args ...string) (w map[string]interface{}) {
	w = make(map[string]interface{})
	for _, raw := range args {
		if strings.Contains(raw, "=") {
			// ok = true
			fs := strings.SplitN(raw, "=", 2)
			value := strings.TrimSpace(fs[1])
			key := strings.TrimSpace(fs[0])
			if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				value := parseArgs(value[1 : len(value)-1])
				w[key] = value
			} else {
				value := parseArg(fs[1])
				w[key] = value
			}

		}
	}
	return
}

func splitargs(raw string) []string {
	quoted := false
	last := ' '
	// mquote := false
	a := strings.FieldsFunc(raw, func(r rune) (e bool) {
		if r == '"' {
			if !quoted {
				last = r
			}
			quoted = !quoted
		}

		if r == '[' && !quoted {
			quoted = !quoted
		} else if r == ']' && quoted {
			quoted = !quoted
		}

		if r == '\'' {
			if !quoted {
				last = r
			}
			if last == r {
				quoted = !quoted

			}
		}
		if last == ' ' {
			last = r
		}
		e = !quoted && r == ','
		if e {
			last = ' '
		}
		return
	})
	return a
}

func parseToJsonOrStruct(raw string, obj ...interface{}) (datas Dict, err error) {
	rawU, err := url.QueryUnescape(raw)
	if err != nil {
		return
	}
	datas = make(Dict)
	for _, field := range strings.Split(rawU, ";") {
		if strings.Contains(field, "=") {
			fs := strings.SplitN(field, "=", 2)
			datas[strings.TrimSpace(fs[0])] = strings.TrimSpace(fs[1])
		} else {
			L("ignore cookie:", field)
		}
	}
	buf, err := json.Marshal(&datas)
	if err != nil {
		return
	}
	if obj != nil {
		err = json.Unmarshal(buf, obj[0])
	}
	return
}

func parseCookie(raw string, url string) (cs []*selenium.Cookie, err error) {
	datas, err := parseToJsonOrStruct(raw)
	if err != nil {
		return
	}
	for k, v := range datas {
		c := &selenium.Cookie{
			Name:   k,
			Value:  v.(string),
			Expiry: math.MaxUint32,
			Domain: url,
			Path:   "/",
		}
		cs = append(cs, c)
	}
	return
}

func remove(slice []string, one string) []string {
	s := -1
	for i, v := range slice {
		if v == one {
			s = i
			break
		}
	}
	if s > 0 {
		return append(slice[:s], slice[s+1:]...)
	} else {
		return slice
	}
}

func removeScript(n *html.Node) {
	// if note is script tag
	if n.Type == html.ElementNode && n.Data == "script" {
		n.Parent.RemoveChild(n)
		return // script tag is gone...
	}
	if n.Type == html.ElementNode && n.Data == "style" {
		n.Parent.RemoveChild(n)
		return // script tag is gone...
	}
	// traverse DOM
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeScript(c)
	}
}

func removeScriptAndCss(raw string) string {

	// func main() {
	for {

		si := strings.Index(raw, "<style")
		ei := strings.Index(raw, "</style>")
		if si > -1 && ei > -1 {
			raw = raw[:si] + raw[ei:]
		}
		si = strings.Index(raw, "<script")
		ei = strings.Index(raw, "</script>")
		// fmt.Println(si, ei, "OK")
		if si > -1 && ei > -1 {
			raw = raw[:si] + raw[ei:]
		} else {
			break
		}

	}

	return raw
	// }

}
