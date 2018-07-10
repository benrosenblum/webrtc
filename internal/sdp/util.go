package sdp

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// SessionBuilderTrack represents a single track in a SessionBuilder
type SessionBuilderTrack struct {
	SSRC    uint32
	IsAudio bool
}

// SessionBuilder provides an easy way to build an SDP for an RTCPeerConnection
type SessionBuilder struct {
	IceUsername, IcePassword, Fingerprint string

	Candidates []string

	Tracks []*SessionBuilderTrack
}

type ConnectionRole int

const (
	ConnectionRoleActive ConnectionRole = iota + 1
	ConnectionRolePassive
	ConnectionRoleActpass
	ConnectionRoleHoldconn
)

func (t ConnectionRole) String() string {
	switch t {
	case ConnectionRoleActive:
		return "active"
	case ConnectionRolePassive:
		return "passive"
	case ConnectionRoleActpass:
		return "actpass"
	case ConnectionRoleHoldconn:
		return "holdconn"
	default:
		return "Unknown"
	}
}

// BaseSessionDescription generates a default SDP response that is ice-lite, initiates the DTLS session and
// supports VP8, VP9, H264 and Opus
func BaseSessionDescription(b *SessionBuilder) *SessionDescription {
	addMediaCandidates := func(m *MediaDescription) *MediaDescription {
		m.Attributes = append(m.Attributes, b.Candidates...)
		m.Attributes = append(m.Attributes, "end-of-candidates")
		return m
	}

	audioMediaDescription := &MediaDescription{
		MediaName:      "audio 9 RTP/SAVPF 111",
		ConnectionData: "IN IP4 127.0.0.1",
		Attributes: []string{
			"setup:active",
			"mid:audio",
			"sendrecv",
			"ice-ufrag:" + b.IceUsername,
			"ice-pwd:" + b.IcePassword,
			"ice-lite",
			"fingerprint:sha-256 " + b.Fingerprint,
			"rtcp-mux",
			"rtcp-rsize",
			"rtpmap:111 opus/48000/2",
			"fmtp:111 minptime=10;useinbandfec=1",
		},
	}

	videoMediaDescription := &MediaDescription{
		MediaName:      "video 9 RTP/SAVPF 96 98 100",
		ConnectionData: "IN IP4 127.0.0.1",
		Attributes: []string{
			"setup:active",
			"mid:video",
			"sendrecv",
			"ice-ufrag:" + b.IceUsername,
			"ice-pwd:" + b.IcePassword,
			"ice-lite",
			"fingerprint:sha-256 " + b.Fingerprint,
			"rtcp-mux",
			"rtcp-rsize",
			"rtpmap:96 VP8/90000",
			"rtpmap:98 VP9/90000",
			"rtpmap:100 H264/90000",
			"fmtp:100 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		},
	}

	mediaStreamsAttribute := "msid-semantic: WMS"
	for i, track := range b.Tracks {
		var attributes *[]string
		if track.IsAudio {
			attributes = &audioMediaDescription.Attributes
		} else {
			attributes = &videoMediaDescription.Attributes
		}
		appendAttr := func(attr string) {
			*attributes = append(*attributes, attr)
		}

		appendAttr("ssrc:" + fmt.Sprint(track.SSRC) + " cname:pion" + strconv.Itoa(i))
		appendAttr("ssrc:" + fmt.Sprint(track.SSRC) + " msid:pion" + strconv.Itoa(i) + " pion" + strconv.Itoa(i))
		appendAttr("ssrc:" + fmt.Sprint(track.SSRC) + " mslabel:pion" + strconv.Itoa(i))
		appendAttr("ssrc:" + fmt.Sprint(track.SSRC) + " label:pion" + strconv.Itoa(i))

		mediaStreamsAttribute += " pion" + strconv.Itoa(i)
	}

	sessionID := strconv.FormatUint(newSessionID(), 10)
	return &SessionDescription{
		ProtocolVersion: 0,
		Origin:          "pion-webrtc " + sessionID + " 2 IN IP4 0.0.0.0",
		SessionName:     "-",
		Timing:          []string{"0 0"},
		Attributes: []string{
			"group:BUNDLE audio video",
			mediaStreamsAttribute,
		},
		MediaDescriptions: []*MediaDescription{
			addMediaCandidates(audioMediaDescription),
			addMediaCandidates(videoMediaDescription),
		},
	}
}

func newSessionID() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}

// Codec
type Codec struct {
	PayloadType        uint8
	Name               string
	ClockRate          uint32
	EncodingParameters string
	Fmtp               string
}

func (c Codec) String() string {
	return fmt.Sprintf("%d %s/%d/%s", c.PayloadType, c.Name, c.ClockRate, c.EncodingParameters)
}

// GetCodecForPayloadType scans the SessionDescription for the given payloadType and returns the codec
func (sd *SessionDescription) GetCodecForPayloadType(payloadType uint8) (Codec, error) {
	codec := Codec{
		PayloadType: payloadType,
	}

	found := false
	payloadTypeString := strconv.Itoa(int(payloadType))
	rtpmapPrefix := "rtpmap:" + payloadTypeString
	fmtpPrefix := "fmtp:" + payloadTypeString

	for _, m := range sd.MediaDescriptions {
		for _, a := range m.Attributes {
			if strings.HasPrefix(a, rtpmapPrefix) {
				// a=rtpmap:<payload type> <encoding name>/<clock rate> [/<encoding parameters>]
				split := strings.Split(a, " ")
				if len(split) == 2 {
					split := strings.Split(split[1], "/")
					codec.Name = split[0]
					parts := len(split)
					if parts > 1 {
						rate, err := strconv.Atoi(parts[1])
						if err != nil {
							return codec, error
						}
						codec.ClockRate = uint32(rate)
					}
					if parts > 2 {
						codec.EncodingParameters = parts[2]
					}
				}
			} else if strings.HasPrefix(a, prefix) {
				// a=fmtp:<format> <format specific parameters>
				split := strings.Split(a, " ")
				if len(split) == 2 {
					codec.Fmtp = split[1]
				}
			}
		}
		if found {
			return codec, nil
		}
	}
	return codec, errors.New("payload type not found")
}
