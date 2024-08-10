package context

import (
	"net"

	"github.com/lijinglin3/clash/constant"

	"github.com/gofrs/uuid/v5"
)

type PacketConnContext struct {
	id         uuid.UUID
	metadata   *constant.Metadata
	packetConn net.PacketConn
}

func NewPacketConnContext(metadata *constant.Metadata) *PacketConnContext {
	id, _ := uuid.NewV4()
	return &PacketConnContext{
		id:       id,
		metadata: metadata,
	}
}

// ID implement constant.PacketConnContext ID
func (pc *PacketConnContext) ID() uuid.UUID {
	return pc.id
}

// Metadata implement constant.PacketConnContext Metadata
func (pc *PacketConnContext) Metadata() *constant.Metadata {
	return pc.metadata
}

// PacketConn implement constant.PacketConnContext PacketConn
func (pc *PacketConnContext) PacketConn() net.PacketConn {
	return pc.packetConn
}

// InjectPacketConn injectPacketConn manually
func (pc *PacketConnContext) InjectPacketConn(pconn constant.PacketConn) {
	pc.packetConn = pconn
}
