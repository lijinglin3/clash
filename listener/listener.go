package listener

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/lijinglin3/clash/adapter/inbound"
	"github.com/lijinglin3/clash/config"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/listener/http"
	"github.com/lijinglin3/clash/listener/mixed"
	"github.com/lijinglin3/clash/listener/redir"
	"github.com/lijinglin3/clash/listener/socks"
	"github.com/lijinglin3/clash/listener/tproxy"
	"github.com/lijinglin3/clash/listener/tunnel"
	"github.com/lijinglin3/clash/log"

	"github.com/samber/lo"
)

var (
	allowLan    = false
	bindAddress = "*"

	tcpListeners = map[constant.Inbound]constant.Listener{}
	udpListeners = map[constant.Inbound]constant.Listener{}

	tunnelTCPListeners = map[string]*tunnel.Listener{}
	tunnelUDPListeners = map[string]*tunnel.PacketConn{}

	// lock for recreate function
	recreateMux sync.Mutex
	tunnelMux   sync.Mutex
)

type Ports struct {
	Port       int `json:"port"`
	SocksPort  int `json:"socks-port"`
	RedirPort  int `json:"redir-port"`
	TProxyPort int `json:"tproxy-port"`
	MixedPort  int `json:"mixed-port"`
}

var tcpListenerCreators = map[constant.InboundType]tcpListenerCreator{
	constant.InboundTypeHTTP:   http.New,
	constant.InboundTypeSocks:  socks.New,
	constant.InboundTypeRedir:  redir.New,
	constant.InboundTypeTproxy: tproxy.New,
	constant.InboundTypeMixed:  mixed.New,
}

var udpListenerCreators = map[constant.InboundType]udpListenerCreator{
	constant.InboundTypeSocks:  socks.NewUDP,
	constant.InboundTypeRedir:  tproxy.NewUDP,
	constant.InboundTypeTproxy: tproxy.NewUDP,
	constant.InboundTypeMixed:  socks.NewUDP,
}

type (
	tcpListenerCreator func(addr string, tcpIn chan<- constant.ConnContext) (constant.Listener, error)
	udpListenerCreator func(addr string, udpIn chan<- *inbound.PacketAdapter) (constant.Listener, error)
)

func AllowLan() bool {
	return allowLan
}

func BindAddress() string {
	return bindAddress
}

func SetAllowLan(al bool) {
	allowLan = al
}

func SetBindAddress(host string) {
	bindAddress = host
}

func createListener(inbound constant.Inbound, tcpIn chan<- constant.ConnContext, udpIn chan<- *inbound.PacketAdapter) {
	addr := inbound.BindAddress
	if portIsZero(addr) {
		return
	}
	tcpCreator := tcpListenerCreators[inbound.Type]
	udpCreator := udpListenerCreators[inbound.Type]
	if tcpCreator == nil && udpCreator == nil {
		log.Errorln("inbound type %s is not supported", inbound.Type)
		return
	}
	if tcpCreator != nil {
		tcpListener, err := tcpCreator(addr, tcpIn)
		if err != nil {
			log.Errorln("create addr %s tcp listener error: %v", addr, err)
			return
		}
		tcpListeners[inbound] = tcpListener
	}
	if udpCreator != nil {
		udpListener, err := udpCreator(addr, udpIn)
		if err != nil {
			log.Errorln("create addr %s udp listener error: %v", addr, err)
			return
		}
		udpListeners[inbound] = udpListener
	}
	log.Infoln("inbound %s created successfully", inbound.ToAlias())
}

func closeListener(inbound constant.Inbound) {
	listener := tcpListeners[inbound]
	if listener != nil {
		if err := listener.Close(); err != nil {
			log.Errorln("close tcp address `%s` error: %s", inbound.ToAlias(), err.Error())
		}
		delete(tcpListeners, inbound)
	}
	listener = udpListeners[inbound]
	if listener != nil {
		if err := listener.Close(); err != nil {
			log.Errorln("close udp address `%s` error: %s", inbound.ToAlias(), err.Error())
		}
		delete(udpListeners, inbound)
	}
}

