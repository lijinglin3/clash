package rule

import (
	"fmt"
	"strconv"

	"github.com/lijinglin3/clash/constant"
)

type PortType int

const (
	PortTypeSrc PortType = iota
	PortTypeDest
	PortTypeInbound
)

// Implements constant.Rule
var _ constant.Rule = (*Port)(nil)

type Port struct {
	adapter  string
	port     constant.Port
	portType PortType
}

func (p *Port) RuleType() constant.RuleType {
	switch p.portType {
	case PortTypeSrc:
		return constant.SrcPort
	case PortTypeDest:
		return constant.DstPort
	case PortTypeInbound:
		return constant.InboundPort
	default:
		panic(fmt.Errorf("unknown port type: %v", p.portType))
	}
}

func (p *Port) Match(metadata *constant.Metadata) bool {
	switch p.portType {
	case PortTypeSrc:
		return metadata.SrcPort == p.port
	case PortTypeDest:
		return metadata.DstPort == p.port
	case PortTypeInbound:
		return metadata.OriginDst.Port() == uint16(p.port)
	default:
		panic(fmt.Errorf("unknown port type: %v", p.portType))
	}
}

func (p *Port) Adapter() string {
	return p.adapter
}

func (p *Port) Payload() string {
	return p.port.String()
}

func (p *Port) ShouldResolveIP() bool {
	return false
}

func (p *Port) ShouldFindProcess() bool {
	return false
}

func NewPort(port, adapter string, portType PortType) (*Port, error) {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, errPayload
	}
	return &Port{
		adapter:  adapter,
		port:     constant.Port(p),
		portType: portType,
	}, nil
}
