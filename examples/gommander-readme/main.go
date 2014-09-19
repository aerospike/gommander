package main

import (
	"code.google.com/p/go.crypto/ssh"
	. "github.com/aerospike/gommander"

	"fmt"
)

func printResponse(responses chan Response, err error) {
	if err != nil {
		panic(err)
	}

	for r := range responses {
		switch r.ExitCode {
		case 0: // Success!
			fmt.Printf("\033[0;97m%s > %s\033[0m\n", r.Node.Host, r.Stdout.String())
		default: // Failure!
			fmt.Printf("\033[0;91m%s > %s\033[0m\n", r.Node.Host, r.Stderr.String())
		}
	}
}

func main() {

	// Parse the ssh key. This is the Vagrant SSH private key.
	vagrantKey, err := Parsekey("/Users/ME/.vagrant.d/insecure_private_key")
	if err != nil {
		panic(err)
	}

	// Array of auth methods, needed for authenticating w/ each node
	authMethods := []ssh.AuthMethod{ssh.PublicKeys(vagrantKey)}

	// Define a cluster
	cluster := NewCluster()
	cluster.AddNode(NewNode("127.0.0.1", 2221, "vagrant", authMethods))
	cluster.AddNode(NewNode("127.0.0.1", 2222, "vagrant", authMethods))

	// Slice of all nodes
	nodes := cluster.Nodes()

	// Connect to the cluster
	nodes.Connect()

	// Say Hello
	printResponse(nodes.Run("echo hello"))

	// Close the connections to the cluster
	nodes.Close()

}