func getNeedCloseAndCreateInbound(originInbounds, newInbounds []constant.Inbound) ([]constant.Inbound, []constant.Inbound) {
	needCloseMap := map[constant.Inbound]bool{}
	needClose := []constant.Inbound{}
	needCreate := []constant.Inbound{}

	for _, inbound := range originInbounds {
		needCloseMap[inbound] = true
	}
	for _, inbound := range newInbounds {
		if needCloseMap[inbound] {
			delete(needCloseMap, inbound)
		} else {
			needCreate = append(needCreate, inbound)
		}
	}
	for inbound := range needCloseMap {
		needClose = append(needClose, inbound)
	}
	return needClose, needCreate
}

// only recreate inbound config listener
func ReCreateListeners(inbounds []constant.Inbound, tcpIn chan<- constant.ConnContext, udpIn chan<- *inbound.PacketAdapter) {
	newInbounds := []constant.Inbound{}
	newInbounds = append(newInbounds, inbounds...)
	for _, inbound := range getInbounds() {
		if inbound.IsFromPortCfg {
			newInbounds = append(newInbounds, inbound)
		}
	}
	reCreateListeners(newInbounds, tcpIn, udpIn)
}

// only recreate ports config listener
func ReCreatePortsListeners(ports Ports, tcpIn chan<- constant.ConnContext, udpIn chan<- *inbound.PacketAdapter) {
	newInbounds := []constant.Inbound{}
	newInbounds = append(newInbounds, GetInbounds()...)
	newInbounds = addPortInbound(newInbounds, constant.InboundTypeHTTP, ports.Port)
	newInbounds = addPortInbound(newInbounds, constant.InboundTypeSocks, ports.SocksPort)
	newInbounds = addPortInbound(newInbounds, constant.InboundTypeRedir, ports.RedirPort)
	newInbounds = addPortInbound(newInbounds, constant.InboundTypeTproxy, ports.TProxyPort)
	newInbounds = addPortInbound(newInbounds, constant.InboundTypeMixed, ports.MixedPort)
	reCreateListeners(newInbounds, tcpIn, udpIn)
}

func addPortInbound(inbounds []constant.Inbound, inboundType constant.InboundType, port int) []constant.Inbound {
	if port != 0 {
		inbounds = append(inbounds, constant.Inbound{
			Type:          inboundType,
			BindAddress:   genAddr(bindAddress, port, allowLan),
			IsFromPortCfg: true,
		})
	}
	return inbounds
}

func reCreateListeners(inbounds []constant.Inbound, tcpIn chan<- constant.ConnContext, udpIn chan<- *inbound.PacketAdapter) {
	recreateMux.Lock()
	defer recreateMux.Unlock()
	needClose, needCreate := getNeedCloseAndCreateInbound(getInbounds(), inbounds)
	for _, inbound := range needClose {
		closeListener(inbound)
	}
	for _, inbound := range needCreate {
		createListener(inbound, tcpIn, udpIn)
	}
}

