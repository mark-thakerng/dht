package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"dht/internal"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func makeHost(port int) (host.Host, error) {
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	addr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(addr)
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	port := flag.Int("port", 7654, "port to listen on")
	peers := flag.String("peers", "", "peers to connect to")
	rendezvous := flag.String("topic", "librum", "topic to subscribe/publish to")
	mode := flag.String("mode", "", "subscribe or publish")

	help := flag.Bool("help", false, "display help")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	host, err := makeHost(*port)
	if err != nil {
		panic(err)
	}
	fmt.Printf("host ID %s\n", host.ID().Pretty())
	fmt.Println("host address: ")
	for _, addr := range host.Addrs() {
		fmt.Println("\t", addr.String())
	}
	gossipSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}
	discoveryPeers := []multiaddr.Multiaddr{}

	if *peers != "" {
		for _, dest := range strings.Split(*peers, ",") {
			multiAddr, err := multiaddr.NewMultiaddr(dest)
			if err != nil {
				panic(err)
			}
			discoveryPeers = append(discoveryPeers, multiAddr)
		}
	}
	dht, err := internal.NewDHT(ctx, host, discoveryPeers)
	if err != nil {
		panic(err)
	}
	go internal.Discover(ctx, host, dht, *rendezvous)
	topic, err := gossipSub.Join(*rendezvous)
	if err != nil {
		panic(err)
	}

	// If subscribe
	subscriber, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}
	if *mode == "sub" {
		subscribe(subscriber, ctx, host.ID())
	} else if *mode == "pub" {
		publish(ctx, topic)
	} else {
		select {}
	}
}

func subscribe(subscriber *pubsub.Subscription, ctx context.Context, hostID peer.ID) {
	for {
		msg, err := subscriber.Next(ctx)
		if err != nil {
			panic(err)
		}

		// only consider messages delivered by other peers
		if msg.ReceivedFrom == hostID {
			continue
		}

		fmt.Printf("got message: %s, from: %s\n", string(msg.Data), msg.ReceivedFrom.Pretty())
	}
}

func publish(ctx context.Context, topic *pubsub.Topic) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Printf("enter message to publish: \n")

			msg := scanner.Text()
			if len(msg) != 0 {
				// publish message to topic
				bytes := []byte(msg)
				topic.Publish(ctx, bytes)
			}
		}
	}
}
