package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	. "github.com/0xdeadbad/nginx-conf-nats/internal"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

type addOpts struct {
	Host    string `short:"h" long:"host" description:"Host to add or remove from nginx configuration" required:"true"`
	Ip      string `short:"i" long:"ip" description:"IP address of the host" required:"true"`
	Port    int    `short:"p" long:"port" description:"Port of the host" required:"true"`
	Https   bool   `short:"s" long:"https" description:"Use HTTPS" required:"false"`
	Ssl     bool   `short:"l" long:"ssl" description:"Use SSL" required:"false"`
	SslCert string `short:"c" long:"ssl-cert" description:"SSL certificate file" required:"false"`
	SslKey  string `short:"k" long:"ssl-key" description:"SSL key file" required:"false"`
}

type removeOpts struct {
	Host string `short:"h" long:"host" description:"Host to add or remove from nginx configuration" required:"true"`
}

var opts struct {
	Op           string
	SubComAdd    addOpts    `command:"add"`
	SubComRemove removeOpts `command:"remove"`
}

func main() {
	// ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	var nginxConf NginxConf

	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	_, err = flags.Parse(&opts)
	if err != nil {
		log.Fatalln(err)
	}

	if opts.SubComAdd.Host == "" && opts.SubComRemove.Host == "" {
		log.Fatalln("Host is required for add or remove operation")
	}

	if opts.SubComAdd.Host != "" {
		opts.Op = "add"
	} else {
		opts.Op = "remove"
	}

	if opts.Op == "add" {
		nginxConf = NginxConf{
			Op:      NginxConfSvcOp(opts.Op),
			Host:    opts.SubComAdd.Host,
			Ip:      opts.SubComAdd.Ip,
			Port:    opts.SubComAdd.Port,
			Https:   opts.SubComAdd.Https,
			Ssl:     opts.SubComAdd.Ssl,
			SslCert: opts.SubComAdd.SslCert,
			SslKey:  opts.SubComAdd.SslKey,
		}
	} else {
		nginxConf = NginxConf{
			Op:   NginxConfSvcOp(opts.Op),
			Host: opts.SubComRemove.Host,
		}
	}

	natsServerUrl := os.Getenv("NATS_SERVER_URL")

	nc, err := nats.Connect(natsServerUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer nc.Close()

	nginxConfSvcReq, err := json.Marshal(nginxConf)
	if err != nil {
		log.Fatalln(err)
	}

	msg, err := nc.Request("nginx-conf-svc", nginxConfSvcReq, time.Second*30)
	if err != nil {
		log.Fatalln(err)
	}

	reply := NginxSvcReply{}
	err = json.Unmarshal(msg.Data, &reply)
	if err != nil {
		log.Fatalln(err)
	}

	if reply.Err != "" {
		log.Fatalln(reply.Err)
	}

	log.Println("Ok")
}
