package main

import (
	"code.google.com/p/go.crypto/ssh"
	. "github.com/aerospike/gommander"

	"bytes"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Dump function for printing fixed width columns
func column(s string, w int) string {
	o := bytes.Buffer{}
	i := 0
	for ; i < len(s); i++ {
		o.WriteRune(rune(s[i]))
		if i >= w {
			break
		}
	}
	for j := i; j < w; j++ {
		o.WriteRune(' ')
	}
	return o.String()
}

// Iterate over responses recevied for a command
// and print it to stdout
func printResponses(responses chan Response, err error) {

	const layout = "2006-01-02 15:04:05 -0700"

	date := time.Now().Format(layout)
	sep := "]"

	if err != nil {
		panic(err)
	}

	for r := range responses {

		println("--------------------------------------------------------------------------------")

		exit := column(fmt.Sprintf("%d", r.ExitCode), 3)
		host := column(r.Node.Host, 15)

		epre := fmt.Sprintf("\033[0;91m%s %s %s %s %s  \033[0m", date, host, column(" ERR", 5), exit, sep)
		opre := fmt.Sprintf("\033[0;90m%s %s %s %s %s  \033[0m", date, host, column(" OUT", 5), exit, sep)

		ostr := strings.Trim(r.Stdout.String(), "\r\n ")
		if len(ostr) > 0 {
			for _, s := range strings.Split(ostr, "\n") {
				fmt.Println(opre + s)
			}
		} else {
			if r.ExitCode == 0 {
				fmt.Println(opre + "<SUCCESS>")
			}
		}

		estr := strings.Trim(r.Stderr.String(), "\r\n ")
		if len(estr) > 0 {
			for _, s := range strings.Split(estr, "\n") {
				fmt.Println(epre + s)
			}
		} else {
			if r.ExitCode != 0 {
				fmt.Println(epre + "<FAILURE>")
			}
		}
	}
}

// Print the description, run the command, then print the results... easy
func run(desc string, fn func() (chan Response, error)) {

	println("################################################################################")
	println(">", desc)
	resp, err := fn()
	printResponses(resp, err)
}

func parseCred(s string) (user string, pass string) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 2:
		user = parts[0]
		pass = parts[1]
	case 1:
		user = parts[0]
		pass = ""
	default:
		user = ""
		pass = ""
	}
	return
}

func parseHost(s string) (host string, port int) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 2:
		host = parts[0]
		port, _ = strconv.Atoi(parts[1])
	case 1:
		host = parts[0]
		port = 22
	default:
		host = ""
		port = 0
	}
	return
}

func parseCredHost(s string) (host string, port int, user string, pass string) {
	parts := strings.Split(s, "@")
	switch len(parts) {
	case 2:
		user, pass = parseCred(parts[0])
		host, port = parseHost(parts[1])
	case 1:
		host, port = parseHost(parts[0])
		user = ""
		pass = ""
	default:
		user = ""
		pass = ""
		host = ""
		port = 0
	}
	return
}

//
// simple -k /path/to/sshkey user@hostname user:pass@hostname
//
func main() {

	username := flag.String("username", "", "Username for connecting to servers.")
	password := flag.String("password", "", "Password for connecting to servers.")
	sshkeyfile := flag.String("sshkey", "", "SSH Key file for connecting to servers.")
	flag.Parse()

	// SSH Key Auth
	var sshkey *ssh.Signer

	if len(*sshkeyfile) > 0 {
		sshkeySigner, err := Parsekey(*sshkeyfile)
		if err != nil {
			panic(err)
		}
		sshkey = &sshkeySigner

	}

	// Define the set of nodes we are using.
	// The following are private IPs for Vagrant
	// Change, remove or add new ones to your heart's desire
	nodes := NodeList{}

	for _, arg := range flag.Args() {

		authMethods := []ssh.AuthMethod{}

		if sshkey != nil {
			authMethods = append(authMethods, ssh.PublicKeys(*sshkey))
		}

		host, port, user, pass := parseCredHost(arg)

		if len(host) == 0 {
			panic(fmt.Sprintf("Invalid host arg: %s", host))
		}

		if len(pass) > 0 {
			authMethods = append(authMethods, ssh.Password(pass))
		} else if len(*password) > 0 {
			authMethods = append(authMethods, ssh.Password(*password))
		}

		if len(user) == 0 && len(*username) > 0 {
			user = *username
		}

		// This is the important step
		nodes = append(nodes, NewNode(host, uint(port), user, authMethods))
	}

	// Connect to the nodes
	nodes.Connect()

	// What is the hostname?
	run("hostname", func() (chan Response, error) {
		return nodes.Run("hostname")
	})

	// What files are there?
	run("ls", func() (chan Response, error) {
		return nodes.Run("ls")
	})

	// Let's see if pipes work
	run("echo \"abc\" | tr \"a-z\" \"A-Z\"", func() (chan Response, error) {
		return nodes.Run("echo \"abc\" | tr \"a-z\" \"A-Z\"")
	})

	// Copy a file to the servers
	run("copy hello.txt", func() (chan Response, error) {
		return nodes.CopyFile("hello.txt", "hello.txt")
	})

	// Read the file contents
	run("cat hello.txt", func() (chan Response, error) {
		return nodes.Run("cat hello.txt")
	})

	// remove the script
	run("rm hello.txt", func() (chan Response, error) {
		return nodes.Run("rm hello.txt")
	})

	// Copy an script to the servers
	run("copy hello.py", func() (chan Response, error) {
		return nodes.Take(2).CopyFile("hello.py", "hello.py")
	})

	// make it executable
	run("chmod +x hello.py", func() (chan Response, error) {
		return nodes.Run("chmod +x hello.py")
	})

	// run the script
	run("run ./hello.py", func() (chan Response, error) {
		return nodes.Run("./hello.py")
	})

	// remove the script
	run("rm hello.py", func() (chan Response, error) {
		return nodes.Run("rm hello.py")
	})

	// What files are there?
	run("ls", func() (chan Response, error) {
		return nodes.Run("ls")
	})

	// run the script ... better fail
	run("run ./hello.py (should fail)", func() (chan Response, error) {
		return nodes.Run("./hello.py")
	})

	// close connections to cluster
	nodes.Close()
}
