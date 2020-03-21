package videoroom

import (
	"fmt"
	"strings"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
)

type rtpamp struct {
	pt        uint8
	name      string
	clockrate int
	channels  int
}

func newrtpmap(value string) *rtpamp {
	m := &rtpamp{}
	value = strings.ReplaceAll(value, "/", " ")
	fmt.Sscanf(value, "%d %s %d %d", &m.pt, &m.name, &m.clockrate, &m.channels)
	return m
}

func (rm *rtpamp) getPayloader() rtp.Payloader {

	switch strings.ToLower(rm.name) {
	case "pcma":

		return &codecs.G711Payloader{}
	case "opus":
		return &codecs.OpusPayloader{}
	case "g722":
		return &codecs.G722Payloader{}
	case "vp8":
		return &codecs.VP8Payloader{}
	case "vp9":
		return &codecs.VP8Payloader{}
	case "h264":
		return &codecs.H264Payloader{}
	case "rsfec":
		return &codecs.OpusPayloader{}
	case "flexfec-03":
		return &codecs.G711Payloader{}
	default:
		return &codecs.G711Payloader{}
	}
}

func getCodecs(sd *sdp.SessionDescription) map[uint8]*webrtc.RTPCodec {

	var lastCodec *webrtc.RTPCodec
	var kind webrtc.RTPCodecType
	allCodecs := make(map[uint8]*webrtc.RTPCodec)
	for _, m := range sd.MediaDescriptions {
		kind = webrtc.NewRTPCodecType(m.MediaName.Media)
		for _, a := range m.Attributes {
			switch a.Key {
			case "rtpmap":
				rm := newrtpmap(a.Value)
				lastCodec = &webrtc.RTPCodec{
					RTPCodecCapability: webrtc.RTPCodecCapability{
						MimeType:  kind.String() + "/" + rm.name,
						ClockRate: uint32(rm.clockrate),
						Channels:  uint16(rm.channels),
						//SDPFmtpLine:  fmtp, delay setting
						//RTCPFeedback: fbs, delay setting
					},
					PayloadType: uint8(rm.pt),
					Payloader:   rm.getPayloader(),
					Type:        kind,
					Name:        rm.name,
				}

				allCodecs[rm.pt] = lastCodec

			case "fmtp":
				var pt uint8
				var param string
				fmt.Sscanf(a.Value, "%d %s", &pt, &param)
				if pt == lastCodec.PayloadType {
					lastCodec.SDPFmtpLine = param
				} else if c, ok := allCodecs[pt]; ok {
					c.SDPFmtpLine = param
				}
			case "rtcp-fb":
				var pt uint8
				var tpe, param string
				fmt.Sscanf(a.Value, "%d %s %s", &pt, &tpe, &param)
				if pt == lastCodec.PayloadType {
					lastCodec.RTCPFeedback = append(lastCodec.RTCPFeedback, webrtc.RTCPFeedback{Type: tpe, Parameter: param})
				} else if c, ok := allCodecs[pt]; ok {
					c.RTCPFeedback = append(c.RTCPFeedback, webrtc.RTCPFeedback{Type: tpe, Parameter: param})
				}
			}
		}
	}

	return allCodecs
}

func initAPI(remoteSDP string) *webrtc.API {
	sd := sdp.SessionDescription{}
	err := sd.Unmarshal([]byte(remoteSDP))
	if err != nil {
		return nil
	}
	cc := getCodecs(&sd)

	m := webrtc.MediaEngine{}

	for _, codec := range cc {
		switch codec.Type {
		case webrtc.RTPCodecTypeAudio:
			if strings.EqualFold(codec.Name, "opus") {
				m.RegisterCodec(codec)
			}
		case webrtc.RTPCodecTypeVideo:
			if strings.EqualFold(codec.Name, "h264") {
				m.RegisterCodec(codec)
			}
		}
	}

	setting := webrtc.SettingEngine{}
	setting.SetEphemeralUDPPortRange(20000, 40000)

	return webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithSettingEngine(setting))
}
