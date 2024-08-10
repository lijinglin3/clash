package statistic

import (
	"net"
	"time"

	"github.com/lijinglin3/clash/constant"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/atomic"
)

type tracker interface {
	ID() string
	Close() error
}

type trackerInfo struct {
	UUID          uuid.UUID          `json:"id"`
	Metadata      *constant.Metadata `json:"metadata"`
	UploadTotal   *atomic.Int64      `json:"upload"`
	DownloadTotal *atomic.Int64      `json:"download"`
	Start         time.Time          `json:"start"`
	Chain         constant.Chain     `json:"chains"`
	Rule          string             `json:"rule"`
	RulePayload   string             `json:"rulePayload"`
}

type TCPTracker struct {
	constant.Conn `json:"-"`
	*trackerInfo
	manager *Manager
}

func (tt *TCPTracker) ID() string {
	return tt.UUID.String()
}

func (tt *TCPTracker) Read(b []byte) (int, error) {
	n, err := tt.Conn.Read(b)
	download := int64(n)
	tt.manager.PushDownloaded(download)
	tt.DownloadTotal.Add(download)
	return n, err
}

func (tt *TCPTracker) Write(b []byte) (int, error) {
	n, err := tt.Conn.Write(b)
	upload := int64(n)
	tt.manager.PushUploaded(upload)
	tt.UploadTotal.Add(upload)
	return n, err
}

func (tt *TCPTracker) Close() error {
	tt.manager.Leave(tt)
	return tt.Conn.Close()
}

func NewTCPTracker(conn constant.Conn, manager *Manager, metadata *constant.Metadata, rule constant.Rule) *TCPTracker {
	uuid, _ := uuid.NewV4()

	t := &TCPTracker{
		Conn:    conn,
		manager: manager,
		trackerInfo: &trackerInfo{
			UUID:          uuid,
			Start:         time.Now(),
			Metadata:      metadata,
			Chain:         conn.Chains(),
			Rule:          "",
			UploadTotal:   atomic.NewInt64(0),
			DownloadTotal: atomic.NewInt64(0),
		},
	}

	if rule != nil {
		t.trackerInfo.Rule = rule.RuleType().String()
		t.trackerInfo.RulePayload = rule.Payload()
	}

	manager.Join(t)
	return t
}

type UDPTracker struct {
	constant.PacketConn `json:"-"`
	*trackerInfo
	manager *Manager
}

func (ut *UDPTracker) ID() string {
	return ut.UUID.String()
}

func (ut *UDPTracker) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := ut.PacketConn.ReadFrom(b)
	download := int64(n)
	ut.manager.PushDownloaded(download)
	ut.DownloadTotal.Add(download)
	return n, addr, err
}

func (ut *UDPTracker) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := ut.PacketConn.WriteTo(b, addr)
	upload := int64(n)
	ut.manager.PushUploaded(upload)
	ut.UploadTotal.Add(upload)
	return n, err
}

func (ut *UDPTracker) Close() error {
	ut.manager.Leave(ut)
	return ut.PacketConn.Close()
}

func NewUDPTracker(conn constant.PacketConn, manager *Manager, metadata *constant.Metadata, rule constant.Rule) *UDPTracker {
	uuid, _ := uuid.NewV4()

	ut := &UDPTracker{
		PacketConn: conn,
		manager:    manager,
		trackerInfo: &trackerInfo{
			UUID:          uuid,
			Start:         time.Now(),
			Metadata:      metadata,
			Chain:         conn.Chains(),
			Rule:          "",
			UploadTotal:   atomic.NewInt64(0),
			DownloadTotal: atomic.NewInt64(0),
		},
	}

	if rule != nil {
		ut.trackerInfo.Rule = rule.RuleType().String()
		ut.trackerInfo.RulePayload = rule.Payload()
	}

	manager.Join(ut)
	return ut
}
