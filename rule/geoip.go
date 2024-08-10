package rule

import (
	"strings"

	"github.com/lijinglin3/clash/component/mmdb"
	"github.com/lijinglin3/clash/constant"
)

// Implements constant.Rule
var _ constant.Rule = (*GEOIP)(nil)

type GEOIP struct {
	country     string
	adapter     string
	noResolveIP bool
}

func (g *GEOIP) RuleType() constant.RuleType {
	return constant.GEOIP
}

func (g *GEOIP) Match(metadata *constant.Metadata) bool {
	ip := metadata.DstIP
	if ip == nil {
		return false
	}

	if strings.EqualFold(g.country, "LAN") {
		return ip.IsPrivate()
	}
	record, _ := mmdb.Instance().Country(ip)
	return strings.EqualFold(record.Country.IsoCode, g.country)
}

func (g *GEOIP) Adapter() string {
	return g.adapter
}

func (g *GEOIP) Payload() string {
	return g.country
}

func (g *GEOIP) ShouldResolveIP() bool {
	return !g.noResolveIP
}

func (g *GEOIP) ShouldFindProcess() bool {
	return false
}

func NewGEOIP(country, adapter string, noResolveIP bool) *GEOIP {
	geoip := &GEOIP{
		country:     country,
		adapter:     adapter,
		noResolveIP: noResolveIP,
	}

	return geoip
}
