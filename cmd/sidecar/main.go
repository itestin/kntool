package main

import (
	"kntool/sidecar"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	r := gin.Default()
	r.GET("/show", sidecar.HandlerShow)
	r.GET("/reset", sidecar.HandlerReset)
	r.PUT("/update/:devices", sidecar.HandlerUpdateDevices)
	r.GET("/latency/:latency", sidecar.HandlerLatency)

	err := r.Run(":2332")
	if err != nil {
		logrus.Fatal(err)
	}
}
