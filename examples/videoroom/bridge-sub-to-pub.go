package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/cihub/seelog"
	"github.com/newzai/janus-go/jwsapi"
	"github.com/newzai/janus-go/jwsapi/jplugin/jvideoroom"
	"github.com/newzai/janus-go/logging"
	"github.com/newzai/janus-go/videoroom"
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

var offer = "v=0\r\no=- 595052046726569363 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1\r\na=msid-semantic: WMS UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111 103 104 9 0 8 106 105 13 110 112 113 126\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:qSQC\r\na=ice-pwd:+c1wkd1mdilWFVJ2N8ITwwNr\r\na=ice-options:trickle\r\na=fingerprint:sha-256 A1:8D:5D:B4:58:B3:5E:17:4F:70:72:46:C5:6A:0B:53:D5:71:89:33:72:4D:3A:42:34:DE:49:03:47:E6:A5:15\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=extmap:5 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\r\na=extmap:6 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\r\na=sendonly\r\na=msid:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7 7ea8fba8-ce06-4e5a-a875-dba416509e9b\r\na=rtcp-mux\r\na=rtpmap:111 opus/48000/2\r\na=rtcp-fb:111 transport-cc\r\na=fmtp:111 minptime=10;useinbandfec=1\r\na=rtpmap:103 ISAC/16000\r\na=rtpmap:104 ISAC/32000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:106 CN/32000\r\na=rtpmap:105 CN/16000\r\na=rtpmap:13 CN/8000\r\na=rtpmap:110 telephone-event/48000\r\na=rtpmap:112 telephone-event/32000\r\na=rtpmap:113 telephone-event/16000\r\na=rtpmap:126 telephone-event/8000\r\na=ssrc:3118459916 cname:zGKq4uLozdgh3NpM\r\na=ssrc:3118459916 msid:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7 7ea8fba8-ce06-4e5a-a875-dba416509e9b\r\na=ssrc:3118459916 mslabel:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7\r\na=ssrc:3118459916 label:7ea8fba8-ce06-4e5a-a875-dba416509e9b\r\nm=video 9 UDP/TLS/RTP/SAVPF 96 97 98 99 100 101 102 122 127 121 125 107 108 109 124 120 123 119 114 115 116\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:qSQC\r\na=ice-pwd:+c1wkd1mdilWFVJ2N8ITwwNr\r\na=ice-options:trickle\r\na=fingerprint:sha-256 A1:8D:5D:B4:58:B3:5E:17:4F:70:72:46:C5:6A:0B:53:D5:71:89:33:72:4D:3A:42:34:DE:49:03:47:E6:A5:15\r\na=setup:actpass\r\na=mid:1\r\na=extmap:14 urn:ietf:params:rtp-hdrext:toffset\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:13 urn:3gpp:video-orientation\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:12 http://www.webrtc.org/experiments/rtp-hdrext/playout-delay\r\na=extmap:11 http://www.webrtc.org/experiments/rtp-hdrext/video-content-type\r\na=extmap:7 http://www.webrtc.org/experiments/rtp-hdrext/video-timing\r\na=extmap:8 http://tools.ietf.org/html/draft-ietf-avtext-framemarking-07\r\na=extmap:9 http://www.webrtc.org/experiments/rtp-hdrext/color-space\r\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=extmap:5 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\r\na=extmap:6 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\r\na=sendonly\r\na=msid:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7 763903e1-83d0-4dc2-9834-6b4a508d283a\r\na=rtcp-mux\r\na=rtcp-rsize\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 goog-remb\r\na=rtcp-fb:96 transport-cc\r\na=rtcp-fb:96 ccm fir\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtpmap:97 rtx/90000\r\na=fmtp:97 apt=96\r\na=rtpmap:98 VP9/90000\r\na=rtcp-fb:98 goog-remb\r\na=rtcp-fb:98 transport-cc\r\na=rtcp-fb:98 ccm fir\r\na=rtcp-fb:98 nack\r\na=rtcp-fb:98 nack pli\r\na=fmtp:98 profile-id=0\r\na=rtpmap:99 rtx/90000\r\na=fmtp:99 apt=98\r\na=rtpmap:100 VP9/90000\r\na=rtcp-fb:100 goog-remb\r\na=rtcp-fb:100 transport-cc\r\na=rtcp-fb:100 ccm fir\r\na=rtcp-fb:100 nack\r\na=rtcp-fb:100 nack pli\r\na=fmtp:100 profile-id=2\r\na=rtpmap:101 rtx/90000\r\na=fmtp:101 apt=100\r\na=rtpmap:102 H264/90000\r\na=rtcp-fb:102 goog-remb\r\na=rtcp-fb:102 transport-cc\r\na=rtcp-fb:102 ccm fir\r\na=rtcp-fb:102 nack\r\na=rtcp-fb:102 nack pli\r\na=fmtp:102 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f\r\na=rtpmap:122 rtx/90000\r\na=fmtp:122 apt=102\r\na=rtpmap:127 H264/90000\r\na=rtcp-fb:127 goog-remb\r\na=rtcp-fb:127 transport-cc\r\na=rtcp-fb:127 ccm fir\r\na=rtcp-fb:127 nack\r\na=rtcp-fb:127 nack pli\r\na=fmtp:127 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f\r\na=rtpmap:121 rtx/90000\r\na=fmtp:121 apt=127\r\na=rtpmap:125 H264/90000\r\na=rtcp-fb:125 goog-remb\r\na=rtcp-fb:125 transport-cc\r\na=rtcp-fb:125 ccm fir\r\na=rtcp-fb:125 nack\r\na=rtcp-fb:125 nack pli\r\na=fmtp:125 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f\r\na=rtpmap:107 rtx/90000\r\na=fmtp:107 apt=125\r\na=rtpmap:108 H264/90000\r\na=rtcp-fb:108 goog-remb\r\na=rtcp-fb:108 transport-cc\r\na=rtcp-fb:108 ccm fir\r\na=rtcp-fb:108 nack\r\na=rtcp-fb:108 nack pli\r\na=fmtp:108 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f\r\na=rtpmap:109 rtx/90000\r\na=fmtp:109 apt=108\r\na=rtpmap:124 H264/90000\r\na=rtcp-fb:124 goog-remb\r\na=rtcp-fb:124 transport-cc\r\na=rtcp-fb:124 ccm fir\r\na=rtcp-fb:124 nack\r\na=rtcp-fb:124 nack pli\r\na=fmtp:124 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4d0032\r\na=rtpmap:120 rtx/90000\r\na=fmtp:120 apt=124\r\na=rtpmap:123 H264/90000\r\na=rtcp-fb:123 goog-remb\r\na=rtcp-fb:123 transport-cc\r\na=rtcp-fb:123 ccm fir\r\na=rtcp-fb:123 nack\r\na=rtcp-fb:123 nack pli\r\na=fmtp:123 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=640032\r\na=rtpmap:119 rtx/90000\r\na=fmtp:119 apt=123\r\na=rtpmap:114 red/90000\r\na=rtpmap:115 rtx/90000\r\na=fmtp:115 apt=114\r\na=rtpmap:116 ulpfec/90000\r\na=ssrc-group:FID 3719848789 1713948480\r\na=ssrc:3719848789 cname:zGKq4uLozdgh3NpM\r\na=ssrc:3719848789 msid:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7 763903e1-83d0-4dc2-9834-6b4a508d283a\r\na=ssrc:3719848789 mslabel:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7\r\na=ssrc:3719848789 label:763903e1-83d0-4dc2-9834-6b4a508d283a\r\na=ssrc:1713948480 cname:zGKq4uLozdgh3NpM\r\na=ssrc:1713948480 msid:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7 763903e1-83d0-4dc2-9834-6b4a508d283a\r\na=ssrc:1713948480 mslabel:UEHIvyhJOCvHzH3Ynk00xF8XrYn67EW2aYF7\r\na=ssrc:1713948480 label:763903e1-83d0-4dc2-9834-6b4a508d283a\r\n"

