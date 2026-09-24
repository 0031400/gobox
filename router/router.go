package router

import "gobox/common"

type Router struct {
	final string
	rules []RouteRule
}

func NewRouter(final string, rules []RouteRule) *Router {
	return &Router{final: final, rules: rules}
}
func (r *Router) Route(addr common.TargetAddr) string {
	for _, rule := range r.rules {
		if rule.Match(addr) {
			return rule.Outbound
		}
	}
	return r.final
}
