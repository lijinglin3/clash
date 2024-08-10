package rule

import (
	"strings"

	"github.com/lijinglin3/clash/constant"
)

// Implements constant.Rule
var _ constant.Rule = (*DomainSuffix)(nil)

type DomainSuffix struct {
	suffix  string
	adapter string
}

func (ds *DomainSuffix) RuleType() constant.RuleType {
	return constant.DomainSuffix
}

func (ds *DomainSuffix) Match(metadata *constant.Metadata) bool {
	domain := metadata.Host
	return strings.HasSuffix(domain, "."+ds.suffix) || domain == ds.suffix
}

func (ds *DomainSuffix) Adapter() string {
	return ds.adapter
}

func (ds *DomainSuffix) Payload() string {
	return ds.suffix
}

func (ds *DomainSuffix) ShouldResolveIP() bool {
	return false
}

func (ds *DomainSuffix) ShouldFindProcess() bool {
	return false
}

func NewDomainSuffix(suffix, adapter string) *DomainSuffix {
	return &DomainSuffix{
		suffix:  strings.ToLower(suffix),
		adapter: adapter,
	}
}
