package webrtc

import (
	"strconv"

	"github.com/pions/webrtc/internal/sdp"
	"github.com/pions/webrtc/pkg/rtp"
	"github.com/pions/webrtc/pkg/rtp/codecs"
	"github.com/pkg/errors"
)

var rtcMediaEngine = &mediaEngine{}

func init() {
	RegisterCodec(NewRTCRtpOpusCodec(PayloadTypeOpus, 48000, 2))
	RegisterCodec(NewRTCRtpVP8Codec(PayloadTypeVP8, 90000))
	RegisterCodec(NewRTCRtpH264Codec(PayloadTypeH264, 90000))
}

// RegisterCodec allows an implementation to provide it's own codec.
func RegisterCodec(codec *RTCRtpCodec) {
	rtcMediaEngine.RegisterCodec(codec)
}

func (m *mediaEngine) ClearCodecs() {
	rtcMediaEngine.ClearCodecs()
}

type mediaEngine struct {
	codecs []*RTCRtpCodec
}

func (m *mediaEngine) RegisterCodec(codec *RTCRtpCodec) {
	m.codecs = append(m.codecs, codec)
}

func (m *mediaEngine) ClearCodecs() {
	m.Codecs = nil
}

func (m *mediaEngine) getCodec(payloadType uint8) (*RTCRtpCodec, error) {
	for _, codec := range m.codecs {
		if codec.PayloadType == payloadType {
			return codec, nil
		}
	}
	return nil, errors.New("Codec not found")
}

func (m *mediaEngine) getCodecSDP(sdpCodec sdp.Codec) (*RTCRtpCodec, error) {
	for _, codec := range m.codecs {
		if codec.Name == sdpCodec.Name &&
			codec.ClockRate == sdpCodec.ClockRate &&
			(sdpCodec.EncodingParameters == "" ||
				strconv.ItoA(int(codec.Channels)) == sdpCodec.EncodingParameters) &&
			codec.SdpFmtpLine == sdpCodec.Fmtp { // TODO: Protocol specific matching?
			return codec, nil
		}
	}
	return nil, errors.New("Codec not found")

}

// NewRTCRtpOpusCodec is a helper to create an Opus codec
func NewRTCRtpOpusCodec(payloadType uint8, clockrate UnsignedLong, channels UnsignedShort) *RTCRtpCodec {
	NewRTCRtpCodec(RTCRtpCodecKindAudio,
		"opus",
		clockrate,
		channels,
		"minptime=10;useinbandfec=1",
		payloadType,
		&codecs.OpusPayloader{})
}

// NewRTCRtpVP8Codec is a helper to create an VP8 codec
func NewRTCRtpVP8Codec(payloadType uint8, clockrate UnsignedLong) *RTCRtpCodec {
	NewRTCRtpCodec(RTCRtpCodecTypeVideo,
		"VP8",
		clockrate,
		0,
		"",
		payloadType,
		&codecs.VP8Payloader{})
}

// NewRTCRtpH264Codec is a helper to create an H264 codec
func NewRTCRtpH264Codec(payloadType uint8, clockrate UnsignedLong) *RTCRtpCodec {
	NewRTCRtpCodec(RTCRtpCodecTypeVideo,
		"H264",
		clockrate,
		0,
		"level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		payloadType,
		&codecs.H264Payloader{})
}

type RTCRtpCodecType int

const (
	RTCRtpCodecTypeAudio RTCRtpCodecType = iota + 1
	RTCRtpCodecTypeVideo
)

func (t RTCRtpCodecType) String() string {
	switch t {
	case RTCRtpCodecKindAudio:
		return "audio"
	case RTCRtpCodecKindVideo:
		return "video"
	default:
		return "Unknown"
	}
}

const (
	PayloadTypeOpus = 111
	PayloadTypeVP8  = 96
	PayloadTypeH264 = 100
)

type RTCRtpCodec struct {
	RTCRtpCodecCapability
	Type        RTCRtpCodecType
	Name        string
	PayloadType uint8
	Payloader   rtp.Payloader
}

func NewRTCRtpCodec(
	typ RTCRtpCodecType,
	name string,
	clockrate UnsignedLong,
	channels UnsignedShort,
	fmtp string,
	payloadType uint8,
	payloader rtp.Payloader,
) *RTCRtpCodec {
	return &RTCRtpCodec{
		RTCRtpCodecCapability{
			MimeType:    typ.String() + "/" + name,
			ClockRate:   clockrate,
			Channels:    channels,
			SdpFmtpLine: fmtp,
		},
		PayloadType: payloadType,
		Payloader:   payloader,
		Type:        typ,
		Name:        name,
	}
}

type RTCRtpCodecCapability struct {
	MimeType    string
	ClockRate   UnsignedLong
	Channels    UnsignedShort
	SdpFmtpLine string
}

type RTCRtpHeaderExtensionCapability struct {
	Uri string
}

type RTCRtpCapabilities struct {
	Codecs           []RTCRtpCodecCapability
	HeaderExtensions []RTCRtpHeaderExtensionCapability
}
