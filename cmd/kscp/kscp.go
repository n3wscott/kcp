package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/controlplane"
	"k8s.io/kubernetes/pkg/controlplane/options"
)

type envConfig struct {
	Port          int      `envconfig:"PORT" default:"8080" required:"true"`
	Endpoints     []string `envconfig:"KCP_ENDPOINTS" default:"https://localhost:2379" required:"true"`
	Dir           string   `envconfig:"KCP_DIR" default:".kcp/data" required:"true"`
	CertFile      string   `envconfig:"KCP_CERTFILE" default:".kcp/data/secrets/peer/cert.pem" required:"true"`
	KeyFile       string   `envconfig:"KCP_KEYFILE" default:".kcp/data/secrets/peer/key.pem" required:"true"`
	TrustedCAFile string   `envconfig:"KCP_CAFILE" default:".kcp/data/secrets/ca/cert.pem" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}

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
	//s := &etcd.Server{
	//	Dir: filepath.Join(dir, "data"),
	//}
	ctx := context.TODO()

	//if err := s.Run(func(cfg etcd.ClientInfo) error {
	//	c, err := clientv3.New(clientv3.Config{
	//		Endpoints: cfg.Endpoints,
	//		TLS:       cfg.TLS,
	//	})
	//	if err != nil {
	//		return err
	//	}
	//	defer c.Close()
	//	r, err := c.Cluster.MemberList(context.Background())
	//	if err != nil {
	//		return err
	//	}
	//	for _, member := range r.Members {
	//		fmt.Fprintf(os.Stderr, "Connected to etcd %d %s\n", member.GetID(), member.GetName())
	//	}

	serverOptions := options.NewServerRunOptions()
	serverOptions.SecureServing.BindPort = env.Port
	serverOptions.SecureServing.ServerCert.CertDirectory = env.Dir
	serverOptions.InsecureServing = nil
	serverOptions.Etcd.StorageConfig.Transport = storagebackend.TransportConfig{
		ServerList:    env.Endpoints,
		CertFile:      env.CertFile,
		KeyFile:       env.KeyFile,
		TrustedCAFile: env.TrustedCAFile,
	}
	cpOptions, err := controlplane.Complete(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	server, err := controlplane.CreateServerChain(cpOptions, ctx.Done())
	if err != nil {
		log.Fatal(err)
	}

	var clientConfig clientcmdapi.Config
	clientConfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		"loopback": {Token: server.LoopbackClientConfig.BearerToken},
	}
	clientConfig.Clusters = map[string]*clientcmdapi.Cluster{
		// admin is the virtual cluster running by default
		"admin": {
			Server:                   server.LoopbackClientConfig.Host,
			CertificateAuthorityData: server.LoopbackClientConfig.CAData,
			TLSServerName:            server.LoopbackClientConfig.TLSClientConfig.ServerName,
		},
		// user is a virtual cluster that is lazily instantiated
		"user": {
			Server:                   server.LoopbackClientConfig.Host + "/clusters/user",
			CertificateAuthorityData: server.LoopbackClientConfig.CAData,
			TLSServerName:            server.LoopbackClientConfig.TLSClientConfig.ServerName,
		},
	}
	clientConfig.Contexts = map[string]*clientcmdapi.Context{
		"admin": {Cluster: "admin", AuthInfo: "loopback"},
		"user":  {Cluster: "user", AuthInfo: "loopback"},
	}
	clientConfig.CurrentContext = "admin"
	if err := clientcmd.WriteToFile(clientConfig, filepath.Join(env.Dir, "admin.kubeconfig")); err != nil {
		log.Fatal(err)
	}

	prepared := server.PrepareRun()

	if err := prepared.Run(ctx.Done()); err != nil {
		log.Fatal(err)
	}
	//}); err != nil {
	//	log.Fatal(err)
	//}

}

/**

curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.1/install.sh | bash -s v0.18.1


*/
