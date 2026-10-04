package ignore

type Matcher struct {
	patterns	[]Pattern
}

func NewMatcher(patterns []Pattern) *Matcher {
	m := &Matcher{}
	m.patterns = patterns
	return m
}


func (m *Matcher) Ignored(rel string, isDir bool) bool {
	ignored := false
	for _, p := range m.patterns {
		if p.dirOnly && !isDir {
			continue
		}
		if p.re.MatchString(rel) {
			ignored = !p.negate
		}
	}
	return ignored
}