var (
	// Media engine
	m       webrtc.MediaEngine
	setting webrtc.SettingEngine
	// API object
	api *webrtc.API
)

//Init init webrt
func Init(portMin int, portMax int) {
	m = webrtc.MediaEngine{}
	m.RegisterDefaultCodecs()
	setting.SetEphemeralUDPPortRange(uint16(portMin), uint16(portMax))
	setting.SetLite(false)
	//setting.LoggerFactory = &myLogger{}
	api = webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithSettingEngine(setting))
}
func main() {
	logger, err := seelog.LoggerFromConfigAsFile("seelog.xml")
	if err == nil {
		logging.SetLogger(logger)
	}
	Init(20000, 40000)
	url := "ws://127.0.0.1:8088/janus"
	ctx, cancel := context.WithCancel(context.Background())

	vrb := NewVideoRoomBridge(ctx, url, 1234)

	err = vrb.Start()
	if err != nil {
		panic(err)
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		seelog.Info("user CTRL+C")
		break
	}
	vrb.Stop()
	<-time.After(time.Second * 1)

	cancel()
	<-time.After(time.Second * 3)
}

//VideoRoomBridge bridge ..
//pull stream from sub , and using pub look back to room.
type VideoRoomBridge struct {
	ctx  context.Context
	url  string
	conn *jwsapi.Connection
	sess *jwsapi.Session
	room uint64
	pub  *videoroom.Publisher
	sub  *videoroom.Subscriber
}

