package webrtc

import (
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
	RTCSdpTypeOffer RTCSdpType = iota + 1
	RTCSdpTypePranswer
	RTCSdpTypeAnswer
	RTCSdpTypeRollback
)

func (t RTCSdpType) String() string {
	switch t {
	case RTCSdpTypeOffer:
		return "offer"
	case RTCSdpTypePranswer:
		return "pranswer"
	case RTCSdpTypeAnswer:
		return "answer"
	case RTCSdpTypeRollback:
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
		useIdentity).
		WithValueAttribute(sdp.AttrKeyGroup, "BUNDLE audio video") // TODO: Support

	var streamlabels string
	for i, tranceiver := range r.rtpTransceivers {
		if tranceiver.Sender == nil ||
			tranceiver.Sender.Track == nil {
			continue
		}
		track := tranceiver.Sender.Track
		cname := "poins"      // TODO: Support RTP streams synchronisation
		steamlabel := "poins" // TODO: Support steam labels
		codec, _ := rtcMediaEngine.getCodec(track.PayloadType)
		media := sdp.NewJSEPMediaDescription().
			WithValueAttribute(sdp.AttrKeyConnectionSetup, sdp.ConnectionRoleActive.String()). // TODO: Support other connection types
			WithValueAttribute(sdp.AttrKeyMID, tranceiver.Mid).
			WithPropertyAttribute(tranceiver.Direction.String()).
			WithICECredentials(util.RandSeq(16), util.RandSeq(32)). // TODO: get credendials form ICE agent
			WithPropertyAttribute(sdp.AttrKeyICELite).              // TODO: get ICE type from ICE Agent
			WithPropertyAttribute(sdp.AttrKeyRtcpMux).              // TODO: support RTCP fallback
			WithPropertyAttribute(sdp.AttrKeyRtcpRsize).            // TODO: Support Reduced-Size RTCP?
			WithRTCRtpCodec(codec).
			WithMediaSource(track.Ssrc, cname, steamlabel, track.Label)
		err := r.addICECandidates(d)
		if err != nil {
			return err
		}
		streamlabels := streamlabels + " " + steamlabel

		d.WithMedia(media)
	}

	d.WithValueAttribute(sdp.AttrKeyMsidSemantic, " "+sdp.SemanticTokenWebRTCMediaStreams+streamlabels)

	return RTCSessionDescription{
		Typ: RTCSdpTypeOffer,
		Sdp: peerConnection.LocalDescription.Marshal(),
	}, nil
}

type RTCAnswerOptions struct {
	VoiceActivityDetection bool
}

// CreateAnswer starts the RTCPeerConnection and generates the localDescription
func (r *RTCPeerConnection) CreateAnswer(options *RTCOfferOptions) (RTCSessionDescription, error) {
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
	d.WithValueAttribute(sdp.AttrKeyMsidSemantic, " "+sdp.SemanticTokenWebRTCMediaStreams+" poins")

	return RTCSessionDescription{
		Typ: RTCSdpTypeAnswer,
		Sdp: peerConnection.LocalDescription.Marshal(),
	}, nil
}

func (r *RTCPeerConnection) getAnswerMedia(typ RTCRtpCodecType) *sdp.MediaDescription {

	cname := "poins"      // TODO: Support RTP streams synchronisation
	steamlabel := "poins" // TODO: Support steam labels

	media := sdp.NewJSEPMediaDescription().
		WithValueAttribute(sdp.AttrKeyConnectionSetup, sdp.ConnectionRoleActive.String()). // TODO: Support other connection types
		WithValueAttribute(sdp.AttrKeyMID, typ.String()).
		WithPropertyAttribute(RTCRtpTransceiverDirectionSendrecv.String()).
		WithICECredentials(util.RandSeq(16), util.RandSeq(32)). // TODO: get credendials form ICE agent
		WithPropertyAttribute(sdp.AttrKeyICELite).              // TODO: get ICE type from ICE Agent
		WithPropertyAttribute(sdp.AttrKeyRtcpMux).              // TODO: support RTCP fallback
		WithPropertyAttribute(sdp.AttrKeyRtcpRsize)             // TODO: Support Reduced-Size RTCP?

	for _, codec := range rtcMediaEngine.getCodecsByKind(typ) {
		media.WithRTCRtpCodec(codec)
	}

	// media.WithMediaSource(track.Ssrc, cname, steamlabel, track.Label) // TODO: figure out what track to add.

	err := r.addICECandidates(d)
	if err != nil {
		return err
	}
}

func (r *RTCPeerConnection) addICECandidates(d *sdp.MediaDescription) error {
	r.portsLock.Lock()
	defer r.portsLock.Unlock()

	basePriority := uint16(rand.Uint32() & (1<<16 - 1))
	for id, c := range ice.HostInterfaces() {
		port, err := network.NewPort(c+":0", []byte(r.icePwd), r.tlscfg, r.generateChannel, r.iceStateChange)
		if err != nil {
			return err
		}

		d.WithCandidate(
			id+1,
			basePriority,
			c,
			port.ListeningAddr.Port,
		)

		basePriority = basePriority + 1
		r.ports = append(r.ports, port)
	}
	d.WithPropertyAttribute("end-of-candidates") // TODO: Support full trickle-ice
}
