package main

import (
	"context"
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	websocket "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

// Configuring libp2p
func SetupLibp2p(ctx context.Context,
	hostKey crypto.PrivKey,
	secret pnet.PSK,
	listenAddrs []multiaddr.Multiaddr,
	ds datastore.Batching,
	opts ...libp2p.Option) (host.Host, *dht.IpfsDHT, peer.ID, error) {
	var ddht *dht.IpfsDHT

	var err error
	var transports = libp2p.DefaultTransports
	//var transports = libp2p.NoTransports
	if secret != nil {
		transports = libp2p.ChainOptions(
			libp2p.NoTransports,
			libp2p.Transport(tcp.NewTCPTransport, websocket.New),
			//libp2p.Transport(websocket.New),
		)
	}

	finalOpts := []libp2p.Option{
		libp2p.Identity(hostKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.PrivateNetwork(secret),
		transports,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			ddht, err = newDHT2(ctx, h, ds)

			return ddht, err

		}),
	}
	finalOpts = append(finalOpts, opts...)

	h, err := libp2p.New(
		finalOpts...,
	)
	if err != nil {
		return nil, nil, "", fmt.Errorf("SetupLibp2p>libp2p.New error: %w", err)
	}

	pid, err := peer.IDFromPublicKey(hostKey.GetPublic())
	if err != nil {
		return nil, nil, "", fmt.Errorf("SetupLibp2p>peer.IDFromPublicKey error: %w", err)
	}
	// Connect to default peers

	return h, ddht, pid, nil
}

// Create a new DHT instance
func newDHT2(ctx context.Context, h host.Host, ds datastore.Batching) (*dht.IpfsDHT, error) {
	var options []dht.Option

	// If no bootstrap peers, this peer acts as a bootstrapping node
	// Other peers can use this peer's IPFS address for peer discovery via DHT
	options = append(options, dht.Mode(dht.ModeAuto))

	kdht, err := dht.New(ctx, h, options...)
	if err != nil {
		return nil, fmt.Errorf("SetupnewDHT2Libp2p>dht.New error: %w", err)
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("SetupnewDHT2Libp2p>kdht.Bootstrap error: %w", err)
	}

	return kdht, nil
}
