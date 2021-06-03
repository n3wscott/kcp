package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kcp-dev/kcp/pkg/etcd"
	"go.etcd.io/etcd/clientv3"
)

func main() {
	dir := filepath.Join(".", ".kcp")
	if fi, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Fatal(err)
		}
	} else {
		if !fi.IsDir() {
			log.Fatal(fmt.Errorf("%q is a file, please delete or select another location", dir))
		}
	}
	s := &etcd.Server{
		Dir: filepath.Join(dir, "data"),
	}
	ctx := context.Background()

	if err := s.Run(func(cfg etcd.ClientInfo) error {
		c, err := clientv3.New(clientv3.Config{
			Endpoints: cfg.Endpoints,
			TLS:       cfg.TLS,
		})
		if err != nil {
			return err
		}
		defer c.Close()
		r, err := c.Cluster.MemberList(context.Background())
		if err != nil {
			return err
		}
		for _, member := range r.Members {
			fmt.Fprintf(os.Stderr, "Connected to etcd %d %s\n", member.GetID(), member.GetName())
		}

		if out, err := json.Marshal(cfg); err != nil {
			return err
		} else {
			fmt.Println(string(out))
			fmt.Println("Dir:", s.Dir)

		}

		//serverOptions := options.NewServerRunOptions()
		//serverOptions.SecureServing.BindPort = env.Port
		//serverOptions.SecureServing.ServerCert.CertDirectory = s.Dir
		//serverOptions.InsecureServing = nil
		//serverOptions.Etcd.StorageConfig.Transport = storagebackend.TransportConfig{
		//	ServerList:    cfg.Endpoints,
		//	CertFile:      cfg.CertFile,
		//	KeyFile:       cfg.KeyFile,
		//	TrustedCAFile: cfg.TrustedCAFile,
		//}
		//cpOptions, err := controlplane.Complete(serverOptions)
		//if err != nil {
		//	return err
		//}
		//
		//server, err := controlplane.CreateServerChain(cpOptions, ctx.Done())
		//if err != nil {
		//	return err
		//}

		//var clientConfig clientcmdapi.Config
		//clientConfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		//	"loopback": {Token: server.LoopbackClientConfig.BearerToken},
		//}
		//clientConfig.Clusters = map[string]*clientcmdapi.Cluster{
		//	// admin is the virtual cluster running by default
		//	"admin": {
		//		Server:                   server.LoopbackClientConfig.Host,
		//		CertificateAuthorityData: server.LoopbackClientConfig.CAData,
		//		TLSServerName:            server.LoopbackClientConfig.TLSClientConfig.ServerName,
		//	},
		//	// user is a virtual cluster that is lazily instantiated
		//	"user": {
		//		Server:                   server.LoopbackClientConfig.Host + "/clusters/user",
		//		CertificateAuthorityData: server.LoopbackClientConfig.CAData,
		//		TLSServerName:            server.LoopbackClientConfig.TLSClientConfig.ServerName,
		//	},
		//}
		//clientConfig.Contexts = map[string]*clientcmdapi.Context{
		//	"admin": {Cluster: "admin", AuthInfo: "loopback"},
		//	"user":  {Cluster: "user", AuthInfo: "loopback"},
		//}
		//clientConfig.CurrentContext = "admin"
		//if err := clientcmd.WriteToFile(clientConfig, filepath.Join(s.Dir, "admin.kubeconfig")); err != nil {
		//	return err
		//}

		<-ctx.Done()
		return ctx.Err()
	}); err != nil {
		log.Fatal(err)
	}

}
