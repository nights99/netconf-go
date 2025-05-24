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
)

// go run proxy/websocket_ssh_proxy.go

var wg sync.WaitGroup
var helloDone = false
var cond = sync.NewCond(&mu)
var mu sync.Mutex

type sshNetconfConn struct {
	conn    net.Conn
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
}

func (c *sshNetconfConn) Reader() io.Reader {
	return c.stdout
}

func (c *sshNetconfConn) Writer() io.Writer {
	return c.stdin
}

func (c *sshNetconfConn) Close() error {
	if c.session != nil {
		c.session.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// localDial establishes an SSH connection and requests the netconf subsystem.
func localDial(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*sshNetconfConn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	sess, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, err
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}
	if err := sess.RequestSubsystem("netconf"); err != nil {
		sess.Close()
		return nil, err
	}
	return &sshNetconfConn{
		conn:    conn,
		session: sess,
		stdin:   stdin,
		stdout:  stdout,
	}, nil
}

func webToSSH(web net.Conn, ssh *sshNetconfConn) {
	defer wg.Done()
	writer := ssh.Writer()
	for {
		payload, err := wsutil.ReadClientBinary(web)

		if err != nil {
			log.Print("Failed to wsread: ", err)
			ssh.Close()
			return
		}
		log.Printf("Ws read: %v\n", string(payload))
		if !helloDone {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()
		}
		n, err := writer.Write(payload)
		if err != nil {
			log.Print("Failed to ncwrite: ", err)
			return
		} else {
			log.Printf("NC write: %d bytes\n", n)
			if f, ok := writer.(interface{ Flush() error }); ok {
				f.Flush()
			}
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

func sshToWeb(web net.Conn, ssh *sshNetconfConn) {
	defer wg.Done()
	bytes := make([]byte, 1024*1024)
	var n, total int
	// TODO Could we just use io.Copy()?
	// try_again:
	reader := ssh.Reader()
	for {
		var err error
		total = 0
		for {
			log.Debugln("Before read")
			n, err = reader.Read(bytes[total:])
			log.Debugln("After read")
			if n > 0 {
				// log.Printf("NC read: got %d bytes: %s\n", n, string(bytes))
				log.Debugf("NC read: got %d bytes %v\n", n, err)
				// bytes2 = append(bytes2, bytes[:n]...)
				total += n
				if total > 4096 {
					log.Printf("NC read: %v \n%v\n", string(bytes), string(bytes[total-4096:total]))
				} else {
					log.Printf("NC read: %v\n", string(bytes))
				}
				if strings.Contains(string(bytes), msgSeperator) ||
					strings.Contains(string(bytes), msgSeperator_v11) ||
					err == io.EOF {
					log.Debugf("NC read: got end marker, %v\n", err)
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
		if !helloDone {
			helloDone = true
			cond.Broadcast()
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
		sshConn, err := localDial(context.Background(), "tcp", "sandbox-iosxr-1.cisco.com:830", sshConfig)
		if err != nil {
			// t.Close()
			panic(err)
		} else {
			// defer t.Close()
		}
		println("Connected!")
		go webToSSH(conn, sshConn)
		go sshToWeb(conn, sshConn)
	}
}