//NewVideoRoomBridge new bridge
func NewVideoRoomBridge(ctx context.Context, url string, room uint64) *VideoRoomBridge {

	vrb := &VideoRoomBridge{
		ctx:  ctx,
		url:  url,
		room: room,
		conn: jwsapi.NewConnection(ctx, url, 1),
	}

	return vrb
}

func (vrb *VideoRoomBridge) onAudioTrack(ctx context.Context, track *webrtc.Track) {

	sendTrack := vrb.pub.GetTrack(webrtc.RTPCodecTypeAudio)
	for {
		select {
		case <-vrb.ctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			packet, err := track.ReadRTP()
			if err != nil {
				return
			}
			if sendTrack != nil {
				packet.SSRC = sendTrack.SSRC()
				sendTrack.WriteRTP(packet)
			}

		}
	}
}

func (vrb *VideoRoomBridge) onVideoTrack(ctx context.Context, track *webrtc.Track) {
	sendTrack := vrb.pub.GetTrack(webrtc.RTPCodecTypeVideo)
	for {
		select {
		case <-vrb.ctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			packet, err := track.ReadRTP()
			if err != nil {
				return
			}
			if sendTrack != nil {
				packet.SSRC = sendTrack.SSRC()
				sendTrack.WriteRTP(packet)
			}

		}
	}
}

//Start ..
func (vrb *VideoRoomBridge) Start() error {

	sess, err := vrb.conn.Create()
	if err != nil {
		return errors.Wrap(err, "Create Janus Session")
	}
	vrb.sess = sess

	handle, err := sess.Attach("janus.plugin.videoroom")
	if err != nil {
		return errors.Wrap(err, "Create Janus Handle")
	}

	vrb.pub = videoroom.NewPublisher(vrb.ctx, api, handle, vrb.room)
	vrb.pub.Object().SetOption(jvideoroom.WithPublisherOptionNewPublisher(vrb.onNewPublisher))
	err = vrb.pub.Join(jwsapi.WithMessageOption("display", "newzai"))
	if err != nil {
		return errors.Wrap(err, "Janus Publisher Join")
	}

	err = vrb.pub.Publish(true, true)
	if err != nil {
		return errors.Wrap(err, "Janus Publish")
	}

	return nil
}

func (vrb *VideoRoomBridge) onNewPublisher(part jvideoroom.Participant) {
	if vrb.sub != nil {
		return
	}

	pubID := part.ID()

	handle, err := vrb.sess.Attach("janus.plugin.videoroom")
	if err != nil {
		return
	}
	handle.SetCallback(jwsapi.WithHandleHangup(vrb.onSubHangup))
	vrb.sub = videoroom.NewSubscriber(vrb.ctx, api, handle, vrb.room, pubID)
	vrb.sub.SetOption(videoroom.WithSubscriberAudioTrack(vrb.onAudioTrack)).
		SetOption(videoroom.WithSubscriberVideoTrack(vrb.onVideoTrack))

	err = vrb.sub.Start()
	if err != nil {
		vrb.sub = nil
	}

}

func (vrb *VideoRoomBridge) onUnpublish(id uint64) {
	if id == vrb.sub.Object().Feed() {
		vrb.sub.Leave()
		vrb.sub.Object().Handle().Detach()
		vrb.sub = nil
	}
}

func (vrb *VideoRoomBridge) onSubHangup(jwsapi.Message) {
	if vrb.sub != nil {
		vrb.sub.Leave()
		vrb.sub.Object().Handle().Detach()
	}
	vrb.sub = nil
}

//Stop destroy..
func (vrb *VideoRoomBridge) Stop() {
	vrb.sess.Destroy()
	vrb.sess = nil
	vrb.sub = nil
	vrb.pub = nil

}
