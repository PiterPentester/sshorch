package sshorch

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// Doc denotes the Yaml document
type Doc struct {
	AliasDefs string `yaml:"define-alias"`
	Exec      []struct {
		Login    string
		Cmd      string
		Out      string
		QuiteCmd string `yaml:"quiet-cmd"`
	}
}

// NewDoc is constructor for Doc
func NewDoc() *Doc {
	return new(Doc)
}

// ParseAliasDefs returns the map os defined aliases
func (e *Doc) ParseAliasDefs() map[string]string {
	aliasDefs := e.AliasDefs
	aliasMap := make(map[string]string)
	lineScanner := bufio.NewScanner(strings.NewReader(aliasDefs))
	for lineScanner.Scan() {
		def := strings.TrimSpace(lineScanner.Text())
		re := regexp.MustCompile(`\S+\s*=\s*\S+`)
		if re.MatchString(def) == false {
			log.Fatalln("Invalid alias definition:", def)
		}
		tok := strings.Split(def, "=")
		if len(tok) != 2 {
			log.Fatalln("Invalid alias definition:", def)
		}
		lval := strings.TrimSpace(tok[0])
		rval := strings.TrimSpace(tok[1])
		aliasMap[lval] = rval
	}
	return aliasMap
}

// ParseYamlDoc parses the exec/doc
func (e *Doc) ParseYamlDoc(execRune []byte) {
	err := yaml.Unmarshal([]byte(execRune), e)
	if err != nil {
		log.Fatalln("Failed to unmarshal yaml string:", err)
	}
}

// GetAliasDefs returns array of 3-tuple alias, username, hostname
func (e *Doc) GetAliasDefs() [][]string {
	ret := [][]string{}
	lineScanner := bufio.NewScanner(strings.NewReader(e.AliasDefs))
	lineScanner.Split(bufio.ScanLines)
	for lineScanner.Scan() {
		line := lineScanner.Text()
		wordScanner := bufio.NewScanner(strings.NewReader(line))
		wordScanner.Split(bufio.ScanWords)

		var alias, eq, userhost, user, host string

		if wordScanner.Scan() {
			alias = wordScanner.Text()
		} else {
			panic("ERROR: " + line)
		}

		if wordScanner.Scan() {
			eq = wordScanner.Text()
		} else {
			panic("ERROR: " + line)
		}
		if eq != "=" {
			panic("ERROR: " + line)
		}

		if wordScanner.Scan() {
			userhost = wordScanner.Text()
		} else {
			panic("ERROR: " + line)
		}

		uh := strings.Split(userhost, "@")
		if len(uh) != 2 {
			panic("ERROR: " + line)
		}

		user = uh[0]
		host = uh[1]

		ret = append(ret, []string{alias, user, host})
	} // lineScanner loop end
	return ret
}

// PrintDoc displays the Doc
func (e *Doc) PrintDoc() {
	fmt.Printf("%+v\n", *e)
	fmt.Printf("%+v\n", e.ParseAliasDefs())
	fmt.Printf("%+v\n", e.GetAliasDefs())
}
