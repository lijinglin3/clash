package tunnel

import (
	"errors"
	"net"
	"net/netip"
	"time"

	cnet "github.com/lijinglin3/clash/common/net"
	"github.com/lijinglin3/clash/common/pool"
	"github.com/lijinglin3/clash/constant"
)

func handleUDPToRemote(packet constant.UDPPacket, pc constant.PacketConn, metadata *constant.Metadata) error {
	addr := metadata.UDPAddr()
	if addr == nil {
		return errors.New("udp addr invalid")
	}

	if _, err := pc.WriteTo(packet.Data(), addr); err != nil {
		return err
	}
	// reset timeout
	pc.SetReadDeadline(time.Now().Add(udpTimeout))

	return nil
}

func handleUDPToLocal(packet constant.UDPPacket, pc net.PacketConn, key string, oAddr, fAddr netip.Addr) {
	buf := pool.Get(pool.UDPBufferSize)
	defer pool.Put(buf)
	defer natTable.Delete(key)
	defer pc.Close()

	for {
		pc.SetReadDeadline(time.Now().Add(udpTimeout))
		n, from, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}

		fromUDPAddr := *from.(*net.UDPAddr)
		if fAddr.IsValid() {
			fromAddr, _ := netip.AddrFromSlice(fromUDPAddr.IP)
			fromAddr = fromAddr.Unmap()
			if oAddr == fromAddr {
				fromUDPAddr.IP = fAddr.AsSlice()
			}
		}

		_, err = packet.WriteBack(buf[:n], &fromUDPAddr)
		if err != nil {
			return
		}
	}
}

func handleSocket(ctx constant.ConnContext, outbound net.Conn) {
	cnet.Relay(ctx.Conn(), outbound)
}
