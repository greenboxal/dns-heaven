package osx

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/greenboxal/dns-heaven"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type Resolver struct {
	udp *dns.Client
	tcp *dns.Client

	config   *dnsheaven.Config
	dns      *DnsConfig
	domains  map[string]*dnsheaven.StandardResolver
	standard *dnsheaven.StandardResolver
}

func New(config *dnsheaven.Config) (*Resolver, error) {
	udp := &dns.Client{
		Net: "udp",
	}

	tcp := &dns.Client{
		Net: "tcp",
	}

	r := &Resolver{
		udp: udp,
		tcp: tcp,

		config:  config,
		domains: make(map[string]*dnsheaven.StandardResolver),
	}

	err := r.refresh()

	if err != nil {
		return nil, err
	}

	go r.run()

	return r, nil
}

func (r *Resolver) Resolve(net string, msg *dns.Msg) (*dns.Msg, error) {
	resolver := r.resolverForRequest(msg)

	return resolver.Lookup(net, msg)
}

func (r *Resolver) run() {
	timer := time.Tick(1 * time.Second)

	for _ = range timer {
		err := r.refresh()

		if err != nil {
			logrus.WithError(err).Error("error refreshing dns config")
			continue
		}

		err = r.hijack()

		if err != nil {
			logrus.WithError(err).Error("error hijacking dns config")
			continue
		}
	}
}

func (r *Resolver) refresh() error {
	cmd := exec.Command("/usr/sbin/scutil", "--dns")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	parsed, err := ParseScutilDns(string(output))

	if err != nil {
		return err
	}

	return r.update(parsed)
}

func (r *Resolver) update(d *DnsInfo) error {
	var standard *ResolverInfo
	var domains []*ResolverInfo

	for _, re := range d.Config.Resolvers {
		if !re.Reachable {
			continue
		}

		if re.IsMdns {
			continue
		}

		if standard == nil && len(re.Domain) == 0 {
			standard = re
		} else if len(re.Domain) > 0 {
			domains = append(domains, re)
		}
	}

	if standard == nil {
		standard = &ResolverInfo{
			Nameservers: []string{
				"8.8.8.8:53",
				"8.8.4.4:53",
			},
		}
	}

	perDomain := map[string]*dnsheaven.StandardResolver{}
	for _, d := range domains {
		perDomain[d.Domain] = r.buildResolver(d)
	}

	r.standard = r.buildResolver(standard)
	r.domains = perDomain

	return nil
}

func (r *Resolver) buildResolver(re *ResolverInfo) *dnsheaven.StandardResolver {
	var timeout time.Duration

	if re.Timeout != 0 {
		timeout = time.Duration(re.Timeout) * time.Second
	} else {
		timeout = time.Duration(r.config.Timeout) * time.Millisecond
	}

	return &dnsheaven.StandardResolver{
		Nameservers: re.Nameservers,
		Interval:    time.Duration(r.config.Interval) * time.Millisecond,
		Timeout:     timeout,
	}
}

func (r *Resolver) hijack() error {
	host, _, err := net.SplitHostPort(r.config.Address)

	if err != nil {
		return err
	}

	// FIXME: This assumes that we're listening on port 53
	content := fmt.Sprintf("nameserver %s", host)

	err = ioutil.WriteFile("/etc/resolv.conf", []byte(content), 0644)

	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) resolverForRequest(msg *dns.Msg) *dnsheaven.StandardResolver {
	if msg.Opcode != dns.OpcodeQuery && msg.Opcode != dns.OpcodeIQuery {
		return r.standard
	}

	qname := msg.Question[0].Name

	// Try to find a domain match
	// FIXME: Maybe this should be the longest match
	for k, v := range r.domains {
		if strings.HasSuffix(strings.ToLower(qname), strings.ToLower(k)+".") {
			return v
		}
	}

	return r.standard
}
