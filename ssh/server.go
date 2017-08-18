package ssh

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nuveo/log"
	"golang.org/x/crypto/ssh"
)

// FakeServer SSH
type FakeServer struct {
	t          *testing.T
	conf       *ssh.ServerConfig
	listener   net.Listener
	Cmd        string
	Reply      string
	ConnDelay  time.Duration
	ExecDelay  time.Duration
	ExitStatus int
}

// ProbeConnection server
func ProbeConnection(ip string, maxRetries int) error {
	counter := 0
	var (
		conn net.Conn
		err  error
	)
	for counter < maxRetries {
		conn, err = net.DialTimeout("tcp", ip+Port, time.Duration(500)*time.Millisecond)
		if err == nil {
			return nil
		}
		counter++
		time.Sleep(time.Duration(250) * time.Millisecond)
	}

	if conn != nil {
		err = conn.Close()
	}
	return err
}

// NewServer creation
func NewServer(t *testing.T, addr, cmd, reply string, exitStatus int, execDelay, connDelay time.Duration) (server *FakeServer, err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	server = &FakeServer{
		t:          t,
		conf:       &ssh.ServerConfig{},
		listener:   listener,
		Cmd:        cmd,
		Reply:      reply,
		ConnDelay:  connDelay,
		ExecDelay:  execDelay,
		ExitStatus: exitStatus,
	}
	server.start()
	return
}

// Stop fake server
func (s *FakeServer) Stop() {
	if s.listener != nil {
		s.listener.Close() // nolint
	}
}

// Address server
func (s *FakeServer) Address() (addr string) {
	addr = s.listener.Addr().String()
	return
}

func (s *FakeServer) start() {
	byt, err := ioutil.ReadFile(filepath.Join(KeysDir, PrivateKeyName))
	if err != nil {
		err = GenerateFNSSHKey(4096)
		if err != nil {
			s.Stop()
			s.t.Fatal(err)
		}
		byt, err = ioutil.ReadFile(filepath.Join(KeysDir, PrivateKeyName))
		if err != nil {
			s.Stop()
			s.t.Fatal(err)
		}
	}
	k, err := ssh.ParsePrivateKey(byt)
	if err != nil {
		s.Stop()
		s.t.Fatal(fmt.Errorf("Could not parse private key: %s", err.Error()))
	}
	pub := k.PublicKey()
	s.conf.AddHostKey(k)
	certChecker := ssh.CertChecker{
		IsUserAuthority: func(k ssh.PublicKey) bool {
			return bytes.Equal(k.Marshal(), pub.Marshal())
		},
		UserKeyFallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() == "root" && bytes.Equal(key.Marshal(), pub.Marshal()) {
				return nil, nil
			}

			return nil, fmt.Errorf("pubkey for %q not acceptable", conn.User())
		},
		IsRevoked: func(c *ssh.Certificate) bool {
			return false
		},
	}
	s.conf.PublicKeyCallback = certChecker.Authenticate
	go s.handleConnections()
}

func (s *FakeServer) handleConnections() {
	for {
		if s.ConnDelay > 0 {
			time.Sleep(s.ConnDelay)
		}
		tcpConn, err := s.listener.Accept()
		if err != nil {
			panic(fmt.Errorf("Failed to accept incoming connection: %s", err))
		}
		_, chans, reqs, err := ssh.NewServerConn(tcpConn, s.conf)
		if err != nil {
			panic(fmt.Errorf("Handshake failed: %s", err))
		}
		go ssh.DiscardRequests(reqs)
		go s.handleChannels(chans)
	}
}

func (s *FakeServer) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go s.handleChannel(newChannel)
	}
}

type channelRequestSuccessMsg struct {
	PeersID uint32 `sshtype:"99"`
}

func (s *FakeServer) handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	ch, requests, err := newChannel.Accept()
	if err != nil {
		s.Stop()
		log.Printf("Could not accept channel (%s)", err)
		return
	}
	defer ch.Close()
	for req := range requests {
		s.parseRequest(req, ch)
		return
	}
}

func (s *FakeServer) parseRequest(req *ssh.Request, ch ssh.Channel) {
	if req.Type != "exec" {
		panic(fmt.Errorf("Unsupported request type: %s", req.Type))
	}
	// first 4 bytes is length, ignore it
	cmd := string(req.Payload[4:])
	if cmd != s.Cmd {
		panic(fmt.Errorf("Unknown cmd: %s", cmd))
	}

	if !req.WantReply {
		panic(fmt.Errorf("Expected that want reply is always set"))
	}
	if s.ExecDelay > 0 {
		time.Sleep(s.ExecDelay)
	}
	req.Reply(true, ssh.Marshal(&channelRequestSuccessMsg{}))
	ch.Write([]byte(s.Reply))

	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, uint32(s.ExitStatus))
	ch.SendRequest("exit-status", false, b.Bytes())
	io.Copy(os.Stdout, ch)
	io.Copy(os.Stderr, ch.Stderr())
}
