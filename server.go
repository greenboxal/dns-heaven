package dnsheaven

import (
	"sync"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config *Config
	tcp    *dns.Server
	udp    *dns.Server
}

func NewServer(config *Config, resolver Resolver) *Server {
	resolve := func(net string) dns.HandlerFunc {
		return func(r dns.ResponseWriter, msg *dns.Msg) {
			result, err := resolver.Resolve(net, msg)

			if err != nil {
				logrus.WithError(err).WithField("req", msg).Error("error resolving request")

				r.WriteMsg(&dns.Msg{
					MsgHdr: dns.MsgHdr{
						Id:                 msg.Id,
						Response:           true,
						Opcode:             msg.Opcode,
						Authoritative:      false,
						Truncated:          false,
						RecursionDesired:   false,
						RecursionAvailable: false,
						Zero:               false,
						AuthenticatedData:  false,
						CheckingDisabled:   false,
						Rcode:              dns.RcodeServerFailure,
					},
				})

				return
			}

			r.WriteMsg(result)
		}
	}

	tcp := &dns.Server{
		Addr:    config.Address,
		Net:     "tcp",
		Handler: resolve("tcp"),
	}

	udp := &dns.Server{
		Addr:    config.Address,
		Net:     "udp",
		Handler: resolve("udp"),
	}

	return &Server{
		config: config,
		tcp:    tcp,
		udp:    udp,
	}
}

func (s *Server) Start() error {
	wg := &sync.WaitGroup{}

	errch := make(chan error)

	wg.Add(2)
	go s.runServer(wg, errch, s.tcp)
	go s.runServer(wg, errch, s.udp)

	go func() {
		wg.Wait()
		close(errch)
	}()

	for err := range errch {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Shutdown() error {
	err1 := s.tcp.Shutdown()
	err2 := s.udp.Shutdown()

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

	return nil
}

func (s *Server) runServer(wg *sync.WaitGroup, err chan<- error, server *dns.Server) {
	err <- server.ListenAndServe()
	wg.Done()
}
