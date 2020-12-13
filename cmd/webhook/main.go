package main

import (
	"flag"

	"kntool/webhook"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	certFile     = flag.String("cert-file", "server.crt", "Cert file")
	keyFile      = flag.String("key-file", "server.pem", "Key file")
	sidecarImage = flag.String("sidecar-image", "daocloud.io/loulan/kntool-sidecar:latest", "Sidecar image")
	sidecarPort  = flag.String("sidecar-port", "2332", "Sidecar port")
)

func main() {
	flag.Parse()
	if *certFile == "" || *keyFile == "" {
		logrus.Fatal("Run 'kntool --help' for usage.")
	}

	r := gin.Default()
	r.POST("/mutate", webhook.HandlerMutate)

	err := r.RunTLS(":9000", *certFile, *keyFile)
	if err != nil {
		logrus.Fatal(err)
	}
}
