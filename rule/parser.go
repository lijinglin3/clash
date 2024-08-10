package rule

import (
	"fmt"

	"github.com/lijinglin3/clash/constant"
)

func ParseRule(tp, payload, target string, params []string) (constant.Rule, error) {
	var (
		parseErr error
		parsed   constant.Rule
	)

	ruleConfigType := constant.RuleConfig(tp)

	switch ruleConfigType {
	case constant.RuleConfigDomain:
		parsed = NewDomain(payload, target)
	case constant.RuleConfigDomainSuffix:
		parsed = NewDomainSuffix(payload, target)
	case constant.RuleConfigDomainKeyword:
		parsed = NewDomainKeyword(payload, target)
	case constant.RuleConfigGeoIP:
		noResolve := HasNoResolve(params)
		parsed = NewGEOIP(payload, target, noResolve)
	case constant.RuleConfigIPCIDR, constant.RuleConfigIPCIDR6:
		noResolve := HasNoResolve(params)
		parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRNoResolve(noResolve))
	case constant.RuleConfigSrcIPCIDR:
		parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRSourceIP(true), WithIPCIDRNoResolve(true))
	case constant.RuleConfigSrcPort:
		parsed, parseErr = NewPort(payload, target, PortTypeSrc)
	case constant.RuleConfigDstPort:
		parsed, parseErr = NewPort(payload, target, PortTypeDest)
	case constant.RuleConfigInboundPort:
		parsed, parseErr = NewPort(payload, target, PortTypeInbound)
	case constant.RuleConfigProcessName:
		parsed, parseErr = NewProcess(payload, target, true)
	case constant.RuleConfigProcessPath:
		parsed, parseErr = NewProcess(payload, target, false)
	case constant.RuleConfigIPSet:
		noResolve := HasNoResolve(params)
		parsed, parseErr = NewIPSet(payload, target, noResolve)
	case constant.RuleConfigMatch:
		parsed = NewMatch(target)
	case constant.RuleConfigRuleSet, constant.RuleConfigScript:
		parseErr = fmt.Errorf("unsupported rule type %s", tp)
	default:
		parseErr = fmt.Errorf("unsupported rule type %s", tp)
	}

	return parsed, parseErr
}
