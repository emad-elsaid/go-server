package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var input = flag.String("input", "tags.txt", "list of inputs")
var output = flag.String("output", "tags.go", "output of tags code")
var pkg = flag.String("pkg", "tag", "package name the output belongs to")
var tpl = flag.String("template", "tag", "template to use either tag or attr")

var imports = map[string][]string{
	"tag":  {`import "fmt"`},
	"attr": {""},
}

var tpls = map[string]string{
	"tag":  `func %s(c ...fmt.Stringer) Element { return Tag("%s", c...) }`,
	"attr": `func %s(v string) Attribute { return Attr("%s", v) }`,
}

func main() {
	flag.Parse()

	template, ok := tpls[*tpl]
	if !ok {
		panic("template is not found")
	}

	// Read the file content
	f, err := os.Open(*input)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	newLines := []string{
		"package " + *pkg,
	}

	importsFortpl := imports[*tpl]
	for _, v := range importsFortpl {
		newLines = append(newLines, v)
	}

	charsToRemove := regexp.MustCompile(`[\-:]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fname := strings.Title(line)
		fname = charsToRemove.ReplaceAllString(fname, "")

		newLine := fmt.Sprintf(template, fname, line)
		newLines = append(newLines, newLine)
	}

	os.WriteFile(*output, []byte(strings.Join(newLines, "\n")), 0777)
}
