package sdp

import (
	"fmt"
	"time"
)

const (
	AttrKeyIdentity        = "identity"
	AttrKeyGroup           = "group"
	AttrKeySsrc            = "ssrc"
	AttrKeySsrcGroup       = "ssrc-group"
	AttrKeyMsidSemantic    = "msid-semantic"
	AttrKeyConnectionSetup = "setup"
	AttrKeyMID             = "mid"
	AttrKeyICELite         = "ice-lite"
	AttrKeyRtcpMux         = "rtcp-mux"
	AttrKeyRtcpRsize       = "rtcp-rsize"
)

const (
	SemanticTokenLipSynchronization     = "LS"
	SemanticTokenFlowIdentification     = "FID"
	SemanticTokenForwardErrorCorrection = "FEC"
	SemanticTokenWebRTCMediaStreams     = "WMS"
)

// API to match draft-ietf-rtcweb-jsep
// Move to webrtc or its own package?

// NewJSEPSessionDescription creates a new SessionDescription with
// some settings that are required by the JSEP spec.
func NewJSEPSessionDescription(fingerprint string, identity bool) *SessionDescription {
	desc := &SessionDescription{
		ProtocolVersion: 0,
		Origin: fmt.Sprintf(
			"- %d %d IN IP4 0.0.0.0",
			username,
			newSessionID(),
			time.Now().Unix(),
		),
		SessionName: "-",
		Timing:      []string{"0 0"},
		Attributes: []string{
			"ice-options:trickle",
		},
	}

	if identity {
		d.WithPropertyAttribute(AttrKeyIdentity)
	}
}

func (d *SessionDescription) WithPropertyAttribute(key string) *SessionDescription {
	d.Attributes = append(d.Attributes, key)
	return d
}

func (d *SessionDescription) WithValueAttribute(key, value string) *SessionDescription {
	d.Attributes = append(d.Attributes, fmt.Sprintf("%s:%s", key, value))
	return d
}

func (d *SessionDescription) WithMedia(md *MediaDescription) *SessionDescription {
	d.MediaDescriptions = append(d.MediaDescriptions, md)
	return d
}

// NewJSEPMediaDescription creates a new MediaDescription with
// some settings that are required by the JSEP spec.
func NewJSEPMediaDescription(codecPrefs []string) *MediaDescription {
	d := &MediaDescription{
		MediaName:      "video 9 UDP/TLS/RTP/SAVPF", // TODO: other transports?
		ConnectionData: "IN IP4 0.0.0.0",
		Attributes:     []string{},
	}
	return d
}

func (d *MediaDescription) WithPropertyAttribute(key string) *MediaDescription {
	d.Attributes = append(d.Attributes, key)
	return d
}

func (d *MediaDescription) WithValueAttribute(key, value string) *MediaDescription {
	d.Attributes = append(d.Attributes, fmt.Sprintf("%s:%s", key, value))
	return d
}

func (d *MediaDescription) WithICECredentials(username, password string) *MediaDescription {
	return d.
		WithValueAttribute("ice-ufrag", username).
		WithValueAttribute("ice-pwd", password)
}

func (d *MediaDescription) WithRTCRtpCodec(codec *RTCRtpCodec) {
	d.MediaName = fmt.Sprintf("%s %d", d.MediaName, codec.payloadType)
	rtpmap := fmt.Sprintf("%d %s/%d", codec.PayloadType, codec.Name, codec.ClockRate)
	if codec.Channels > 0 {
		rtpmap = rtpmap + fmt.Sprintf("/%d", codec.Channels)
	}
	d.WithValueAttribute("rtpmap", rtpmap)
	if codec.SdpFmtpLine != "" {
		d.WithValueAttribute("fmtp", fmt.Sprintf("%d %s", codec.PayloadType, codec.SdpFmtpLine))
	}
	return d
}

func (d *MediaDescription) WithMediaSource(ssrc, cname, streamLabel, label string) *MediaDescription {
	return d.
		WithValueAttribute("ssrc", fmt.Sprintf("%s cname:%s", ssrc, cname)). // Deprecated but not pased out?
		WithValueAttribute("ssrc", fmt.Sprintf("%s msid:%s", ssrc, streamLabel, label)).
		WithValueAttribute("ssrc", fmt.Sprintf("%s mslabel:%s", ssrc, streamLabel)). // Deprecated but not pased out?
		WithValueAttribute("ssrc", fmt.Sprintf("%s label:%s", ssrc, label))          // Deprecated but not pased out?
}

func (d *MediaDescription) WithCandidate(id, int, basePriority uint16, ip string, port int) *MediaDescription {
	return d.
		WithValueAttribute("candidate",
			fmt.Sprintf("udpcandidate %d udp %d %s %d typ host", id, basePriority, ip, port))
}
