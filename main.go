package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket" // Add this
)

const PubSubDiscoveryTopic string = "browser-peer-discovery"

func main() {
	ctx := context.Background()

	// 1. Get the port from Render's environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000" // Fallback for local testing
	}

	privk, err := LoadIdentity("identity.key")
	if err != nil {
		panic(err)
	}

	// 2. Configure for WebSockets
	// Note: We listen on 'ws' because Render's Load Balancer handles the 'wss' (SSL) part for us.
	wsAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", port)

	opts := []libp2p.Option{
		libp2p.Identity(privk),
		libp2p.Transport(websocket.New), // Use WebSockets for Render compatibility
		libp2p.ListenAddrStrings(wsAddr),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	// Enable Relay V2
	_, err = relay.New(host)
	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	discoveryTopic, err := ps.Join(PubSubDiscoveryTopic)
	if err != nil {
		panic(err)
	}
	_, err = discoveryTopic.Subscribe()
	if err != nil {
		panic(err)
	}

	log.Printf("PeerID: %s", host.ID().String())
	for _, addr := range host.Addrs() {
		log.Printf("Listening on: %s/p2p/%s\n", addr.String(), host.ID())
	}

	select {}
}
