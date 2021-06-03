package etcd

import (
	"crypto/tls"
	"fmt"
	"time"

	"go.etcd.io/etcd/embed"
)

type Server struct {
	Dir string
}

type ClientInfo struct {
	Endpoints []string
	TLS       *tls.Config `json:"-"`

	CertFile      string
	KeyFile       string
	TrustedCAFile string
}

func (s *Server) Run(fn func(ClientInfo) error) error {
	cfg := getCfg(s.Dir)

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return err
	}
	defer e.Close()

	clientConfig, err := cfg.ClientTLSInfo.ClientConfig()
	if err != nil {
		return err
	}

	select {
	case <-e.Server.ReadyNotify():
		return fn(ClientInfo{
			Endpoints:     []string{cfg.ACUrls[0].String()},
			TLS:           clientConfig,
			CertFile:      cfg.ClientTLSInfo.CertFile,
			KeyFile:       cfg.ClientTLSInfo.KeyFile,
			TrustedCAFile: cfg.ClientTLSInfo.TrustedCAFile,
		})
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		return fmt.Errorf("server took too long to start")
	case e, _ := <-e.Err():
		return e
	}
}
