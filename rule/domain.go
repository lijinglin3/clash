package rule

import (
	"strings"

	"github.com/lijinglin3/clash/constant"
)

// Implements constant.Rule
var _ constant.Rule = (*Domain)(nil)

type Domain struct {
	domain  string
	adapter string
}

func (d *Domain) RuleType() constant.RuleType {
	return constant.Domain
}

func (d *Domain) Match(metadata *constant.Metadata) bool {
	return metadata.Host == d.domain
}

func (d *Domain) Adapter() string {
	return d.adapter
}

func (d *Domain) Payload() string {
	return d.domain
}

func (d *Domain) ShouldResolveIP() bool {
	return false
}

func (d *Domain) ShouldFindProcess() bool {
	return false
}

func NewDomain(domain, adapter string) *Domain {
	return &Domain{
		domain:  strings.ToLower(domain),
		adapter: adapter,
	}
}
