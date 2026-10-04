package ignore

import (
	"os"
	"regexp"
	"strings"
	"bufio"
)

type Pattern struct {
	re		*regexp.Regexp
	negate 		bool
	dirOnly		bool
}

func CompileFile(path string) ([]Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	ps := []Pattern{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if p, ok := Compile(line); ok {
			ps = append(ps, p)
		}
	}
	return ps, sc.Err()
}

func Compile(line string) (Pattern, bool) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" || strings.HasPrefix(line, "#") {
		return Pattern{}, false
	}
	
	p := Pattern{}
	if strings.HasPrefix(line, "!") {
		p.negate = true
		line = strings.TrimPrefix(line, "!")
	}

	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	anchored := strings.Contains(line, "/")
	line = strings.TrimPrefix(line, "/")
	
	var sb strings.Builder
	if anchored {
		sb.WriteString("^")
	} else {
		sb.WriteString("^(?:.*/)?")
	}

	for i := 0; i < len(line); i++ {
		c := line[i]

		switch {
		case c == '*' && strings.HasPrefix(line[i:], "**/"):
			sb.WriteString("(?:.*/)?")
			i += 2
		case c == '*' && strings.HasPrefix(line[i:], "**") && i+2 == len(line):
			sb.WriteString(".*")
			i++
		case c == '*':
			sb.WriteString("[^/]*")
		case c == '?':
			sb.WriteString("[^/]")
		case c == '[':
			j := strings.IndexByte(line[i:], ']')
			if j < 0 {
				sb.WriteString("\\[")
				break
			}
			class := line[i+1 : i+j]
			if strings.HasPrefix(class, "!") {
				class = "^" + class[1:]
			}
			sb.WriteString("[")
			sb.WriteString(class)
			sb.WriteString("]")

			i += j
		case c == '\\' && i+1 < len(line):
			i++
			sb.WriteString(regexp.QuoteMeta(string(line[i])))
		default:
			sb.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	sb.WriteString("$")

	re, err := regexp.Compile(sb.String())
	if err != nil {
		return p, false
	}
	p.re = re
	return p, true
}
