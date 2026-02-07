package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"camera-viewer/stream"

	"github.com/joho/godotenv"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

var (
	rtspStream *stream.RTSPStream
	webrtcPeer *stream.WebRTCPeer
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	username := os.Getenv("RTSP_USERNAME")
	password := os.Getenv("RTSP_PASSWORD")
	host := os.Getenv("RTSP_HOST")
	port := os.Getenv("RTSP_PORT")

	rtspUrl := fmt.Sprintf("rtsp://%s:%s@%s:%s/cam/realmonitor?channel=1&subtype=0", username, password, host, port)
	
	rtspStream = stream.NewRTSPStream(rtspUrl)
	
	// rtspStream is a pointer to the RTSPStream object but Go automatically dereferences it for us.
	err = rtspStream.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to RTSP stream: %v", err)
	}
	
	// Defer is used to close the RTSP stream after the main function exits.
	defer rtspStream.Close()

	log.Println("Connected to RTSP stream")
	
	// Get the detected codec from the RTSP stream
	codec := rtspStream.GetCodec()
	log.Printf("Camera is using codec: %s", codec)

	webrtcPeer, err = stream.NewWebRTCPeer()
	if err != nil {
		log.Fatalf("Failed to create WebRTC peer: %v", err)
	}
	
	defer webrtcPeer.Close()

	// Create video track with the appropriate codec
	var codecMimeType string
	if codec == "H264" {
		codecMimeType = webrtc.MimeTypeH264
	} else if codec == "H265" {
		codecMimeType = webrtc.MimeTypeH265
	} else {
		log.Fatalf("Unsupported codec: %s", codec)
	}
	
	err = webrtcPeer.CreateVideoTrack("video", codecMimeType)
	if err != nil {
		log.Fatalf("Failed to create video track: %v", err)
	}
	
	// Set up packet handler AFTER creating the video track
	// This handler will be called automatically for each RTP packet received from the camera
	rtspStream.SetPacketHandler(func(packet *rtp.Packet) {
		// Forward the packet to the WebRTC peer
		err := webrtcPeer.WriteRTPPacket(packet)
		if err != nil {
			log.Printf("Failed to write packet to video track: %v", err)
		}
	})

	// Set up connection state monitoring
	webrtcPeer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Connection state changed: %s", state)
	})

	// Set up ICE candidate handling
	// When we discover a new way someone can reach us, log it
	webrtcPeer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		log.Printf("ICE candidate: %s", candidate.String())
	})
	
	log.Println("WebRTC peer created and ready")
	log.Println("Packets will be automatically forwarded from RTSP to WebRTC via callback")

	http.HandleFunc("/api/offer", corsMiddleware(handleOffer))
	http.HandleFunc("/api/answer", corsMiddleware(handleAnswer))

	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// CORS middleware - allows requests from any origin
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the actual handler
		next(w, r)
	}
}


// Passing a pointer to the http.Request type since it is a complex object and therefore should be a pointer.
// So the second param is a pointer of http.Request type.
// ResponseWriter is an interface and by default interface are passed by reference and therefore we don't need to pass a pointer.
// All HTTP handlers in Go MUST have this exact signature (http.ResponseWriter, *http.Request) - it's not your choice
// r is a pointer: Yes, r points to the same http.Request object that the HTTP server created when the request came in
func handleOffer(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received offer request")

	offerSDP, err := webrtcPeer.CreateOffer()
	if err != nil {
		log.Printf("Failed to create offer: %v", err)
		http.Error(w, "Failed to create offer", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"type": "offer",
		"sdp": offerSDP,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Println("Sent offer response")
}

func handleAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received answer request")

	var answer struct {
		Type string `json:"type"`
		SDP string `json:"sdp"`
	}

	// Need to pass memory address so that the decoder can modify the original answer object
	// Passing the struct by value will create a copy
	err := json.NewDecoder(r.Body).Decode(&answer)
	if err != nil {
		log.Printf("Failed to decode answer: %v", err)
		http.Error(w, "Failed to decode answer", http.StatusBadRequest)
		return
	}

	err = webrtcPeer.SetAnswer(answer.SDP)
	if err != nil {
		log.Printf("Failed to set answer: %v", err)
		http.Error(w, "Failed to set answer", http.StatusInternalServerError)
		return
	}

	log.Println("Sent answer response")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})

	log.Println("Successfully set SDP answer - WebRTC connection established!")
}


	