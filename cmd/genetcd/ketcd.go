package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kcp-dev/kcp/pkg/etcd"
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

	cfg, err := s.Generate()
	if err != nil {
		log.Fatal(err)
	}

	if out, err := json.Marshal(cfg); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(out))
		fmt.Println("Dir:", s.Dir)
	}
}

/**

.kcp
└── data
    ├── apiserver.crt
    ├── apiserver.key
    └── secrets
        ├── ca
        │   ├── cert.pem
        │   └── key.pem
        ├── client
        │   ├── cert.pem
        │   └── key.pem
        └── peer
            ├── cert.pem
            └── key.pem


kubectl create secret tls kcp-ca \
  --cert=.kcp/data/secrets/ca/cert.pem \
  --key=.kcp/data/secrets/ca/key.pem

kubectl create secret tls kcp-client \
  --cert=.kcp/data/secrets/client/cert.pem \
  --key=.kcp/data/secrets/client/key.pem

kubectl create secret tls kcp-peer \
  --cert=.kcp/data/secrets/peer/cert.pem \
  --key=.kcp/data/secrets/peer/key.pem


*/
