package main

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
	"videostreamer/src/common/transport"
)

type sfuClient struct {
	SID          string
	connectionID uint64
	peer         transport.PeerClient
	ws           transport.WebsocketClient
}

func newSfuClient(conf config) sfuClient {
	c := sfuClient{SID: conf.SID}
	c.ws = transport.NewWebsocketClient(conf.SfuAddr, c.processMessage)

	codecSelector := transport.GetCodecSelector(conf.MediaDevicesConfig)
	c.peer = transport.NewPeerClient(conf.PeerConfig, codecSelector)

	stream := transport.GetMediaDeviceStream(codecSelector, conf.MediaDevicesConfig)
	transport.BindStreamTrackListeners(stream, c.peer.AddMediaTrack)

	c.subscribeICEEvents()
	c.peer.SetupOfferDescription()

	c.sendDescription(joinMethod)

	return c
}

func (c *sfuClient) generateConnectionID() {
	c.connectionID = uint64(uuid.New().ID())
}

func (c *sfuClient) subscribeICEEvents() {
	c.peer.OnICECandidate(c.sendTrickle)

	c.peer.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.WithField("s", connectionState.String()).Info("Connection State has changed\n")
	})
}

func (c *sfuClient) Close() {
	if c.ws != nil {
		c.ws.Close()
	}

	if c.peer != nil {
		c.peer.Close()
	}
}
