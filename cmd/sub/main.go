package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	. "github.com/0xdeadbad/nginx-conf-nats/internal"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func handleNginxConfMsg(msg *nats.Msg, nginxPid int, tmpl *template.Template) error {
	var nginxConf NginxConf
	err := json.Unmarshal(msg.Data, &nginxConf)
	if err != nil {
		reply := NginxSvcReply{Err: err.Error()}
		// TODO: treat the error
		replyBytes, err := json.Marshal(reply)
		log.Println(err)

		// TODO: treat the error
		err = msg.Respond(replyBytes)
		log.Println(err)

		return err
	}

	nginxConfFilesDir := os.Getenv("NGINX_CONF_FILES_DIR")

	switch nginxConf.Op {
	case NginxConfSvcOpAdd:
		// Add a new server block to the nginx config
		// and render the server block template

		f, err := os.Create(fmt.Sprintf("%s/%s.conf", nginxConfFilesDir, nginxConf.Host))
		if err != nil {
			reply := NginxSvcReply{Err: err.Error()}
			// TODO: treat the error
			replyBytes, err := json.Marshal(reply)
			log.Println(err)

			// TODO: treat the error
			err = msg.Respond(replyBytes)
			log.Println(err)

			return err
		}
		defer f.Close()

		err = tmpl.Execute(f, nginxConf)
		if err != nil {
			reply := NginxSvcReply{Err: err.Error()}
			// TODO: treat the error
			replyBytes, err := json.Marshal(reply)
			log.Println(err)

			// TODO: treat the error
			err = msg.Respond(replyBytes)
			log.Println(err)

			return err
		}

	case NginxConfSvcOpRemove:
		// Remove a server block from the nginx config

		err = os.Remove(fmt.Sprintf("%s/%s.conf", nginxConfFilesDir, nginxConf.Host))
		if err != nil {
			reply := NginxSvcReply{Err: err.Error()}
			// TODO: treat the error
			replyBytes, err := json.Marshal(reply)
			log.Println(err)

			// TODO: treat the error
			err = msg.Respond(replyBytes)
			log.Println(err)

			return err
		}
	}

	// Send SIGHUP to the nginx process to reload the config
	err = syscall.Kill(nginxPid, syscall.SIGHUP)
	if err != nil {
		reply := NginxSvcReply{Err: err.Error()}
		// TODO: treat the error
		replyBytes, err := json.Marshal(reply)
		log.Println(err)

		// TODO: treat the error
		err = msg.Respond(replyBytes)
		log.Println(err)

		return err
	}

	// Respond to the request

	reply := NginxSvcReply{}
	replyBytes, err := json.Marshal(reply)
	if err != nil {
		// TODO: treat the error
		log.Println(err)
	}

	err = msg.Respond(replyBytes)
	if err != nil {
		// TODO: treat the error
		log.Println(err)
	}

	log.Printf("Nginx conf updated: %+v\n", nginxConf)

	return nil
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	natsServerUrl := os.Getenv("NATS_SERVER_URL")
	nginxPidFile := os.Getenv("NGINX_PID_FILE")
	templateFile := os.Getenv("NGINX_CONF_TEMPLATE_FILE")

	tmpl, err := template.New(templateFile).ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}

	buff, err := os.ReadFile(nginxPidFile)
	if err != nil {
		panic(err)
	}

	nginxPid, err := strconv.Atoi(strings.Trim(string(buff), "\n"))
	if err != nil {
		panic(err)
	}

	nc, err := nats.Connect(natsServerUrl)
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	ch := make(chan *nats.Msg, 8)
	sub, err := nc.ChanSubscribe("nginx-conf-svc", ch)
	if err != nil {
		panic(err)
	}

	log.Println("Nginx conf service is running")
mainFor:
	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			close(ch)
			break mainFor
		case msg := <-ch:
			err = handleNginxConfMsg(msg, nginxPid, tmpl)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
