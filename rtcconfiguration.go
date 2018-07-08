package webrtc

import (
	"strings"
	"time"
)

type RTCIceCredentialType int

const (
	// RTCIceCredentialTypePassword describes username+pasword based credentials
	RTCIceCredentialTypePassword ICECredentialType = iota + 1
	// RTCIceCredentialTypeOauth describes token based credentials
	RTCIceCredentialTypeOauth
)

func (t RTCIceCredentialType) String() string {
	switch t {
	case RTCIceCredentialTypePassword:
		return "password"
	case RTCIceCredentialTypeOauth:
		return "oauth"
	default:
		return "Unknown"
	}
}

type RTCCertificate struct {
	expires time.Time
	// TODO: Finish during DTLS implementation
}

func (c RTCCertificate) Equals(other RTCCertificate) bool {
	return c.expires == other.expires
}

// RTCICEServer describes a single ICE server, as well as required credentials
type RTCICEServer struct {
	URLs           []string
	Username       string
	Credential     string
	CredentialType RTCIceCredential
}

type RTCIceCredential interface{}

type RTCOAuthCredential struct {
	MacKey      string
	AccessToken string
}

func (c RTCICEServer) serverType() RTCServerType {
	for _, url := range c.URLs {
		if strings.HasPrefix(url, "stun:") {
			return RTCServerTypeSTUN
		}
		if strings.HasPrefix(url, "turn:") {
			return RTCServerTypeTURN
		}
	}
	return RTCServerTypeUnknown
}

type RTCIceTransportPolicy int

const (
	Relay RTCIceTransportPolicy = iota + 1
	All
)

func (t RTCIceTransportPolicy) String() string {
	switch t {
	case Relay:
		return "relay"
	case All:
		return "all"
	default:
		return "Unknown"
	}
}

type RTCBundlePolicy int

const (
	RTCRtcpMuxPolicyBalanced RTCBundlePolicy = iota + 1
	RTCRtcpMuxPolicyMaxCompat
	RTCRtcpMuxPolicyMaxBundle
)

func (t RTCBundlePolicy) String() string {
	switch t {
	case RTCRtcpMuxPolicyBalanced:
		return "balanced"
	case RTCRtcpMuxPolicyMaxCompat:
		return "max-compat"
	case RTCRtcpMuxPolicyMaxBundle:
		return "max-bundle"
	default:
		return "Unknown"
	}
}

type RTCRtcpMuxPolicy int

const (
	Negotiate RTCRtcpMuxPolicy = iota + 1
	Require
)

func (t RTCRtcpMuxPolicy) String() string {
	switch t {
	case Negotiate:
		return "negotiate"
	case Require:
		return "require"
	default:
		return "Unknown"
	}
}

// RTCConfiguration contains RTCPeerConfiguration options
type RTCConfiguration struct {
	ICEServers           []RTCICEServer // An array of RTCIceServer objects, each describing one server which may be used by the ICE agent; these are typically STUN and/or TURN servers. If this isn't specified, the ICE agent may choose to use its own ICE servers; otherwise, the connection attempt will be made with no STUN or TURN server available, which limits the connection to local peers.
	IceTransportPolicy   RTCIceTransportPolicy
	BundlePolicy         RTCBundlePolicy
	RtcpMuxPolicy        RTCRtcpMuxPolicy
	PeerIdentity         string
	Certificates         []RTCCert
	IceCandidatePoolSize Octet
}
