package slurpylog

import (
	"github.com/cactus/gologit"
	"net"
	"strings"
)

type SyslogServerUDP struct {
	Running      bool
	pc			 *net.UDPConn
	handler      SyslogMsgHandler
	SyslogCh     chan *SyslogMsg
	Closing      chan bool
}

// Reads from a udp socket, parses syslog messages, and sends them
// down supplied channel.
func (u *SyslogServerUDP) acceptLoop() {
	buf := make([]byte, 1024)
	for {
		n, _, err := u.pc.ReadFromUDP(buf)
		if err != nil {
			gologit.Println(err)
			continue
		}
		if n == 0 {
			continue
		}
		m, err := parseSyslogMsg(buf[0:n])
		if err != nil {
			continue
		}
		u.SyslogCh <- m
	}
}

func (u *SyslogServerUDP) handlerLoop() {
	for {
		select {
		case msg := <-u.SyslogCh:
			u.handler(msg)
		case <-u.Closing:
			return
		}
	}
}

func (u *SyslogServerUDP) Close() {
	if !u.Running {
		return
	}
	u.Running = false
	u.Closing <- true
	u.pc.Close()
}

func (u *SyslogServerUDP) Start() {
	if u.Running {
		return
	}
	u.Running = true
	go u.handlerLoop()
	go u.acceptLoop()
}

func ListenUDP(proto string, inaddr string, handler SyslogMsgHandler) (*SyslogServerUDP, error) {
	proto = strings.ToLower(proto)
	addr, err := net.ResolveUDPAddr(proto, inaddr)
	if err != nil {
		return nil, err
	}
	ludp, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	sch := make(chan *SyslogMsg, 100)
	clch := make(chan bool)
	srv := &SyslogServerUDP{pc: ludp, SyslogCh: sch, handler: handler, Closing: clch}
	go srv.Start()
	return srv, nil
}
