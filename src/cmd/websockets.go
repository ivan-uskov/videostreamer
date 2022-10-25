package main

import (
	"encoding/json"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
	"github.com/sourcegraph/jsonrpc2"
	"videostreamer/src/common/transport"
)

const joinMethod = "join"
const offerMethod = "offer"
const answerMethod = "answer"
const trickleMethod = "trickle"

type Candidate struct {
	Target    int                  `json:"target"`
	Candidate *webrtc.ICECandidate `json:"candidate"`
}

type SendOffer struct {
	SID   string                     `json:"sid"`
	Offer *webrtc.SessionDescription `json:"offer"`
}

type SendAnswer struct {
	SID    string                     `json:"sid"`
	Answer *webrtc.SessionDescription `json:"answer"`
}

type TrickleResponse struct {
	Params ResponseCandidate `json:"params"`
	Method string            `json:"method"`
}

type ResponseCandidate struct {
	Target    int                      `json:"target"`
	Candidate *webrtc.ICECandidateInit `json:"candidate"`
}

type Response struct {
	Params *webrtc.SessionDescription `json:"params"`
	Result *webrtc.SessionDescription `json:"result"`
	Method string                     `json:"method"`
	Id     uint64                     `json:"id"`
}

func (c *sfuClient) sendDescription(method string) {
	c.generateConnectionID()

	var params []byte
	if method == answerMethod {
		params = transport.MarshalJson(&SendAnswer{
			Answer: c.peer.LocalDescription(),
			SID:    c.SID,
		})
	} else {
		params = transport.MarshalJson(&SendOffer{
			Offer: c.peer.LocalDescription(),
			SID:   c.SID,
		})
	}

	answerRequestJSON := transport.MarshalJson(&jsonrpc2.Request{
		Method: method,
		Params: (*json.RawMessage)(&params),
		ID: jsonrpc2.ID{
			IsString: false,
			Str:      "",
			Num:      c.connectionID,
		},
	})
	c.ws.SendMessage(answerRequestJSON)
}

func (c *sfuClient) sendTrickle(candidate *webrtc.ICECandidate) {
	candidateJSON := transport.MarshalJson(Candidate{
		Candidate: candidate,
		Target:    0,
	})
	trickleJSON := transport.MarshalJson(&jsonrpc2.Request{
		Method: trickleMethod,
		Params: (*json.RawMessage)(&candidateJSON),
	})
	c.ws.SendMessage(trickleJSON)
}

func (c *sfuClient) processMessage(message []byte) {
	var response Response
	transport.UnmarshalJson(message, &response)

	if response.Id == c.connectionID {
		log.Println(response)
		res := *response.Result
		c.peer.SetRemoteDescription(res)
	} else if response.Id != 0 && response.Method == offerMethod {
		c.peer.SetRemoteDescription(*response.Params)
		c.peer.SetupAnswerDescription()
		c.sendDescription(answerMethod)
	} else if response.Method == trickleMethod {
		var trickleResponse TrickleResponse
		transport.UnmarshalJson(message, &trickleResponse)
		c.peer.AddICECandidate(*trickleResponse.Params.Candidate)
	}
}
