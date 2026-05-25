package main

// set ANDROID_NDK_HOME=C:\Users\Jon\AppData\Local\Android\Sdk\ndk\21.3.6528147
// set ANDROID_HOME=C:\Users\Jon\AppData\Local\Android\Sdk

import (
	"context"
	"io"
	"net"
	"netconf-go/internal/cliargs"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

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

type proxySession struct {
	ctx    context.Context
	cancel context.CancelFunc
	web    net.Conn
	ssh    *sshNetconfConn
	once   sync.Once
}

func (ps *proxySession) close() {
	ps.once.Do(func() {
		ps.cancel()
		if ps.web != nil {
			ps.web.Close()
		}
		if ps.ssh != nil {
			ps.ssh.Close()
		}
	})
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
func localDial(_ context.Context, network, addr string, config *ssh.ClientConfig) (*sshNetconfConn, error) {
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

func webToSSH(ps *proxySession) {
	defer wg.Done()
	writer := ps.ssh.Writer()

	for {
		select {
		case <-ps.ctx.Done():
			return
		default:
		}

		payload, err := wsutil.ReadClientBinary(ps.web)
		if err != nil {
			log.Printf("web read failed: %v", err)
			ps.close()
			return
		}
		log.Printf("Ws read: %v\n", string(payload))
		if !helloDone {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()
		}

		if _, err := writer.Write(payload); err != nil {
			log.Printf("Failed to ncwrite: %v", err)
			ps.close()
			return
		}
		if f, ok := writer.(interface{ Flush() error }); ok {
			f.Flush()
		}
	}
}

const (
	// msgSeperator is used to separate sent messages via NETCONF
	msgSeperator     = "]]>]]>"
	msgSeperator_v11 = "\n##\n"
)

func sshToWeb(ps *proxySession) {
	defer wg.Done()
	bytes := make([]byte, 1024*1024)
	reader := ps.ssh.Reader()

	for {
		select {
		case <-ps.ctx.Done():
			return
		default:
		}

		total := 0
		var err error
		for {
			log.Debugln("Before read")
			n, readErr := reader.Read(bytes[total:])
			log.Debugln("After read")
			if n > 0 {
				log.Debugf("NC read: got %d bytes %v\n", n, readErr)
				total += n
				if total > 4096 {
					log.Printf("NC read: %v \n%v\n", string(bytes), string(bytes[total-4096:total]))
				} else {
					log.Printf("NC read: %v\n", string(bytes[:total]))
				}
				if strings.Contains(string(bytes[:total]), msgSeperator) ||
					strings.Contains(string(bytes[:total]), msgSeperator_v11) ||
					readErr == io.EOF {
					log.Debugf("NC read: got end marker, %v\n", readErr)
					err = readErr
					break
				}
			}
			if readErr != nil {
				log.Printf("NC read err: %v\n", readErr)
				err = readErr
				break
			}
		}

		if !helloDone {
			helloDone = true
			cond.Broadcast()
		}

		log.Printf("Ws write: %d bytes\n", total)
		if total > 0 {
			if writeErr := wsutil.WriteServerBinary(ps.web, bytes[:total]); writeErr != nil {
				log.Printf("WS write err: %v\n", writeErr)
				ps.close()
				return
			}
		}
		if err != nil {
			ps.close()
			return
		}
	}
}

func main() {
	cliargs.AddFlags(pflag.CommandLine)
	pflag.Parse()

	cfg, err := cliargs.Load(pflag.CommandLine, ".")
	if err != nil {
		panic(err)
	}

	level, err := log.ParseLevel(cfg.Debug)
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)

	targetAddr := net.JoinHostPort(cfg.Address, strconv.Itoa(cfg.Port))
	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
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
		sshConn, err := localDial(context.Background(), "tcp", targetAddr, sshConfig)
		if err != nil {
			// t.Close()
			panic(err)
		} else {
			// defer t.Close()
		}
		println("Connected to ssh")
		println("Waiting for websocket connection")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		upgrader := ws.Upgrader{}
		if _, err = upgrader.Upgrade(conn); err != nil {
			log.Printf("upgrade error: %v", err)
			conn.Close()
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		ps := &proxySession{
			ctx:    ctx,
			cancel: cancel,
			web:    conn,
			ssh:    sshConn,
		}

		wg.Add(2)
		go webToSSH(ps)
		go sshToWeb(ps)

		go func() {
			<-ctx.Done()
			log.Println("connection closed")
		}()
	}
}
