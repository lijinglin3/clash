package redir

import (
	"net"

	"github.com/lijinglin3/clash/adapter/inbound"
	"github.com/lijinglin3/clash/constant"
)

type Listener struct {
	listener net.Listener
	addr     string
	closed   bool
}

// RawAddress implements constant.Listener
func (l *Listener) RawAddress() string {
	return l.addr
}

// Address implements constant.Listener
func (l *Listener) Address() string {
	return l.listener.Addr().String()
}

// Close implements constant.Listener
func (l *Listener) Close() error {
	l.closed = true
	return l.listener.Close()
}

func New(addr string, in chan<- constant.ConnContext) (constant.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	rl := &Listener{
		listener: l,
		addr:     addr,
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				if rl.closed {
					break
				}
				continue
			}
			go handleRedir(c, in)
		}
	}()

	return rl, nil
}

func handleRedir(conn net.Conn, in chan<- constant.ConnContext) {
	target, err := parserPacket(conn)
	if err != nil {
		conn.Close()
		return
	}
	conn.(*net.TCPConn).SetKeepAlive(true)
	in <- inbound.NewSocket(target, conn, constant.REDIR)
}
