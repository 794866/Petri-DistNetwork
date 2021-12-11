package utils

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	SshPort        = "22"
	PrivateKeyPath = "/home/uri/.ssh/id_rsa"
)

func ConnectSSH(user string, host string) *ssh.Client {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{getPublicKey(PrivateKeyPath),},
		// allow any host key to be used (non-prod)
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoDSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
		// optional tcp connect timeout
		Timeout: 5 * time.Second,
	}
	// Connect via ssh
	client, err := ssh.Dial("tcp", host+":"+SshPort, config)
	if err != nil {
		panic(err)
	}
	return client
}

// From: https://medium.com/tarkalabs/ssh-recipes-in-go-part-one-5f5a44417282
func getPublicKey(path string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}

// Source: https://medium.com/tarkalabs/ssh-recipes-in-go-part-one-5f5a44417282
func RunCommandSSH(cmd string, conn *ssh.Client, wg *sync.WaitGroup) {
	// start session
	sess, err := conn.NewSession()
	if err != nil {
		panic(err)
	}
	//ending sess at the end
	defer sess.Close()
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	// Execute command
	err = sess.Run(cmd)
	if wg != nil {
		wg.Done()
	}
}
