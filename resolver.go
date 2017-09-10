package dnsheaven

import (
	"github.com/miekg/dns"
)

type Resolver interface {
	Resolve(net string, req *dns.Msg) (*dns.Msg, error)
}
