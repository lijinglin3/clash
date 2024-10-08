package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"net"

	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
)

type Base struct {
	name  string
	addr  string
	iface string
	tp    constant.AdapterType
	udp   bool
	rmark int
}

// Name implements constant.ProxyAdapter
func (b *Base) Name() string {
	return b.name
}

// Type implements constant.ProxyAdapter
func (b *Base) Type() constant.AdapterType {
	return b.tp
}

// StreamConn implements constant.ProxyAdapter
func (b *Base) StreamConn(c net.Conn, metadata *constant.Metadata) (net.Conn, error) {
	return c, errors.New("no support")
}

// ListenPacketContext implements constant.ProxyAdapter
func (b *Base) ListenPacketContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.PacketConn, error) {
	return nil, errors.New("no support")
}

// SupportUDP implements constant.ProxyAdapter
func (b *Base) SupportUDP() bool {
	return b.udp
}

// MarshalJSON implements constant.ProxyAdapter
func (b *Base) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": b.Type().String(),
	})
}

// Addr implements constant.ProxyAdapter
func (b *Base) Addr() string {
	return b.addr
}

// Unwrap implements constant.ProxyAdapter
func (b *Base) Unwrap(metadata *constant.Metadata) constant.Proxy {
	return nil
}

// DialOptions return []dialer.Option from struct
func (b *Base) DialOptions(opts ...dialer.Option) []dialer.Option {
	if b.iface != "" {
		opts = append(opts, dialer.WithInterface(b.iface))
	}

	if b.rmark != 0 {
		opts = append(opts, dialer.WithRoutingMark(b.rmark))
	}

	return opts
}

type BasicOption struct {
	Interface   string `group:"interface-name,omitempty" proxy:"interface-name,omitempty"`
	RoutingMark int    `group:"routing-mark,omitempty"   proxy:"routing-mark,omitempty"`
}

type BaseOption struct {
	Name        string
	Addr        string
	Type        constant.AdapterType
	UDP         bool
	Interface   string
	RoutingMark int
}

func NewBase(opt BaseOption) *Base {
	return &Base{
		name:  opt.Name,
		addr:  opt.Addr,
		tp:    opt.Type,
		udp:   opt.UDP,
		iface: opt.Interface,
		rmark: opt.RoutingMark,
	}
}

type conn struct {
	net.Conn
	chain constant.Chain
}

// Chains implements constant.Connection
func (c *conn) Chains() constant.Chain {
	return c.chain
}

// AppendToChains implements constant.Connection
func (c *conn) AppendToChains(a constant.ProxyAdapter) {
	c.chain = append(c.chain, a.Name())
}

func NewConn(c net.Conn, a constant.ProxyAdapter) constant.Conn {
	return &conn{c, []string{a.Name()}}
}

type packetConn struct {
	net.PacketConn
	chain constant.Chain
}

// Chains implements constant.Connection
func (c *packetConn) Chains() constant.Chain {
	return c.chain
}

// AppendToChains implements constant.Connection
func (c *packetConn) AppendToChains(a constant.ProxyAdapter) {
	c.chain = append(c.chain, a.Name())
}

func newPacketConn(pc net.PacketConn, a constant.ProxyAdapter) constant.PacketConn {
	return &packetConn{pc, []string{a.Name()}}
}
