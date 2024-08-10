package context

import (
	"net"

	"github.com/lijinglin3/clash/constant"

	"github.com/gofrs/uuid/v5"
)

type ConnContext struct {
	id       uuid.UUID
	metadata *constant.Metadata
	conn     net.Conn
}

func NewConnContext(conn net.Conn, metadata *constant.Metadata) *ConnContext {
	id, _ := uuid.NewV4()
	return &ConnContext{
		id:       id,
		metadata: metadata,
		conn:     conn,
	}
}

// ID implement constant.ConnContext ID
func (c *ConnContext) ID() uuid.UUID {
	return c.id
}

// Metadata implement constant.ConnContext Metadata
func (c *ConnContext) Metadata() *constant.Metadata {
	return c.metadata
}

// Conn implement constant.ConnContext Conn
func (c *ConnContext) Conn() net.Conn {
	return c.conn
}
