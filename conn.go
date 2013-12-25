package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func udpAcceptor(pc *net.UDPConn, ch chan<- *SyslogMsg) {
	buf := make([]byte, 1024)
	for {
		n, _, err := pc.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf("read %d\n", n)
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

func handleConn(c net.Conn, ch chan<- *SyslogMsg) {
	bufc := bufio.NewReader(c)
	defer c.Close()

	pk, err := bufc.Peek(1)
	if err != nil {
		return
	}
	if !bytes.Equal(pk, []byte("<")) {
		for {
			s, err := bufc.ReadString(' ')
			if err != nil {
				log.Println(err)
				break
			}
			s = strings.TrimSpace(s)
			readsize, err := strconv.Atoi(s)
			if err != nil {
				log.Println(err)
				break
			}

			var result []byte
			for i := 0; i < readsize; i++ {
				b, err := bufc.ReadByte()
				if err != nil {
					log.Println(err)
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
	} else {
		for {
			result, err := bufc.ReadBytes('\n')
			if err != nil {
				log.Println(err)
				break
			}
			result = bytes.TrimSpace(result)
			if len(result) > 0 {
				m, err := parseSyslogMsg(result)
				if err != nil {
					log.Println(err)
					continue
				}
				ch <- m
			}
		}
	}
}

func tcpAcceptor(ln net.Listener, ch chan<- *SyslogMsg) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("got conn")
		go handleConn(conn, ch)
	}
}
