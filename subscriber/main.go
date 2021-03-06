package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	hostname, _ := os.Hostname()
	hostname += "|" + RandStringBytes(12)
	server := flag.String("server", "tcp://127.0.0.1:1883", "The full URL of the mqtt server to connect to")
	topic := flag.String("topic", "", "Topic to publish the messages on")
	qos := flag.Int("qos", 0, "The QoS to send the messages at")
	clientID := flag.String("id", hostname, "A clientID for the connection")
	cleanSession := flag.Bool("clean", true, "clean session for connection")

	flag.Parse()
	log.Println("clientID: ", *clientID)
	log.Println("connecting: ", *server)

	opts := mqtt.NewClientOptions().AddBroker(*server)
	opts.SetClientID(*clientID)
	opts.SetCleanSession(*cleanSession)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetProtocolVersion(4)

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		fmt.Printf("!!!!!! mqtt connection lost error: %s\n" + err.Error())
	})

	opts.SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
		fmt.Println("...... mqtt reconnecting ......")
	})

	opts.OnConnect = func(client mqtt.Client) {
		log.Printf("subscribing %s at qos-%d\n", *topic, *qos)
		if token := client.Subscribe(*topic, byte(*qos), sampleSubs); token.Wait() && token.Error() != nil {
			log.Print(token.Error())
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
		return
	}

	log.Println("connect success: ", *server)

	<-signals

	//client.Unsubscribe(*topic)
	client.Disconnect(5)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func sampleSubs(_ mqtt.Client, msg mqtt.Message) {
	if msg.Qos() == byte(2) {
		msg.Ack()
	}

	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
	fmt.Printf("FULL_MSG: %+v\n", msg)
	fmt.Println("---------------------------------")
}
