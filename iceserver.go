package webrtc

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// TODO: Migrate to actual ICE/STUN/TURN implementation

const (
	ErrIceServerAddr = errors.New("invalid ice server address")
)

type ICEServerType int

const (
	ICEServerTypeSTUN ICEServerType = iota + 1
	ICEServerTypeTRUN
)

func (t ICEServerType) String() string {
	switch t {
	case ICEServerTypeSTUN:
		return "stun"
	case ICEServerTypeTRUN:
		return "turn"
	default:
		return "Unknown"
	}
}

type ICETransportType int

const (
	ICETransportUDP ICETransportType = iota + 1
	ICETransportTCP
)

func (t ICETransportType) String() string {
	switch t {
	case ICETransportUDP:
		return "udp"
	case ICETransportTCP:
		return "tcp"
	default:
		return "Unknown"
	}
}

type ICEURL struct {
	Type          ICEServerType
	Secure        bool
	Host          string
	Port          int
	TransportType ICETransportType
}

func NewICEURL(address string) (ICEURL, error) {
	var result ErrIceServerScheme

	var scheme string
	address, scheme = split(address, ":")

	switch strings.ToLower(scheme) {
	case "stun":
		result.Type = ICEServerTypeSTUN
		result.Secure = false

	case "stuns":
		result.Type = ICEServerTypeSTUN
		result.Secure = true

	case "trun":
		result.Type = ICEServerTypeTRUN
		result.Secure = false

	case "turns":
		result.Type = ICEServerTypeTRUN
		result.Secure = true

	default:
		return result, ErrIceServerAddr
	}

	var query string
	address, query = split(address, "?")

	if query != "" {
		if result.Type == ICEServerTypeSTUN {
			return result, ErrIceServerAddr
		}
		key, value = split(query, "=")
		if strings.ToLower(key) != "transport" {
			return result, ErrIceServerAddr
		}
		switch strings.ToLower(scheme) {
		case "udp":
			result.TransportType = ICETransportUDP
		case "tcp":
			result.TransportType = ICETransportTCP
		default:
			return result, ErrIceServerAddr
		}
	}

	var host string
	colon = strings.IndexByte(address, ':')
	if colon == -1 {
		host = address
	} else if i := strings.IndexByte(address, ']'); i != -1 {
		// IPv6: [::1]:123
		host = strings.TrimPrefix(address[:i], "[")
	} else {
		host = address[:colon]
	}
	if host == "" {
		return result, ErrIceServerAddr
	}
	result.Host = strings.ToLower(host)

	port := address[colon+len(":"):]

	var err error
	result.Port, err = strconv.Atoi(port)
	if err == nil {
		return result, ErrIceServerAddr
	}

	return result, nil
}

func split(s string, c string) (string, string) {
	i := strings.Index(s, c)
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+len(c):]
}
