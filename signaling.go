package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pions/webrtc/internal/network"
	"github.com/pions/webrtc/internal/sdp"
	"github.com/pions/webrtc/internal/util"
	"github.com/pions/webrtc/pkg/ice"
	"github.com/pkg/errors"
)

/*
                      setRemote(OFFER)               setLocal(PRANSWER)
                          /-----\                               /-----\
                          |     |                               |     |
                          v     |                               v     |
           +---------------+    |                +---------------+    |
           |               |----/                |               |----/
           |  have-        | setLocal(PRANSWER)  | have-         |
           |  remote-offer |------------------- >| local-pranswer|
           |               |                     |               |
           |               |                     |               |
           +---------------+                     +---------------+
                ^   |                                   |
                |   | setLocal(ANSWER)                  |
  setRemote(OFFER)  |                                   |
                |   V                  setLocal(ANSWER) |
           +---------------+                            |
           |               |                            |
           |               |<---------------------------+
           |    stable     |
           |               |<---------------------------+
           |               |                            |
           +---------------+          setRemote(ANSWER) |
                ^   |                                   |
                |   | setLocal(OFFER)                   |
  setRemote(ANSWER) |                                   |
                |   V                                   |
           +---------------+                     +---------------+
           |               |                     |               |
           |  have-        | setRemote(PRANSWER) |have-          |
           |  local-offer  |------------------- >|remote-pranswer|
           |               |                     |               |
           |               |----\                |               |----\
           +---------------+    |                +---------------+    |
                          ^     |                               ^     |
                          |     |                               |     |
                          \-----/                               \-----/
                      setLocal(OFFER)               setRemote(PRANSWER)
*/

type RTCSignalingState int

const (
	RTCSignalingStateStable RTCSignalingState = iota + 1
	RTCSignalingStateHaveLocalOffer
	RTCSignalingStateHaveRemoteOffer
	RTCSignalingStateHaveLocalPranswer
	RTCSignalingStateHaveRemotePranswer
	RTCSignalingStateClosed
)

func (t RTCSignalingState) String() string {
	switch t {
	case RTCSignalingStateStable:
		return "stable"
	case RTCSignalingStateHaveLocalOffer:
		return "have-local-offer"
	case RTCSignalingStateHaveRemoteOffer:
		return "have-remote-offer"
	case RTCSignalingStateHaveLocalPranswer:
		return "have-local-pranswer"
	case RTCSignalingStateHaveRemotePranswer:
		return "have-remote-pranswer"
	case RTCSignalingStateClosed:
		return "closed"
	default:
		return "Unknown"
	}
}

type RTCSdpType int

const (
	Offer RTCSdpType = iota + 1
	Pranswer
	Answer
	Rollback
)

func (t RTCSdpType) String() string {
	switch t {
	case Offer:
		return "offer"
	case Pranswer:
		return "pranswer"
	case Answer:
		return "answer"
	case Rollback:
		return "rollback"
	default:
		return "Unknown"
	}
}

type RTCSessionDescription struct {
	Typ RTCSdpType
	Sdp string
}

// SetRemoteDescription sets the SessionDescription of the remote peer
func (r *RTCPeerConnection) SetRemoteDescription(desc RTCSessionDescription) error {
	if r.remoteDescription != nil {
		return errors.Errorf("remoteDescription is already defined, SetRemoteDescription can only be called once")
	}

	r.remoteDescription = &RTCSessionDescription{
		&sdp.SessionDescription{}}
	return r.remoteDescription.Unmarshal(desc.Sdp)
}

type RTCOfferOptions struct {
	VoiceActivityDetection bool
	IceRestart             bool
}

