package webrtc

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/pions/webrtc/internal/dtls"
	"github.com/pions/webrtc/internal/network"
	"github.com/pions/webrtc/internal/sdp"
	"github.com/pions/webrtc/pkg/ice"
	"github.com/pions/webrtc/pkg/rtp"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type RTCPeerConnectionState int

const (
	RTCPeerConnectionStateNew RTCPeerConnectionState = iota + 1
	RTCPeerConnectionStateConnecting
	RTCPeerConnectionStateConnected
	RTCPeerConnectionStateDisconnected
	RTCPeerConnectionStateFailed
	RTCPeerConnectionStateClosed
)

func (t RTCPeerConnectionState) String() string {
	switch t {
	case RTCPeerConnectionStateNew:
		return "new"
	case RTCPeerConnectionStateConnecting:
		return "connecting"
	case RTCPeerConnectionStateConnected:
		return "connected"
	case RTCPeerConnectionStateDisconnected:
		return "disconnected"
	case RTCPeerConnectionStateFailed:
		return "failed"
	case RTCPeerConnectionStateClosed:
		return "closed"
	default:
		return "Unknown"
	}
}

// RTCPeerConnection represents a WebRTC connection between itself and a remote peer
type RTCPeerConnection struct {
	// ICE
	OnICEConnectionStateChange func(iceConnectionState ice.ConnectionState)

	config RTCConfiguration
	tlscfg *dtls.TLSCfg

	iceUfrag string
	icePwd   string
	iceState ice.ConnectionState

	portsLock sync.RWMutex
	ports     []*network.Port

	// Signaling
	pendingLocalDescription *RTCSessionDescription
	currentLocalDescription *RTCSessionDescription
	LocalDescription        *sdp.SessionDescription

	pendingRemoteDescription *RTCSessionDescription
	currentRemoteDescription *RTCSessionDescription
	remoteDescription        *sdp.SessionDescription

	// Tracks
	localTracks []*sdp.SessionBuilderTrack

	idpLoginUrl *string

	IsClosed          bool
	NegotiationNeeded bool

	lastOffer  string
	lastAnswer string

	signalingState     RTCSignalingState
	iceGatheringState  RTCIceGatheringState
	iceConnectionState RTCIceConnectionState
	connectionState    RTCPeerConnectionState

	// Media
	rtpTransceivers []*RTCRtpTransceiver
	Ontrack         func(*Track)
}

// New creates a new RTCPeerConfiguration with the provided configuration
func New(config RTCConfiguration) (*RTCPeerConnection, error) {

	r := &RTCPeerConnection{
		config:             config,
		signalingState:     RTCSignalingStateStable,
		iceGatheringState:  RTCIceGatheringStateNew,
		iceConnectionState: RTCIceConnectionStateNew,
		connectionState:    RTCPeerConnectionStateNew,
	}
	err := r.setConfiguration(config, false)
	if err != nil {
		return err
	}

	r.tlscfg = dtls.NewTLSCfg()

	// TODO: Initialize ICE Agent

	return r, nil
}

