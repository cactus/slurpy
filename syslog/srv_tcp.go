package syslog

import (
	"bufio"
	"bytes"
	"github.com/cactus/gologit"
	"net"
	"strconv"
	"strings"
)

type SyslogServerTCP struct {
	Running      bool
	ln			 net.Listener
	handler      SyslogMsgHandler
	SyslogCh     chan *SyslogMsg
	Closing      chan bool
}


func (t *SyslogServerTCP) handlerLoop() {
	for {
		select {
		case msg := <-t.SyslogCh:
			t.handler(msg)
		case <-t.Closing:
			return
		}
	}
}

// Reads a "octet counting" framing syslog message, parses, and sends
// the result down the channel
func (t *SyslogServerTCP) readSyslogOctetFraming(bufc *bufio.Reader) {
	for {
		s, err := bufc.ReadString(' ')
		if err != nil {
			gologit.Println(err)
			break
		}
		s = strings.TrimSpace(s)
		readsize, err := strconv.Atoi(s)
		if err != nil {
			gologit.Println(err)
			break
		}

		var result []byte
		for i := 0; i < readsize; i++ {
			b, err := bufc.ReadByte()
			if err != nil {
				gologit.Println(err)
				break
			}
			result = append(result, b)
		}
		if len(result) > 0 {
			m, err := parseSyslogMsg(result)
			if err != nil {
				continue
			}
			t.SyslogCh <- m
		}
	}
}

// Reads a "non-transparent" framing syslog message, parses, and sends
// the result down the channel
func (t *SyslogServerTCP) readSyslogTrailerFraming(bufc *bufio.Reader) {
	for {
		result, err := bufc.ReadBytes('\n')
		if err != nil {
			gologit.Println(err)
			break
		}
		result = bytes.TrimSpace(result)
		if len(result) > 0 {
			m, err := parseSyslogMsg(result)
			if err != nil {
				gologit.Println(err)
				continue
			}
			t.SyslogCh <- m
		}
	}
}

// handles a single conn, reading and parsing syslog messages.
// Syslog messages are sent down the channel.
func (t *SyslogServerTCP) connLoop(c net.Conn) {
	bufc := bufio.NewReader(c)
	defer c.Close()

	pk, err := bufc.Peek(1)
	if err != nil {
		return
	}

	// if first char is `<`, then we are doing "non-transparent" framing
	// (old style). If it is not, then we assume "octet counting" framing.
	// See: http://tools.ietf.org/html/rfc6587
	if bytes.Equal(pk, []byte("<")) {
		t.readSyslogTrailerFraming(bufc)
	} else {
		t.readSyslogOctetFraming(bufc)
	}
}

// Accepts on a listener, creates conns, reads from a tcp connection, parses
// syslog messages, and sends them down supplied channel.
func (t *SyslogServerTCP) acceptLoop() {
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			gologit.Println(err)
			continue
		}
		go t.connLoop(conn)
	}
}

func (t *SyslogServerTCP) Close() {
	if !t.Running {
		return
	}
	t.Running = false
	t.Closing <- true
	t.ln.Close()
}

func (t *SyslogServerTCP) Start() {
	if t.Running {
		return
	}
	t.Running = true
	go t.handlerLoop()
	go t.acceptLoop()
}


func ListenTCP(proto string, inaddr string, handler SyslogMsgHandler) (*SyslogServerTCP, error) {
	proto = strings.ToLower(proto)
	ltcp, err := net.Listen(proto, inaddr)
	if err != nil {
		return nil, err
	}
	sch := make(chan *SyslogMsg, 100)
	clch := make(chan bool)
	srv := &SyslogServerTCP{ln: ltcp, SyslogCh: sch, handler: handler, Closing: clch}
	go srv.Start()
	return srv, nil
}

