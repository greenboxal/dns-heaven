package dnsheaven

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// ResolvError type
type ResolvError struct {
	qname, net  string
	nameservers []string
}

// Error formats a ResolvError
func (e ResolvError) Error() string {
	errmsg := fmt.Sprintf("%s resolv failed on %s (%s)", e.qname, strings.Join(e.nameservers, "; "), e.net)
	return errmsg
}

// Resolver type
type StandardResolver struct {
	Nameservers []string
	Timeout     time.Duration
	Interval    time.Duration
}

// Lookup will ask each nameserver in top-to-bottom fashion, starting a new request
// in every second, and return as early as possbile (have an answer).
// It returns an error if no request has succeeded.
func (r *StandardResolver) Lookup(net string, req *dns.Msg) (message *dns.Msg, err error) {
	var wg sync.WaitGroup

	c := &dns.Client{
		Net:          net,
		ReadTimeout:  r.Timeout,
		WriteTimeout: r.Timeout,
	}

	qname := req.Question[0].Name
	res := make(chan *dns.Msg, 1)

	L := func(nameserver string) {
		defer wg.Done()

		r, _, err := c.Exchange(req, nameserver)

		if err != nil {
			logrus.WithError(err).WithField("qname", qname).WithField("ns", nameserver).Error("error resolving query")
			return
		}

		if r != nil && r.Rcode != dns.RcodeSuccess {
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		}

		select {
		case res <- r:
		default:
		}
	}

	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()

	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range r.Nameservers {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case r := <-res:
			return r, nil
		case <-ticker.C:
			continue
		}
	}

	// wait for all the namservers to finish
	wg.Wait()
	select {
	case r := <-res:
		return r, nil
	default:
		return nil, ResolvError{qname, net, r.Nameservers}
	}
}
