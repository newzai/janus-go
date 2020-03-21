package videoroom

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/newzai/janus-go/jwsapi"
	"github.com/newzai/janus-go/jwsapi/jplugin/jvideoroom"
	"github.com/newzai/janus-go/logging"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

//Track track
type Track struct {
	track *webrtc.Track
	seqNo uint16
}

//WriteRTP write rtp
func (t *Track) WriteRTP(packet *rtp.Packet) error {
	t.seqNo++
	packet.SequenceNumber = t.seqNo
	return t.track.WriteRTP(packet)
}

//SSRC return ssrc
func (t *Track) SSRC() uint32 {
	return t.track.SSRC()
}

//Publisher a publisher user,
type Publisher struct {
	BaseSession
	jPub    *jvideoroom.Publisher
	tracks  []*Track
	senders []*webrtc.RTPSender
}

//PublisherOption option
type PublisherOption func(*Publisher)

//WithPublisherConfigure set webrtc configure
func WithPublisherConfigure(configure webrtc.Configuration) PublisherOption {
	return func(p *Publisher) {
		p.configure = configure
	}
}

//NewPublisher new publihser
func NewPublisher(ctx context.Context, api *webrtc.API, h *jwsapi.Handle, room uint64, opts ...jvideoroom.PublisherOption) *Publisher {
	p := &Publisher{
		BaseSession: BaseSession{
			ctx:    ctx,
			api:    api,
			handle: h,
			configure: webrtc.Configuration{
				SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
			},
			remoteCandidates: make(chan jwsapi.Message, 8),
		},
		jPub: jvideoroom.NewPublisher(ctx, h, room, opts...),
	}
	return p
}

//Object return jvideoroom.Publisher
func (p *Publisher) Object() *jvideoroom.Publisher {
	return p.jPub
}

//ID return id info
func (p *Publisher) ID() string {
	return fmt.Sprintf("[%d.%d]", p.jPub.Room(), p.jPub.ID())
}

//SetOption set option
func (p *Publisher) SetOption(opts ...PublisherOption) *Publisher {
	for _, opt := range opts {
		opt(p)
	}
	return p
}

//Join join to the
func (p *Publisher) Join(opts ...jwsapi.MessageOption) error {
	err := p.jPub.Join(opts...)
	return err
}

//Publish start send stream
func (p *Publisher) Publish(audio bool, video bool, opts ...jwsapi.MessageOption) error {

	pc, err := p.api.NewPeerConnection(p.configure)
	if err != nil {
		return errors.Wrap(err, "NewPeerConnection")
	}
	p.pc = pc

	p.handle.SetCallback(jwsapi.WithHandleHangup(p.onHangup))
	p.handle.SetCallback(jwsapi.WithHandleWebrtcup(p.onWebrtcup))
	p.handle.SetCallback(jwsapi.WithHandleMedia(p.onMedia))
	p.handle.SetCallback(jwsapi.WithHandleSlowLink(p.onSlowlink))

	pc.OnConnectionStateChange(p.onPeerConnectionState)
	pc.OnICECandidate(p.onICECandidate)
	pc.OnICEConnectionStateChange(p.onICEConnectionStateChange)

	audioTrack, err := pc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "pionA0")
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "NewTrack(Audio)")

	}
	audioSender, err := pc.AddTrack(audioTrack)
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "AddTrack(Audio)")

	}
	videoTrack, err := pc.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pionV0")
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "NewTrack(Video)")

	}

	videoSender, err := pc.AddTrack(videoTrack)
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "AddTrack(Video)")

	}

	p.tracks = append(p.tracks, &Track{audioTrack, 0}, &Track{videoTrack, 0})
	p.senders = append(p.senders, audioSender, videoSender)

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "CreateOffer(Video)")

	}

	err = pc.SetLocalDescription(offer)
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "SetLocalDescription(Video)")
	}

	answer, err := p.jPub.Publish(audio, video, false, offer.SDP, true, opts...)
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "publish")
	}
	err = pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answer,
	})
	if err != nil {
		pc.Close()
		return errors.Wrap(err, "SetRemoteDescription")
	}
	return nil
}

//Unpublish cancel publish
func (p *Publisher) Unpublish() error {

	if p.pc != nil {
		p.pc.Close()
		p.pc = nil

	}
	return p.jPub.Unpublish()
}

//GetTrack return track by kind
func (p *Publisher) GetTrack(kind webrtc.RTPCodecType) *Track {

	switch kind {
	case webrtc.RTPCodecTypeAudio:
		return p.tracks[0]
	case webrtc.RTPCodecTypeVideo:
		return p.tracks[1]
	default:
		return nil
	}
}

func (p *Publisher) onHangup(msg jwsapi.Message) {
	if p.pc != nil {
		p.pc.Close()
		p.pc = nil
	}
}
func (p *Publisher) onWebrtcup(msg jwsapi.Message) {
	logging.Infof("%s webrtcup ", p.ID())
}

func (p *Publisher) onMedia(msg jwsapi.Message) {

}
func (p *Publisher) onSlowlink(msg jwsapi.Message) {

}

func (p *Publisher) onICEConnectionStateChange(state webrtc.ICEConnectionState) {
	logging.Infof("%s ICEConnectionState %s", p.ID(), state.String())
}
func (p *Publisher) onPeerConnectionState(state webrtc.PeerConnectionState) {
	logging.Infof("%s PeerConnectionState %s", p.ID(), state.String())
}
