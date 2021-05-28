package render

import (
	"bufio"
	"html/template"
	"os"
	"regexp"
)

var extendsRegex *regexp.Regexp

func init() {
	var err error
	extendsRegex, err = regexp.Compile(`\{\{\/\* *?extends +?"(.+?)" *?\*\/\}\}`)
	if err != nil {
		panic(err)
	}
}

type templates struct {
	set map[string]*template.Template
}

// Lookup returns the compiled template by its filename or nil if there is no such template
func (t *templates) Lookup(name string) *template.Template {
	if t == nil {
		return nil
	}
	return t.set[name]
}

// getLayoutForTemplate scans the first line of the template file for the extends keyword
func getLayoutForTemplate(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	b := scanner.Bytes()
	if l := extendsRegex.FindSubmatch(b); l != nil {
		return string(l[1])
	}

	return ""
}