// CreateOffer starts the RTCPeerConnection and generates the localDescription
func (r *RTCPeerConnection) CreateOffer(options *RTCOfferOptions) (RTCSessionDescription, error) {
	if options != nil {
		panic("TODO handle options")
	}
	if r.IsClosed {
		return &InvalidStateError{Err: ErrConnectionClosed}
	}
	useIdentity := r.idpLoginUrl != nil
	if useIdentity {
		panic("TODO handle identity provider")
	}

	d := sdp.BaseDescription(
		r.tlscfg.Fingerprint(),
		useIdentity,
	)

	// mediaStreamsAttribute := ": WMS"

	for _, tranceiver := range r.rtpTransceivers {
		if tranceiver.Sender == nil ||
			tranceiver.Sender.Track == nil {
			continue
		}
		codec, _ := rtcMediaEngine.getCodec(tranceiver.Sender.Track.PayloadType)
		media := sdp.NewJSEPMediaDescription().
			WithValueAttribute("setup", sdp.ConnectionRoleActive.String()). // TODO: Support other connection types
			WithValueAttribute("mid", tranceiver.Mid).
			WithPropertyAttribute(tranceiver.Direction.String()).
			WithICECredentials(util.RandSeq(16), util.RandSeq(32)). // TODO: get credendials form ICE agent
			WithPropertyAttribute("ice-lite").                      // TODO: get ICE type from ICE Agent
			WithPropertyAttribute("rtcp-mux").                      // TODO: support RTCP fallback
			WithPropertyAttribute("rtcp-rsize").                    // TODO: Support Reduced-Size RTCP?
			WithRTCRtpCodec(codec)

			// TODO: Add ssrc lines

		d.WithMedia(media)
	}

	// d.WithValueAttribute("msid-semantic") // ??
}

type RTCAnswerOptions struct {
	VoiceActivityDetection bool
}

// CreateAnswer starts the RTCPeerConnection and generates the localDescription
func (r *RTCPeerConnection) CreateAnswer(options *RTCOfferOptions) error {
	if options != nil {
		panic("TODO handle options")
	}
	if r.IsClosed {
		return &InvalidStateError{Err: ErrConnectionClosed}
	}
	useIdentity := r.idpLoginUrl != nil
	if useIdentity {
		panic("TODO handle identity provider")
	}

	d := sdp.BaseDescription(
		r.tlscfg.Fingerprint(),
		useIdentity,
	)

	d.WithMedia(getAnswerMedia(RTCRtpCodecTypeAudio))
	d.WithMedia(getAnswerMedia(RTCRtpCodecTypeVideo))

	// TODO: Move to ICE agent initialisation

	// r.LocalDescription = sdp.BaseSessionDescription(&sdp.SessionBuilder{
	// 	IceUsername: r.iceUfrag,
	// 	IcePassword: r.icePwd,
	// 	Fingerprint: r.tlscfg.Fingerprint(),
	// 	Candidates:  candidates,
	// 	Tracks:      r.localTracks,
	// })

	return nil
}

func (r *RTCPeerConnection) getAnswerMedia(typ RTCRtpCodecType) *sdp.MediaDescription {
	media := sdp.NewJSEPMediaDescription().
		WithValueAttribute("setup", sdp.ConnectionRoleActive.String()). // TODO: Support other connection types
		WithValueAttribute("mid", typ.String()).
		WithPropertyAttribute(tranceiver.Direction.String()).
		WithICECredentials(util.RandSeq(16), util.RandSeq(32)). // TODO: get credendials form ICE agent
		WithPropertyAttribute("ice-lite").                      // TODO: get ICE type from ICE Agent
		WithPropertyAttribute("rtcp-mux").                      // TODO: support RTCP fallback
		WithPropertyAttribute("rtcp-rsize")                     // TODO: Support Reduced-Size RTCP?

	for _, codec := range rtcMediaEngine.getCodecsByKind(typ) {
		media.WithRTCRtpCodec(codec)
	}

	// Gather candidates per MediaDescription
	// TODO: Refactor to new API
	r.iceUfrag = util.RandSeq(16)
	r.icePwd = util.RandSeq(32)

	r.portsLock.Lock()
	defer r.portsLock.Unlock()

	candidates := []string{}
	basePriority := uint16(rand.Uint32() & (1<<16 - 1))
	for id, c := range ice.HostInterfaces() {
		port, err := network.NewPort(c+":0", []byte(r.icePwd), r.tlscfg, r.generateChannel, r.iceStateChange)
		if err != nil {
			return err
		}
		candidates = append(candidates, fmt.Sprintf("candidate:udpcandidate %d udp %d %s %d typ host", id+1, basePriority, c, port.ListeningAddr.Port))
		basePriority = basePriority + 1
		r.ports = append(r.ports, port)
	}
}
