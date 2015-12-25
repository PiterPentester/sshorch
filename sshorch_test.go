package sshorch_test

import (
	"fmt"
	"log"
	"os/user"
	"sshorch"
)

var execStr = `
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
        uname
`

// ExampleYamlParser demo yaml parser
func ExampleYamlParser() {
	d := sshorch.NewDoc()
	d.ParseYamlDoc(execStr)
	d.PrintDoc()
	// Output: {AliasDefs:my-machine = root@blackhost.com
	// friend-machine = joey@whitehost.com
	//  Exec:[{Login:my-machine Cmd:echo "Hello World" Out:Hello World QuiteCmd:tar -xzvf foo.tar.gz} {Login:alice@bob.com Cmd:echo "Alice"
	// hostname
	// uname
	//  Out: QuiteCmd:}]}
	// map[my-machine:root@blackhost.com friend-machine:joey@whitehost.com]
	// [[my-machine root blackhost.com] [friend-machine joey whitehost.com]]
}

// ExampleSSHOrch demo sshorch
func ExampleSSHOrch() {
	curr, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	username := curr.Username

	s := sshorch.NewSSHOrch()
	s.InitSSHConnection("localAlias", username, "localhost")
	//s.ShowClientMap()
	fmt.Println(s.ExecSSH(username+"@localhost", "echo 'Hello World' | md5"))
	fmt.Println(s.ExecSSH("localAlias", "echo 'Hello World' | md5"))
	// Output: e59ff97941044f85df5297e1c302d260
	//
	// e59ff97941044f85df5297e1c302d260
}
