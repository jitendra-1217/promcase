package main

import (
	"fmt"
	"github.com/jitendra-1217/promcase/promcase"
	"github.com/jitendra-1217/promcase/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	promcase.InitQueue(utils.GetEnvAsInt("QUEUE_LEN", "10000"))
	promcase.ProcessQueue()
	go listenUdp()
	go serveTcp()
	<-make(chan bool, 1)
}

// listenUdp function starts listening for Udp metrics
func listenUdp() {
	conn, err := net.ListenUDP(
		"udp",
		&net.UDPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: utils.GetEnvAsInt("UDP_PORT", "10001")})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	buf := make([]byte, utils.GetEnvAsInt("UDP_MESSAGE_MAX_LEN", "1024"))
	for {
		l, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Error(err)
			continue
		}
		message, err := promcase.NewMessage(string(buf[0:l]), addr.String())
		if err != nil {
			log.WithFields(log.Fields{}).Error(err)
			continue
		}
		promcase.Queue <- message
	}
}

// serveTcp function servers /metrics endpoint on Tcp
func serveTcp() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", utils.GetEnv("TCP_PORT", "10002")), nil))
}