func PatchTunnel(tunnels []config.Tunnel, tcpIn chan<- constant.ConnContext, udpIn chan<- *inbound.PacketAdapter) {
	tunnelMux.Lock()
	defer tunnelMux.Unlock()

	type addrProxy struct {
		network string
		addr    string
		target  string
		proxy   string
	}

	tcpOld := lo.Map(
		lo.Keys(tunnelTCPListeners),
		func(key string, _ int) addrProxy {
			parts := strings.Split(key, "/")
			return addrProxy{
				network: "tcp",
				addr:    parts[0],
				target:  parts[1],
				proxy:   parts[2],
			}
		},
	)
	udpOld := lo.Map(
		lo.Keys(tunnelUDPListeners),
		func(key string, _ int) addrProxy {
			parts := strings.Split(key, "/")
			return addrProxy{
				network: "udp",
				addr:    parts[0],
				target:  parts[1],
				proxy:   parts[2],
			}
		},
	)
	oldElm := lo.Union(tcpOld, udpOld)

	newElm := lo.FlatMap(
		tunnels,
		func(tunnel config.Tunnel, _ int) []addrProxy {
			return lo.Map(
				tunnel.Network,
				func(network string, _ int) addrProxy {
					return addrProxy{
						network: network,
						addr:    tunnel.Address,
						target:  tunnel.Target,
						proxy:   tunnel.Proxy,
					}
				},
			)
		},
	)

	needClose, needCreate := lo.Difference(oldElm, newElm)

	for _, elm := range needClose {
		key := fmt.Sprintf("%s/%s/%s", elm.addr, elm.target, elm.proxy)
		if elm.network == "tcp" {
			tunnelTCPListeners[key].Close()
			delete(tunnelTCPListeners, key)
		} else {
			tunnelUDPListeners[key].Close()
			delete(tunnelUDPListeners, key)
		}
	}

	for _, elm := range needCreate {
		key := fmt.Sprintf("%s/%s/%s", elm.addr, elm.target, elm.proxy)
		if elm.network == "tcp" {
			l, err := tunnel.New(elm.addr, elm.target, elm.proxy, tcpIn)
			if err != nil {
				log.Errorln("Start tunnel %s error: %s", elm.target, err.Error())
				continue
			}
			tunnelTCPListeners[key] = l
			log.Infoln("Tunnel(tcp/%s) proxy %s listening at: %s", elm.target, elm.proxy, tunnelTCPListeners[key].Address())
		} else {
			l, err := tunnel.NewUDP(elm.addr, elm.target, elm.proxy, udpIn)
			if err != nil {
				log.Errorln("Start tunnel %s error: %s", elm.target, err.Error())
				continue
			}
			tunnelUDPListeners[key] = l
			log.Infoln("Tunnel(udp/%s) proxy %s listening at: %s", elm.target, elm.proxy, tunnelUDPListeners[key].Address())
		}
	}
}

func GetInbounds() []constant.Inbound {
	return lo.Filter(getInbounds(), func(inbound constant.Inbound, idx int) bool {
		return !inbound.IsFromPortCfg
	})
}

// GetInbounds return the inbounds of proxy servers
func getInbounds() []constant.Inbound {
	var inbounds []constant.Inbound
	for inbound := range tcpListeners {
		inbounds = append(inbounds, inbound)
	}
	for inbound := range udpListeners {
		if _, ok := tcpListeners[inbound]; !ok {
			inbounds = append(inbounds, inbound)
		}
	}
	return inbounds
}

// GetPorts return the ports of proxy servers
func GetPorts() *Ports {
	ports := &Ports{}
	for _, inbound := range getInbounds() {
		fillPort(inbound, ports)
	}
	return ports
}

func fillPort(inbound constant.Inbound, ports *Ports) {
	if inbound.IsFromPortCfg {
		port := getPort(inbound.BindAddress)
		switch inbound.Type {
		case constant.InboundTypeHTTP:
			ports.Port = port
		case constant.InboundTypeSocks:
			ports.SocksPort = port
		case constant.InboundTypeTproxy:
			ports.TProxyPort = port
		case constant.InboundTypeRedir:
			ports.RedirPort = port
		case constant.InboundTypeMixed:
			ports.MixedPort = port
		default:
			// do nothing
		}
	}
}

func portIsZero(addr string) bool {
	_, port, err := net.SplitHostPort(addr)
	if port == "0" || port == "" || err != nil {
		return true
	}
	return false
}

func genAddr(host string, port int, allowLan bool) string {
	if allowLan {
		if host == "*" {
			return fmt.Sprintf(":%d", port)
		}
		return fmt.Sprintf("%s:%d", host, port)
	}

	return fmt.Sprintf("127.0.0.1:%d", port)
}

func getPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}
