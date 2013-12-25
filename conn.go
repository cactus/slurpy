package main

import (
	"bufio"
	"bytes"
	"github.com/cactus/gologit"
	"net"
	"strconv"
	"strings"
)

func udpAcceptor(pc *net.UDPConn, ch chan<- *SyslogMsg) {
	buf := make([]byte, 1024)
	for {
		n, _, err := pc.ReadFromUDP(buf)
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
		ch <- m
	}
}

func readSyslogOctetFraming(bufc *bufio.Reader, ch chan<- *SyslogMsg) {
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
			ch <- m
		}
	}
}

func readSyslogTrailerFraming(bufc *bufio.Reader, ch chan<- *SyslogMsg) {
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
			ch <- m
		}
	}
}

func handleConn(c net.Conn, ch chan<- *SyslogMsg) {
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
		readSyslogTrailerFraming(bufc, ch)
	} else {
		readSyslogOctetFraming(bufc, ch)
	}
}

func tcpAcceptor(ln net.Listener, ch chan<- *SyslogMsg) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			gologit.Println(err)
			continue
		}
		gologit.Println("got conn")
		go handleConn(conn, ch)
	}
}
