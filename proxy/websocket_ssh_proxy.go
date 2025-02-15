package main

// set ANDROID_NDK_HOME=C:\Users\Jon\AppData\Local\Android\Sdk\ndk\21.3.6528147
// set ANDROID_HOME=C:\Users\Jon\AppData\Local\Android\Sdk

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/ssh"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	ncssh "github.com/nemith/netconf/transport/ssh"
)

// go run proxy/websocket_ssh_proxy.go

var wg sync.WaitGroup

func webToSSH(web net.Conn, ssh *ncssh.Transport) {
	defer wg.Done()
	// var n int
	// var err error
	writer, err := ssh.MsgWriter()
	if err != nil {
		log.Print("Failed to get writer: ", err)
		return
	}
	for {
		payload, err := wsutil.ReadClientBinary(web)

		if err != nil {
			log.Print("Failed to wsread: ", err)
			ssh.Close()
			return
		}
		log.Printf("Ws read: %v\n", string(payload))
		n, err := writer.Write(payload)
		if err != nil {
			log.Print("Failed to ncwrite: ", err)
			return
		} else {
			log.Printf("NC write: %d bytes\n", n)
		}

		// var output string
		// if strings.Compare(string(payload), "GetYang") == 0 {
		// 	// mods := GetModNames2()
		// 	output = strings.Join(mods, " ")
		// } else if strings.HasPrefix(string(payload), "GetEntries") {
		// 	// var mod string
		// 	// fmt.Sscanf(string(payload), "GetEntries: %s", &mod)
		// 	args := strings.Split(string(payload), " ")
		// 	// entries := GetEntries(args[1:])
		// 	entries = append([]string{":"}, entries...)
		// 	entries = append(args[1:], entries...)
		// 	entries = append([]string{"GetEntries"}, entries...)
		// 	fmt.Printf("GetEntries: %v\n", entries)
		// 	output = strings.Join(entries, " ")
		// } else {

		// Each ClientConn can support multiple interactive sessions,
		// represented by a Session.
		// session, err := ssh.NewSession()
		// if err != nil {
		// 	log.Fatal("Failed to create session: ", err)
		// }
		// defer session.Close()
		// output, err := session.CombinedOutput(string(payload))
		// if err != nil {
		// 	// handle err
		// 	log.Print("Failed to sshout: ", err)
		// } else {
		// 	println("SSH output: ", string(output), err)
		// }

		// err = wsutil.WriteServerText(web, []byte(output))
		// if err != nil {
		// 	log.Print("Failed to wsend: ", err)
		// }
	}
}

const (
	// msgSeperator is used to separate sent messages via NETCONF
	msgSeperator     = "]]>]]>"
	msgSeperator_v11 = "\n##\n"
)

func sshToWeb(web net.Conn, ssh *ncssh.Transport) {
	defer wg.Done()
	bytes := make([]byte, 1024*1024)
	// bytes2 := make([]byte, 1024*1024)
	var n, total int
	var err error
	// TODO Could we just use io.Copy()?
	// try_again:
	reader, err := ssh.MsgReader()
	if err != nil {
		log.Print("Failed to get reader: ", err)
		return
	}
	for {
		total = 0
		for {
			log.Debugln("Before read")
			n, err = reader.Read(bytes[total:])
			log.Debugln("After read")
			if n > 0 {
				// log.Printf("NC read: got %d bytes: %s\n", n, string(bytes))
				log.Debugf("NC read: got %d bytes\n", n)
				// bytes2 = append(bytes2, bytes[:n]...)
				total += n
				if strings.Contains(string(bytes), msgSeperator) ||
					strings.Contains(string(bytes), msgSeperator_v11) ||
					err == io.EOF {
					log.Debugf("NC read: got end marker\n")
					// log.Printf("NC read: %v \n%v\n", bytes, bytes2[total-4096:total])
					break
				}
			} else if err != nil {
				log.Printf("NC read err: %v\n", err)
				// return
				break
				// reader.Close()
				// goto try_again

			}
		}
		// log.Printf("Ws write: %d bytes %v\n", total, bytes[total-100:total])
		log.Printf("Ws write: %d bytes\n", total)
		err = wsutil.WriteServerBinary(web, bytes[:total])
		if err != nil {
			log.Printf("WS write err: %v\n", err)
		}
	}

}

func main() {
	log.SetLevel(log.DebugLevel)
	// var t transport.Transport
	// user, password := "cisco", "cisco123"
	user, password := "admin", "C1sco12345"
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// init
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		// handle error
		panic(err)
	}
	for {
		wg.Add(2)

		println("Waiting for connection...")
		conn, err := listener.Accept()
		if err != nil {
			// handle error
		}
		upgrader := ws.Upgrader{}
		if _, err = upgrader.Upgrade(conn); err != nil {
			// handle error
		}
		// err = t.Dial("sjc24lab-srv7:10007", sshConfig)
		// err = t.Dial("172.26.228.148:64374", sshConfig)
		transport2, err := ncssh.Dial(context.Background(), "tcp", "sandbox-iosxr-1.cisco.com:830", sshConfig)
		if err != nil {
			// t.Close()
			panic(err)
		} else {
			// defer t.Close()
		}
		println("Connected!")
		// var s *netconf.Session
		// s, err = netconf.Open(transport2)
		// if err != nil {
		// 	panic(err)
		// }

		go webToSSH(conn, transport2)
		go sshToWeb(conn, transport2)

		// wg.Wait()
	}
}
