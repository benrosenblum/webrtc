package webrtc

import (
	"strings"
	"time"
)

type RTCICECredentialType int

const (
	// RTCICECredentialTypePassword describes username+pasword based credentials
	RTCICECredentialTypePassword RTCICECredentialType = iota + 1
	// RTCICECredentialTypeOauth describes token based credentials
	RTCICECredentialTypeOauth
)

func (t RTCICECredentialType) String() string {
	switch t {
	case RTCICECredentialTypePassword:
		return "password"
	case RTCICECredentialTypeOauth:
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
	CredentialType RTCICECredential
}

type RTCICECredential interface{}

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

type RTCICETransportPolicy int

const (
	Relay RTCICETransportPolicy = iota + 1
	All
)

func (t RTCICETransportPolicy) String() string {
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
	ICEServers           []RTCICEServer // An array of RTCICEServer objects, each describing one server which may be used by the ICE agent; these are typically STUN and/or TURN servers. If this isn't specified, the ICE agent may choose to use its own ICE servers; otherwise, the connection attempt will be made with no STUN or TURN server available, which limits the connection to local peers.
	ICETransportPolicy   RTCICETransportPolicy
	BundlePolicy         RTCBundlePolicy
	RtcpMuxPolicy        RTCRtcpMuxPolicy
	PeerIdentity         string
	Certificates         []RTCCert
	ICECandidatePoolSize Octet
}
