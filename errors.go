package webrtc

import (
	"errors"
	"fmt"
)

const (
	ErrConnectionClosed = errors.New("connection closed")
)

type InvalidStateError struct {
	Err error
}

func (e *AddrError) Error() string {
	return fmt.Sprintf("invalid state error: %v", e.Err)
}

const (
	ErrNoConfig = errors.New("no configuration provided")
)

type UnknownError struct {
	Err error
}

func (e *UnknownError) Error() string {
	return fmt.Sprintf("unknown error: %v", e.Err)
}

const (
	ErrCertificateExpired = errors.New("certificate expired")
	ErrNoTurnCred         = errors.New("turn server credentials required")
	ErrTurnCred           = errors.New("invalid turn server credentials")
	ErrExistingTrack      = errors.New("track aready exists")
)

type InvalidAccessError struct {
	Err error
}

func (e *InvalidAccessError) Error() string {
	return fmt.Sprintf("invalid access error: %v", e.Err)
}

const ()

type NotSupportedError struct {
	Err error
}

func (e *NotSupportedError) Error() string {
	return fmt.Sprintf("not supported error: %v", e.Err)
}

const (
	ErrModPeerIdentity         = errors.New("peer identity cannot be modified")
	ErrModCertificates         = errors.New("certificates cannot be modified")
	ErrModRtcpMuxPolicy        = errors.New("rtcp mux policy cannot be modified")
	ErrModIceCandidatePoolSize = errors.New("ice candidate pool size cannot be modified")
)

type InvalidModificationError struct {
	Err error
}

func (e *InvalidModificationError) Error() string {
	return fmt.Sprintf("invalid modification error: %v", e.Err)
}

type SyntaxError struct {
	Err error
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %v", e.Err)
}

type OverconstrainedError struct {
	Err error
}

func (e *OverconstrainedError) Error() string {
	return fmt.Sprintf("overconstrained error: %v", e.Err)
}
