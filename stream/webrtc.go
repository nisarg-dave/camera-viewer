package stream

import (
	"fmt"
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type WebRTCPeer struct{
	peerConnection *webrtc.PeerConnection
	videoTrack *webrtc.TrackLocalStaticRTP // Video channel we will send packets through to the browser. I.e., this is what is used to send the video stream using RTP (Real-time Transport Protocol) packets coming from the camera.
}

func NewWebRTCPeer() (*WebRTCPeer, error) {
	// Configure the WebRTC peer connection
	// ICE (Interactive Connectivity Establishment) is the process of establishing a connection between two peers.
	// We use a STUN server to get the public IP address of the peer.
	// STUN servers are useful when peer is behind a router with a NAT (Network Address Translation).
	// We are using Google's free stun server to get the public IP address of the peer.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	return &WebRTCPeer{
		peerConnection: peerConnection,
	}, nil
}

// CreateVideoTrack creates a video track for sending video to the browser
// codecMimeType should be either webrtc.MimeTypeH264 or webrtc.MimeTypeH265
func (p *WebRTCPeer) CreateVideoTrack(trackID string, codecMimeType string) error {
	// Create a video track with the specified codec
	// 90000 is the standard clock rate for video
	// This sends RTP packets over the track to the browser.
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: codecMimeType},
		"video", // The track ID is the name of the track
		"camera-stream", // The track label is the name of the track
	)
	if err != nil {
		return fmt.Errorf("failed to create video track: %w", err)
	}

	p.videoTrack = videoTrack
	
	// Add the video track to the peer connection
	_, err = p.peerConnection.AddTrack(videoTrack)
	if err != nil {
		return fmt.Errorf("failed to add video track to peer connection: %w", err)
	}

	log.Printf("Video track created with codec %s and added to peer connection", codecMimeType)
	return nil
}

// CreateOffer generates an SDP offer to send to the browser
func (p *WebRTCPeer) CreateOffer() (string, error){
	// Create an offer
	offer, err := p.peerConnection.CreateOffer(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create offer: %w", err)
	}

	// Set the offer to the peer connection
	err = p.peerConnection.SetLocalDescription(offer)
	if err != nil {
		return "", fmt.Errorf("failed to set local description: %w", err)
	}

	return offer.SDP, nil
}

// SetAnswer processes the SDP answer from the browser
func (p *WebRTCPeer) SetAnswer(answerSDP string) error {
	// Create an answer object from the SDP string
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP: answerSDP,
	}

	// Set the answer to the peer connection
	err := p.peerConnection.SetRemoteDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	log.Println("Answer set to peer connection")
	return nil
}

// OnICECandidate sets up a handler for when ICE candidates are found
// Called when we find a network path (send to browser)
// Parameter is like a callback function. It is a function that is called when the event happens.
func (p *WebRTCPeer) OnICECandidate(handler func(*webrtc.ICECandidate)) {
	p.peerConnection.OnICECandidate(handler)
}

// OnConnectionStateChange sets up a handler for connection state changes
func (p *WebRTCPeer) OnConnectionStateChange(handler func(webrtc.PeerConnectionState)) {
	p.peerConnection.OnConnectionStateChange(handler)
}

// Close closes the peer connection
func (p *WebRTCPeer) Close() error {
	if p.peerConnection != nil {
		return p.peerConnection.Close()
	}
	return nil
}

// WriteRTPPacket writes an RTP packet object to the video track
// This is used when you have an *rtp.Packet from the RTSP stream
func (p *WebRTCPeer) WriteRTPPacket(packet *rtp.Packet) error {
	if p.videoTrack == nil {
		return fmt.Errorf("video track not created")
	}

	// Marshal the RTP packet to bytes
	data, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal RTP packet: %w", err)
	}

	// Write the marshaled packet to the video track
	_, err = p.videoTrack.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write packet to video track: %w", err)
	}
	return nil
}

// GetVideoTrack returns the video track
func (p *WebRTCPeer) GetVideoTrack() *webrtc.TrackLocalStaticRTP {
	return p.videoTrack
}