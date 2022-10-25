package transport

import (
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

type PeerConfig struct {
	StunConfig
	TurnConfig
}

type StunConfig struct {
	StunAddress string `envconfig:"stun_address"`
}

type TurnConfig struct {
	TurnAddress  string `envconfig:"turn_address"`
	TurnUsername string `envconfig:"turn_username"`
	TurnPassword string `envconfig:"turn_password"`
}

type Candidate struct {
	Target    int                  `json:"target"`
	Candidate *webrtc.ICECandidate `json:"candidate"`
}

func buildWebRTCConfiguration(conf PeerConfig) webrtc.Configuration {
	servers := []webrtc.ICEServer{{URLs: []string{conf.StunAddress}}}
	if conf.TurnAddress != "" {
		servers = append(servers, webrtc.ICEServer{
			URLs:       []string{conf.TurnAddress},
			Username:   conf.TurnUsername,
			Credential: conf.TurnPassword,
		})
	}

	return webrtc.Configuration{
		ICEServers:   servers,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}
}

type PeerClient interface {
	AddMediaTrack(t mediadevices.Track)
	SetupOfferDescription()
	SetupAnswerDescription()
	SetRemoteDescription(desc webrtc.SessionDescription)
	AddICECandidate(candidate webrtc.ICECandidateInit)
	LocalDescription() *webrtc.SessionDescription
	OnICECandidate(handler func(*webrtc.ICECandidate))
	OnICEConnectionStateChange(handler func(connectionState webrtc.ICEConnectionState))
	Close()
}

type peerClient struct {
	conn *webrtc.PeerConnection
}

func NewPeerClient(config PeerConfig, codec *mediadevices.CodecSelector) PeerClient {
	c := buildWebRTCConfiguration(config)
	mediaEngine := webrtc.MediaEngine{}
	codec.Populate(&mediaEngine)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
	conn, err := api.NewPeerConnection(c)
	if err != nil {
		log.WithError(err).Fatal("new peer connection")
	}

	client := &peerClient{conn}

	return client
}

func (c *peerClient) AddMediaTrack(t mediadevices.Track) {
	_, err := c.conn.AddTransceiverFromTrack(
		t,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly},
	)
	if err != nil {
		log.WithError(err).Fatal("add transceiver from track")
	}
}

func (c *peerClient) SetupOfferDescription() {
	offer, err := c.conn.CreateOffer(nil)
	if err != nil {
		log.WithError(err).Fatal("create offer")
	}

	err = c.conn.SetLocalDescription(offer)
	if err != nil {
		log.WithError(err).Fatal("set local description")
	}
}

func (c *peerClient) SetupAnswerDescription() {
	answer, err := c.conn.CreateAnswer(nil)
	if err != nil {
		log.WithError(err).Fatal("create answer")
	}

	err = c.conn.SetLocalDescription(answer)
	if err != nil {
		log.WithError(err).Fatal("set local description")
	}
}

func (c *peerClient) SetRemoteDescription(desc webrtc.SessionDescription) {
	if err := c.conn.SetRemoteDescription(desc); err != nil {
		log.WithError(err).Fatal("set remote description")
	}
}

func (c *peerClient) AddICECandidate(candidate webrtc.ICECandidateInit) {
	err := c.conn.AddICECandidate(candidate)
	if err != nil {
		log.WithError(err).Fatal("add ice candidate")
	}
}

func (c *peerClient) LocalDescription() *webrtc.SessionDescription {
	return c.conn.LocalDescription()
}

func (c *peerClient) OnICECandidate(handler func(*webrtc.ICECandidate)) {
	c.conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			handler(candidate)
		}
	})
}

func (c *peerClient) OnICEConnectionStateChange(handler func(connectionState webrtc.ICEConnectionState)) {
	c.conn.OnICEConnectionStateChange(handler)
}

func (c *peerClient) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
