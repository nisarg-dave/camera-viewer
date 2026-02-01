package stream

import (
	"fmt"
	"log"

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
func (p *WebRTCPeer) CreateVideoTrack(trackID string) error {
	// Create a video track
	// H264 is the codec, 90000 is the clock rate for the video
	// This sends RTP packets over the track to the browser.
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
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

	log.Println("Video track created and added to peer connection")
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