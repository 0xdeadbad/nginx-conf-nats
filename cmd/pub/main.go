package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	. "github.com/0xdeadbad/nginx-conf-nats/internal"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

var opts struct {
	Op      string `short:"o" long:"op" description:"Operation to perform" required:"true"`
	Host    string `short:"h" long:"host" description:"Host to add or remove from nginx configuration" required:"true"`
	Ip      string `short:"i" long:"ip" description:"IP address of the host" required:"true"`
	Port    int    `short:"p" long:"port" description:"Port of the host" required:"true"`
	Https   bool   `short:"s" long:"https" description:"Use HTTPS" required:"false"`
	Ssl     bool   `short:"l" long:"ssl" description:"Use SSL" required:"false"`
	SslCert string `short:"c" long:"ssl-cert" description:"SSL certificate file" required:"false"`
	SslKey  string `short:"k" long:"ssl-key" description:"SSL key file" required:"false"`
}

func main() {
	// ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	_, err = flags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	nginxConf := NginxConf{
		Op:      NginxConfSvcOp(opts.Op),
		Host:    opts.Host,
		Ip:      opts.Ip,
		Port:    opts.Port,
		Https:   opts.Https,
		Ssl:     opts.Ssl,
		SslCert: opts.SslCert,
		SslKey:  opts.SslKey,
	}

	natsServerUrl := os.Getenv("NATS_SERVER_URL")

	nc, err := nats.Connect(natsServerUrl)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	nginxConfSvcReq, err := json.Marshal(nginxConf)
	if err != nil {
		panic(err)
	}

	msg, err := nc.Request("nginx-conf-svc", nginxConfSvcReq, time.Second*30)
	if err != nil {
		panic(err)
	}

	reply := NginxSvcReply{}
	err = json.Unmarshal(msg.Data, &reply)
	if err != nil {
		panic(err)
	}

	if reply.Err != "" {
		log.Println(reply.Err)
		os.Exit(1)
	}

	log.Println("Ok")
}
