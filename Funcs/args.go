package Funcs

import "strings"

type Dict map[string]interface{}

func splitArgsTrim(raw string) (as []string, kargs Dict) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	for _, w := range strings.SplitN(raw, ":", 2) {
		as = append(as, parseArg(w))
	}

	kas := []string{}
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

					L("isKargs", w2)
					ok = true
				}
			} else if strings.Count(w2, "[") == 1 || strings.Count(w2, "]") == 1 {

				if strings.Count(w2, "'")%2 == 0 || strings.Count(w2, "\"")%2 == 0 {

					L("isKargs", w2)
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
