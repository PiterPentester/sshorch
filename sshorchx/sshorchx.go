package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sshorch"
)

func usage() {
	sample := `
	---
	define-alias: |
	    my-machine = root@blackhost.com
	    friend-machine = joey@whitehost.com
	exec:
	    - login: my-machine
	      cmd: echo "Hello World"
	      out: Hello World
	      quiet-cmd: tar -xzvf foo.tar.gz
	    - login: alice@bob.com
	      cmd: |
	        echo "Alice"
	        hostname
	        uname -a
	---
	`
	fmt.Println("sshorchx: SSH Orchestration\n")
	fmt.Println("sshorchx <YAML file with sshorch commands>\n")
	fmt.Println("YAML file sample:")
	fmt.Println(sample)
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(-1)
	}

	// reading the yaml doc
	doc, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	// SSHOrch
	s := sshorch.NewSSHOrch()

	// parse the yaml doc
	d := sshorch.NewDoc()
	d.ParseYamlDoc(doc)
	//d.PrintDoc()
	for _, tuple := range d.GetAliasDefs() {
		// alias, user, host
		s.InitSSHConnection(tuple[0], tuple[1], tuple[2])
		fmt.Printf("Registered: %v\n", tuple)
	}

	for _, exec := range d.Exec {
		if s.LoginExists(exec.Login) == false {
			tuple, e := s.ValidateUserHost(exec.Login)
			if e == false {
				panic("ERROR: " + exec.Login + " : Must be an alias or user@host")
			}
			s.InitSSHConnection(tuple[0], tuple[1], tuple[2])
		}
		o := s.ExecSSH(exec.Login, exec.Cmd)
		fmt.Println(o)
	}

}
