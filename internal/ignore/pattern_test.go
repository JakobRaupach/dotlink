package ignore

import "testing"

func TestCompile(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		negate  bool
		dirOnly bool
		match   []string
		noMatch []string
	}{
		{
			name:    "simple wildcard matches at any depth",
			line:    "*.log",
			match:   []string{"a.log", "dir/a.log", "a/b/c.log"},
			noMatch: []string{"a.logx", "log", "a.log/b"},
		},
		{
			name:    "star does not cross slash",
			line:    "dir/*.go",
			match:   []string{"dir/a.go"},
			noMatch: []string{"dir/sub/a.go", "x/dir/a.go"},
		},
		{
			name:    "leading slash anchors to root",
			line:    "/build",
			match:   []string{"build"},
			noMatch: []string{"src/build", "builds"},
		},
		{
			name:    "slash in the middle anchors to root",
			line:    "src/gen",
			match:   []string{"src/gen"},
			noMatch: []string{"x/src/gen", "src/gen/x"},
		},
		{
			name:    "trailing slash is directory only",
			line:    "node_modules/",
			dirOnly: true,
			match:   []string{"node_modules", "a/node_modules"},
			noMatch: []string{"node_modules_old"},
		},
		{
			name:    "negation",
			line:    "!keep.log",
			negate:  true,
			match:   []string{"keep.log", "a/keep.log"},
			noMatch: []string{"other.log"},
		},
		{
			name:    "leading double star",
			line:    "**/foo",
			match:   []string{"foo", "a/foo", "a/b/foo"},
			noMatch: []string{"afoo", "foo/a"},
		},
		{
			name:    "trailing double star",
			line:    "foo/**",
			match:   []string{"foo/a", "foo/a/b"},
			noMatch: []string{"foo", "bar/foo/a"},
		},
		{
			name:    "middle double star matches zero or more dirs",
			line:    "a/**/b",
			match:   []string{"a/b", "a/x/b", "a/x/y/b"},
			noMatch: []string{"a/xb", "b", "x/a/b"},
		},
		{
			name:    "double star inside a segment acts like star",
			line:    "foo**bar",
			match:   []string{"foobar", "fooXbar"},
			noMatch: []string{"foo/bar"},
		},
		{
			name:    "question mark matches one non-slash char",
			line:    "file?.txt",
			match:   []string{"file1.txt", "dir/fileA.txt"},
			noMatch: []string{"file.txt", "file12.txt"},
		},
		{
			name:    "character class",
			line:    "[abc].txt",
			match:   []string{"a.txt", "c.txt"},
			noMatch: []string{"d.txt", "ab.txt"},
		},
		{
			name:    "negated character class",
			line:    "[!abc].txt",
			match:   []string{"d.txt", "z.txt"},
			noMatch: []string{"a.txt", "b.txt"},
		},
		{
			name:    "unclosed bracket is literal",
			line:    "[abc",
			match:   []string{"[abc"},
			noMatch: []string{"a"},
		},
		{
			name:    "dots are literal",
			line:    "a.b",
			match:   []string{"a.b"},
			noMatch: []string{"axb"},
		},
		{
			name:    "escaped hash is literal",
			line:    `\#file`,
			match:   []string{"#file"},
			noMatch: []string{"file"},
		},
		{
			name:    "trailing carriage return is ignored",
			line:    "*.log\r",
			match:   []string{"a.log"},
			noMatch: []string{"a.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := Compile(tt.line)
			if !ok {
				t.Fatalf("compile(%q) returned ok=false", tt.line)
			}
			if p.negate != tt.negate {
				t.Errorf("negate = %v, want %v", p.negate, tt.negate)
			}
			if p.dirOnly != tt.dirOnly {
				t.Errorf("dirOnly = %v, want %v", p.dirOnly, tt.dirOnly)
			}
			for _, path := range tt.match {
				if !p.re.MatchString(path) {
					t.Errorf("%q should match %q (regex %s)", tt.line, path, p.re)
				}
			}
			for _, path := range tt.noMatch {
				if p.re.MatchString(path) {
					t.Errorf("%q should not match %q (regex %s)", tt.line, path, p.re)
				}
			}
		})
	}
}

func TestCompileSkipsBlankAndComments(t *testing.T) {
	for _, line := range []string{"", "\r\n", "# a comment", "#"} {
		if _, ok := Compile(line); ok {
			t.Errorf("compile(%q) should be skipped", line)
		}
	}
}
