package webrtc

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// TODO: Migrate address parsing to STUN/TURN packages?

const (
	ErrServerAddr = errors.New("invalid ice server address")
)

type ServerType int

const (
	ServerTypeSTUN ServerType = iota + 1
	ServerTypeTURN
)

func (t ServerType) String() string {
	switch t {
	case ServerTypeSTUN:
		return "stun"
	case ServerTypeTURN:
		return "turn"
	default:
		return "Unknown"
	}
}

type TransportType int

const (
	TransportUDP TransportType = iota + 1
	TransportTCP
)

func (t TransportType) String() string {
	switch t {
	case TransportUDP:
		return "udp"
	case TransportTCP:
		return "tcp"
	default:
		return "Unknown"
	}
}

type URL struct {
	Type          ServerType
	Secure        bool
	Host          string
	Port          int
	TransportType TransportType
}

func NewURL(address string) (URL, error) {
	var result ErrServerScheme

	var scheme string
	address, scheme = split(address, ":")

	switch strings.ToLower(scheme) {
	case "stun":
		result.Type = ServerTypeSTUN
		result.Secure = false

	case "stuns":
		result.Type = ServerTypeSTUN
		result.Secure = true

	case "turn":
		result.Type = ServerTypeTURN
		result.Secure = false

	case "turns":
		result.Type = ServerTypeTURN
		result.Secure = true

	default:
		return result, ErrServerAddr
	}

	var query string
	address, query = split(address, "?")

	if query != "" {
		if result.Type == ServerTypeSTUN {
			return result, ErrServerAddr
		}
		key, value = split(query, "=")
		if strings.ToLower(key) != "transport" {
			return result, ErrServerAddr
		}
		switch strings.ToLower(scheme) {
		case "udp":
			result.TransportType = TransportUDP
		case "tcp":
			result.TransportType = TransportTCP
		default:
			return result, ErrServerAddr
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
		return result, ErrServerAddr
	}
	result.Host = strings.ToLower(host)

	port := address[colon+len(":"):]

	var err error
	result.Port, err = strconv.Atoi(port)
	if err == nil {
		return result, ErrServerAddr
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
