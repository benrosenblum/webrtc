package webrtc

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

}

// RTCSample contains media, and the amount of samples in it
type RTCSample struct {
	Data    []byte
	Samples uint32
}

type RTCTrack struct {
	Kind    string
	Id      string
	Label   string
	ssrc    uint32
	Samples chan<- RTCSample
	source  <-chan RTCSample
}

func newRTCTrack(payloadType RTCRtpCodecType, id string) *RTCTrack {

}

// AddTrack adds a new track to the RTCPeerConnection
// This function returns a channel to push buffers on, and an error if the channel can't be added
// Closing the channel ends this stream
// func (r *RTCPeerConnection) AddTrack(mediaType TrackType, clockRate uint32) (samples chan<- RTCSample, err error) {
// 	if mediaType != VP8 && mediaType != H264 && mediaType != Opus {
// 		panic("TODO Discarding packet, need media parsing")
// 	}
//
// 	trackInput := make(chan RTCSample, 15)
// 	go func() {
// 		ssrc := rand.Uint32()
// 		sdpTrack := &sdp.SessionBuilderTrack{SSRC: ssrc}
// 		var payloader rtp.Payloader
// 		var payloadType uint8
// 		switch mediaType {
// 		case Opus:
// 			sdpTrack.IsAudio = true
// 			payloader = &codecs.OpusPayloader{}
// 			payloadType = 111
//
// 		case VP8:
// 			payloader = &codecs.VP8Payloader{}
// 			payloadType = 96
//
// 		case H264:
// 			payloader = &codecs.H264Payloader{}
// 			payloadType = 100
// 		}
//
// 		r.localTracks = append(r.localTracks, sdpTrack)
// 		packetizer := rtp.NewPacketizer(1400, payloadType, ssrc, payloader, rtp.NewRandomSequencer(), clockRate)
// 		for {
// 			in := <-trackInput
// 			packets := packetizer.Packetize(in.Data, in.Samples)
// 			for _, p := range packets {
// 				for _, port := range r.ports {
// 					port.Send(p)
// 				}
// 			}
// 		}
// 	}()
// 	return trackInput, nil
// }

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
