package main

// set ANDROID_NDK_HOME=C:\Users\Jon\AppData\Local\Android\Sdk\ndk\21.3.6528147
// set ANDROID_HOME=C:\Users\Jon\AppData\Local\Android\Sdk

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// go run proxy/websocket_ssh_proxy.go

var wg sync.WaitGroup

func webToSSH(web net.Conn, ssh *ssh.Client) {
	defer wg.Done()
	for {
		payload, err := wsutil.ReadClientText(web)
		if err != nil {
			// handle err
			log.Print("Failed to wsread: ", err)
			return
		}
		println(string(payload))

		var output string
		if strings.Compare(string(payload), "GetYang") == 0 {
			mods := GetModNames2()
			output = strings.Join(mods, " ")
		} else if strings.HasPrefix(string(payload), "GetEntries") {
			// var mod string
			// fmt.Sscanf(string(payload), "GetEntries: %s", &mod)
			args := strings.Split(string(payload), " ")
			entries := GetEntries(args[1:])
			entries = append([]string{":"}, entries...)
			entries = append(args[1:], entries...)
			entries = append([]string{"GetEntries"}, entries...)
			fmt.Printf("GetEntries: %v\n", entries)
			output = strings.Join(entries, " ")
		} else {
			// Each ClientConn can support multiple interactive sessions,
			// represented by a Session.
			session, err := ssh.NewSession()
			if err != nil {
				log.Fatal("Failed to create session: ", err)
			}
			defer session.Close()
			output, err := session.CombinedOutput(string(payload))
			if err != nil {
				// handle err
				log.Print("Failed to sshout: ", err)
			} else {
				println("SSH output: ", string(output), err)
			}
		}
		err = wsutil.WriteServerText(web, []byte(output))
		if err != nil {
			log.Print("Failed to wsend: ", err)
		}
	}
}

func sshToWeb(web net.Conn, ssh *ssh.Client) {
	defer wg.Done()

}

func main() {

	key, err := ioutil.ReadFile("/home/jon/.ssh/id_new")
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	client, err := ssh.Dial("tcp", "localhost:22",
		&ssh.ClientConfig{
			User: "jon",
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey()})
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}

	// init
	listener, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		// handle error
	}
	for {
		wg.Add(1)

		conn, err := listener.Accept()
		if err != nil {
			// handle error
		}
		upgrader := ws.Upgrader{}
		if _, err = upgrader.Upgrade(conn); err != nil {
			// handle error
		}
		go webToSSH(conn, client)
		//go sshToWeb(conn, client)
		wg.Wait()
	}
}
