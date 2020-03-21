package videoroom

import (
	"context"

	"github.com/newzai/janus-go/jwsapi"
	"github.com/pion/webrtc/v2"
)

//BaseSession base session
type BaseSession struct {
	ctx              context.Context
	api              *webrtc.API
	pc               *webrtc.PeerConnection
	configure        webrtc.Configuration
	handle           *jwsapi.Handle
	remoteCandidates chan jwsapi.Message
}

func (s *BaseSession) doRemoteCandidate(candidates chan jwsapi.Message) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg, ok := <-candidates:
			if !ok {
				return
			}
			candidate, ok := msg.String("candidate")
			if !ok {
				continue
			}
			sdpMLineIndex, _ := msg.Uint16("sdpMLineIndex")
			sdpMid, _ := msg.String("sdpMid")
			iceCandidate := webrtc.ICECandidateInit{
				Candidate:     candidate,
				SDPMLineIndex: &sdpMLineIndex,
				SDPMid:        &sdpMid,
			}
			err := s.pc.AddICECandidate(iceCandidate)
			if err != nil {
				panic(candidate + "---" + err.Error())
			}
		}
	}
}

func (s *BaseSession) onCandidate(msg jwsapi.Message) {

	s.remoteCandidates <- msg

}

func (s *BaseSession) onTrickle(msg jwsapi.Message) {
	candidate, ok := msg.SubMessage("candidate")
	if !ok {
		return
	}
	if completed := msg.Bool("completed"); completed {
		return
	}
	s.onCandidate(candidate)

}

func (s *BaseSession) onICECandidate(candidate *webrtc.ICECandidate) {

	if candidate == nil {
		s.handle.Trickle(jwsapi.Message{
			"completed": true,
		})
	} else {
		s.handle.Trickle(jwsapi.Message{
			"candidate":     candidate.String(),
			"sdpMLineIndex": 0,
			"sdpMid":        "audio",
		})
	}
}
