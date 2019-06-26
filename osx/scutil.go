package osx

import (
	"fmt"
	"strconv"
	"strings"
)

type DnsInfo struct {
	Config *DnsConfig
	Scoped *DnsConfig
}

type DnsConfig struct {
	Resolvers []*ResolverInfo
}

type ResolverInfo struct {
	SearchDomains []string
	Nameservers   []string
	Domain        string
	Reachable     bool
	Timeout       int
	IsMdns        bool
}

// FIXME: This parser is pretty lame and probably will break if anything changes
func ParseScutilDns(data string) (*DnsInfo, error) {
	var currentConfig *DnsConfig
	var currentResolver *ResolverInfo

	lines := strings.Split(data, "\n")

	info := &DnsInfo{
		Config: &DnsConfig{},
		Scoped: &DnsConfig{},
	}

	for _, l := range lines {
		if l == "DNS configuration" {
			currentConfig = info.Config
			continue
		} else if l == "DNS configuration (for scoped queries)" {
			currentConfig = info.Scoped
			continue
		} else if strings.HasPrefix(l, "resolver #") {
			if currentConfig == nil {
				continue
			}

			currentResolver = &ResolverInfo{}
			currentConfig.Resolvers = append(currentConfig.Resolvers, currentResolver)
		} else if strings.HasPrefix(l, "  ") || strings.HasPrefix(l, "\t") {
			if currentResolver == nil {
				continue
			}

			parts := strings.SplitN(l, ":", 2)

			if len(parts) != 2 {
				continue
			}

			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if strings.HasPrefix(name, "search domain") {
				currentResolver.SearchDomains = append(currentResolver.SearchDomains, value)
			} else if strings.HasPrefix(name, "nameserver") {
				currentResolver.Nameservers = append(currentResolver.Nameservers, fmt.Sprintf("[%s]:53", value))
			} else if name == "reach" {
				currentResolver.Reachable = !strings.Contains(value, "Not Reachable")
			} else if name == "domain" {
				currentResolver.Domain = value
			} else if name == "timeout" {
				timeout, err := strconv.Atoi(value)

				if err != nil {
					continue
				}

				currentResolver.Timeout = timeout
			} else if name == "options" {
				currentResolver.IsMdns = strings.Contains(value, "mdns")
			}
		}
	}

	return info, nil
}
