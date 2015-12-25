package sshorch

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

// SSHOrch is the SSH Orchestration struct
type SSHOrch struct {
	lookup   map[string]*ssh.Client
	aliasMap map[string]string
}

// NewSSHOrch returns the initialized SSHOrch struct
func NewSSHOrch() *SSHOrch {
	s := new(SSHOrch)
	s.lookup = make(map[string]*ssh.Client)
	s.aliasMap = make(map[string]string)
	return s
}

func (s *SSHOrch) addLookup(username, server string, client *ssh.Client) {
	s.lookup[username+"@"+server] = client
}

func (s *SSHOrch) addLookupAlias(alias string, client *ssh.Client) {
	s.lookup[alias] = client
}

func (s *SSHOrch) doLookup(tag string) *ssh.Client {
	client := s.lookup[tag]
	if client == nil {
		log.Fatalln("Illegal user or hostname:", tag)
	}
	return client
}

// ShowClientMap displays the lookup
func (s *SSHOrch) ShowClientMap() {
	for k, v := range s.lookup {
		fmt.Printf("%s:%s\n", k, v.LocalAddr())
	}
}

func sshRegister(username, server string) {
	idFile := os.Getenv("HOME") + "/.ssh/id_rsa.pub"
	if _, err := os.Stat(idFile); os.IsNotExist(err) {
		log.Println("File " + idFile + " does not exist")
	}

	fmt.Println("One time setup for ", username+"@"+server, "!")
	fmt.Print("Password: ")
	passwd, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatalln("Failed to read password:", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(string(passwd))},
	}

	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatalln("Failed to dial:", err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatalln("Failed to create session:", err)
	}
	defer session.Close()

	idRsaPub, err := ioutil.ReadFile(idFile)
	if err != nil {
		log.Fatalln("Failed to read"+idFile+" for user "+username+":", err)
	}

	// ssh-copy-id might not be present on old systems
	cmd := "echo \"" + string(idRsaPub) + "\" > tmp.pubkey;" +
		"mkdir -p .ssh;" +
		"touch .ssh/authorized_keys;" +
		"sed -i.bak -e '/" + username + "@" + server + "/d' .ssh/authorized_keys;" +
		"cat tmp.pubkey >> .ssh/authorized_keys;" +
		"rm tmp.pubkey;"

	if err := session.Run(cmd); err != nil {
		log.Fatalln(err)
	}
}

// LoginExists tells if the userhost/alias already exists in the lookup
func (s *SSHOrch) LoginExists(login string) bool {
	return s.lookup[login] != nil
}

// ValidateUserHost checks user@host format
func (s *SSHOrch) ValidateUserHost(uh string) ([]string, bool) {
	_uh := append([]string{""}, strings.Split(uh, "@")...)
	e := len(_uh) == 3 &&
		len(strings.Split(_uh[1], " ")) == 1 &&
		len(strings.Split(_uh[2], " ")) == 1
	return _uh, e
}

// ExecSSH executes the command cmd on the server for the given user
func (s *SSHOrch) ExecSSH(userserver, cmd string) string {
	client := s.doLookup(userserver)

	session, err := client.NewSession()
	if err != nil {
		log.Fatalln("Failed to create session:", err)
	}
	defer session.Close()
	/*
		stdout, err := session.StdoutPipe()
		if err != nil {
			log.Fatalln("Failed to pipe session stdout:", err)
		}

		stderr, err := session.StderrPipe()
		if err != nil {
			log.Fatalln("Failed to pipe session stderr:", err)
		}
	*/

	buf, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Fatalln("Failed to execute cmd:", err)
	}

	// Network read pushed to background
	/*readExec := func(r io.Reader, ch chan []byte) {
		if str, err := ioutil.ReadAll(r); err != nil {
			ch <- str
		}
	}
	outCh := make(chan []byte)
	go readExec(stdout, outCh)
	*/
	return string(buf)
}

// InitSSHConnection initializes the SSH connections with the server
// for given user
func (s *SSHOrch) InitSSHConnection(alias, username, server string) {
	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ag := agent.NewClient(conn)
	auths := []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)}

	config := &ssh.ClientConfig{
		User: username,
		Auth: auths,
	}
relogin:
	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		log.Println(err)
		sshRegister(username, server)
		goto relogin
	}
	s.addLookup(username, server, client)
	if len(alias) > 0 {
		s.addLookupAlias(alias, client)
	}
}
