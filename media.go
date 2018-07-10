package webrtc

import (
	"math/rand"

	"github.com/pions/webrtc/pkg/rtp"
)

type RTCRtpReceiver struct {
	Track         *RTCTrack
	receiverTrack *RTCTrack
	// receiverTransport
	// receiverRtcpTransport
}

func newRTCRtpReceiver(kind, id string) {
	// TODO: receiving side
}

type RTCRtpSender struct {
	Track       *RTCTrack
	senderTrack *RTCTrack
	// senderTransport
	// senderRtcpTransport
}

func newRTCRtpSender(track *RTCTrack) {
	s := &RTCRtpSender{
		senderTrack: track,
	}
	return s
}

type RTCRtpTransceiverDirection int

const (
	RTCRtpTransceiverDirectionSendrecv RTCRtpTransceiverDirection = iota + 1
	RTCRtpTransceiverDirectionSendonly
	RTCRtpTransceiverDirectionRecvonly
	RTCRtpTransceiverDirectionInactive
)

func (t RTCRtpTransceiverDirection) String() string {
	switch t {
	case RTCRtpTransceiverDirectionSendrecv:
		return "sendrecv"
	case RTCRtpTransceiverDirectionSendonly:
		return "sendonly"
	case RTCRtpTransceiverDirectionRecvonly:
		return "recvonly"
	case RTCRtpTransceiverDirectionInactive:
		return "inactive"
	default:
		return "Unknown"
	}
}

type RTCRtpTransceiver struct {
	Mid              string
	Sender           *RTCRtpSender
	Receiver         *RTCRtpReceiver
	Direction        RTCRtpTransceiverDirection
	currentDirection RTCRtpTransceiverDirection
	firedDirection   RTCRtpTransceiverDirection
	receptive        bool
	stopped          bool
}

func (t *RTCRtpTransceiver) setSendingTrack(track *RTCTrack) {
	t.Sender.senderTrack = track

	switch t.Direction {
	case RTCRtpTransceiverDirectionRecvonly:
		t.Direction = RTCRtpTransceiverDirectionSendrecv
	case RTCRtpTransceiverDirectionInactive:
		t.Direction = RTCRtpTransceiverDirectionSendonly
	default:
		panic("Invalid state change in RTCRtpTransceiver.setSending")
	}
}

func newRTCRtpTransceiver(receiver RTCRtpReceiver,
	sender RTCRtpSender,
	direction RTCRtpTransceiverDirection) *RTCRtpTransceiver {

	r := &RTCRtpTransceiver{
		Receiver:  receiver,
		Sender:    sender,
		Direction: direction,
	}
	return r
}

func (t *RTCRtpTransceiver) Stop() error {
	panic("TODO")
}

// RTCSample contains media, and the amount of samples in it
type RTCSample struct {
	Data    []byte
	Samples uint32
}

type RTCTrack struct {
	PayloadType int
	Kind        RTCRtpCodecType
	Id          string
	Label       string
	Ssrc        uint32
	Samples     chan<- RTCSample
	source      <-chan RTCSample
}

func newRTCTrack(payloadType int, id, label string) (*RTCTrack, error) {
	codec, err := rtcMediaEngine.getCodec(payloadType)
	if err != nil {
		return nil, err
	}

	trackInput := make(chan RTCSample, 15) // Is the buffering needed?
	Ssrc := rand.Uint32()
	go func() {
		packetizer := rtp.NewPacketizer(
			1400,
			payloadType,
			ssrc,
			codec.Payloader,
			rtp.NewRandomSequencer(),
			codec.ClockRate,
		)
		for {
			in := <-trackInput
			packets := packetizer.Packetize(in.Data, in.Samples)
			for _, p := range packets {
				for _, port := range r.ports {
					port.Send(p)
				}
			}
		}
	}()

	t := &RTCTrack{
		PayloadType: payloadType,
		Kind:        codec.Type,
		Id:          id,
		Label:       label,
		ssrc:        Ssrc,
		Samples:     trackInput,
	}

	// TODO: register track

	return t
}

func (r *RTCPeerConnection) AddTrack(track *RTCTrack) (*RTCRtpSender, error) {
	if r.IsClosed {
		return nil, &InvalidStateError{Err: ErrConnectionClosed}
	}
	for _, tranceiver := range r.rtpTransceivers {
		if tranceiver.Sender.Track == nil {
			continue
		}
		if track.Id == tranceiver.Sender.Track.Id {
			return nil, &InvalidAccessError{Err: ErrExistingTrack}
		}
	}
	var tranciever *RTCRtpTransceiver
	for _, t := range r.rtpTransceivers {
		if !t.stopped &&
			!t.hasSent &&
			t.Sender.Track == nil &&
			t.Receiver.Track != nil &&
			t.Receiver.Track.Kind == track.Kind {
			tranciever = t
			break
		}
	}
	if tranciever != nil {
		tranciever.setSendingTrack(track)
	} else {
		var receiver *RTCRtpReceiver
		sender := newRTCRtpSender(track)
		tranciever := newRTCRtpTransceiver(
			receiver,
			sender,
			RTCRtpTransceiverDirectionSendonly,
		)
	}

}

func (r *RTCPeerConnection) GetSenders() []RTCRtpSender {
	result := make([]RTCRtpSender, len(r.rtpTransceivers))
	for i, tranceiver := range r.rtpTransceivers {
		result[i] = tranceiver.Sender
	}
	return result
}

func (r *RTCPeerConnection) GetReceivers() []RTCRtpReceiver {
	result := make([]RTCRtpReceiver, len(r.rtpTransceivers))
	for i, tranceiver := range r.rtpTransceivers {
		result[i] = tranceiver.Receiver
	}
	return result
}

func (r *RTCPeerConnection) GetTransceivers() []RTCRtpTransceiver {
	result := make([]RTCRtpTransceiver, len(r.rtpTransceivers))
	for i, tranceiver := range r.rtpTransceivers {
		result[i] = *tranceiver
	}
	return result
}