func (r *RTCPeerConnection) SetConfiguration(config RTCConfiguration) error {
	current := r.config

	if current.PeerIdentity != "" &&
		config.PeerIdentity != "" &&
		config.PeerIdentity != current.PeerIdentity {
		return &InvalidModificationError{Err: ErrModPeerIdentity}
	}

	if len(current.Certificates) > 0 &&
		len(config.Certificates) > 0 {
		if len(config.Certificates) != len(current.Certificates) {
			return &InvalidModificationError{Err: ErrModCertificates}
		}
		for i, cert := range config.Certificates {
			if !current.Certificates[i].Equals(cert) {
				return &InvalidModificationError{Err: ErrModCertificates}
			}
		}
	}

	now := time.Now()
	for _, cert := range config.certificates {
		if now.After(cert.expires) {
			return nil, &InvalidAccessError{Err: ErrCertificateExpired}
		}
		// TODO: Check certificate 'origin'
	}

	if config.BundlePolicy != current.BundlePolicy {
		return &InvalidModificationError{Err: ErrModRtcpMuxPolicy}
	}

	if config.RtcpMuxPolicy != current.RtcpMuxPolicy {
		return &InvalidModificationError{Err: ErrModRtcpMuxPolicy}
	}

	if r.LocalDescription != nil &&
		config.IceCandidatePoolSize != current.IceCandidatePoolSize {
		return &InvalidModificationError{Err: ErrModIceCandidatePoolSize}
	}

	if len(config.ICEServers) > 0 {
		for _, server := range config.ICEServers {
			for _, url := range server.URLs {
				iceurl, err := NewICEURL(url)
				if err != nil {
					return &SyntaxError{Err: err}
				}

				passCred, isPass := x.(string)
				oauthCred, isOauth := x.(RTCOAuthCredential)
				noPass := !isPass && !isOauth

				if iceurl.Type == ICEServerTypeTRUN {
					if server.Username == "" ||
						noPass {
						return &InvalidAccessError{Err: ErrNoTrunCred}
					}
					if server.CredentialType == RTCIceCredentialTypePassword &&
						!isPass {
						return &InvalidAccessError{Err: ErrTrunCred}
					}
					if server.CredentialType == RTCIceCredentialTypeOauth &&
						!isOauth {
						return &InvalidAccessError{Err: ErrTrunCred}
					}
				}

				// TODO: Add to ICE agent valid server list
			}
		}
	}

	r.config = config
}

// Public

func (r *RTCPeerConnection) SetIdentityProvider(provider string) error {
	panic("TODO SetIdentityProvider")
}

// Close ends the RTCPeerConnection
func (r *RTCPeerConnection) Close() error {
	r.portsLock.Lock()
	defer r.portsLock.Unlock()

	// Walk all ports remove and close them
	for _, p := range r.ports {
		if err := p.Close(); err != nil {
			return err
		}
	}
	r.ports = nil
	return nil
}

// Private
func (r *RTCPeerConnection) generateChannel(ssrc uint32, payloadType uint8) (buffers chan<- *rtp.Packet) {
	if r.Ontrack == nil {
		return nil
	}

	// TODO: Identify RTCRtpCodec instead

	var codec TrackType
	ok, codecStr := sdp.GetCodecForPayloadType(payloadType, r.remoteDescription)
	if !ok {
		fmt.Printf("No codec could be found in RemoteDescription for payloadType %d \n", payloadType)
		return nil
	}

	switch codecStr {
	case "VP8":
		codec = VP8
	case "VP9":
		codec = VP9
	case "opus":
		codec = Opus
	case "H264":
		codec = H264
	default:
		fmt.Printf("Codec %s in not supported by pion-WebRTC \n", codecStr)
		return nil
	}

	bufferTransport := make(chan *rtp.Packet, 15)
	go r.Ontrack(codec, bufferTransport)
	return bufferTransport
}

// Private
func (r *RTCPeerConnection) iceStateChange(p *network.Port) {
	updateAndNotify := func(newState ice.ConnectionState) {
		if r.OnICEConnectionStateChange != nil && r.iceState != newState {
			r.OnICEConnectionStateChange(newState)
		}
		r.iceState = newState
	}

	if p.ICEState == ice.Failed {
		if err := p.Close(); err != nil {
			fmt.Println(errors.Wrap(err, "Failed to close Port when ICE went to failed"))
		}

		r.portsLock.Lock()
		defer r.portsLock.Unlock()
		for i := len(r.ports) - 1; i >= 0; i-- {
			if r.ports[i] == p {
				r.ports = append(r.ports[:i], r.ports[i+1:]...)
			}
		}

		if len(r.ports) == 0 {
			updateAndNotify(ice.Disconnected)
		}
	} else {
		updateAndNotify(ice.Connected)
	}
}
