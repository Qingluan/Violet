package Funcs

import (
	"encoding/json"
	"math"
	"net/url"
	"strings"

	"github.com/tebeka/selenium"
)

type Dict map[string]interface{}

func splitArgsTrim(raw string) (as []string, kargs Dict) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	for _, w := range strings.SplitN(raw, ":", 2) {
		as = append(as, parseArg(w))
	}

	kas := []string{}
	// L("argc :", len(as))
	if len(as) > 1 {
		if strings.Contains(as[1], ",") {
			argsStr := as[1]
			as = []string{as[0]}
			for _, w2 := range splitargs(argsStr) {
				if isKargs(w2) {
					kas = append(kas, parseArg(w2))
				} else {
					as = append(as, parseArg(w2))

				}

			}
		} else {
			needremove := []string{}
			for i := range as {
				w2 := as[i]
				if isKargs(w2) {
					kas = append(kas, parseArg(w2))
					needremove = append(needremove, w2)
				}
			}
			for _, is := range needremove {
				as = remove(as, is)
			}
		}

	}
	// else {
	// 	if isKargs(as[1]) {
	// 		as = []string{as[0]}
	// 		kas = append(kas, parseArg(as[1]))
	// 	}
	// }
	kargs = parseKargs(kas...)
	return
}

func isKargs(raw string) (ok bool) {
	w2 := strings.TrimSpace(raw)

	if strings.Contains(w2, "=") {

		if !strings.HasPrefix(w2, "\"") && !strings.HasPrefix(w2, "'") {
			// L("l1", w2)
			if strings.HasSuffix(w2, "\"") || strings.HasSuffix(w2, "\"") {
				// L("l2", w2)

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
		fs := strings.SplitN(field, "=", 2)
		datas[strings.TrimSpace(fs[0])] = strings.TrimSpace(fs[1])
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
