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
