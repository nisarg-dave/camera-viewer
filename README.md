# Camera Viewer

A Go + React application for viewing RTSP camera streams in the browser using WebRTC.

## 🎯 What This Project Does

Converts RTSP camera feeds to WebRTC for browser playback. Your Dahua NVR speaks RTSP, but browsers can't play RTSP directly - this Go backend acts as a bridge, converting RTSP to WebRTC so you can view your cameras in a React web app.

## 🏗️ Architecture
```
Camera (RTSP) → Go Backend (Converter) → Browser (WebRTC)
```

### Data Flow

1. **RTSP Connection**: Go backend connects to Dahua NVR via RTSP
2. **RTP Packet Reading**: Receives raw video packets from camera
3. **WebRTC Conversion**: Forwards packets to WebRTC peer connection
4. **Browser Display**: React app displays video via WebRTC

### Signaling Flow

WebRTC requires a "handshake" to establish connections. We use HTTP endpoints for this:
```
Browser (React)              Go Backend
---------------              ----------

1. Request stream
   | POST /api/stream/start
   |-------------------------->
                              Create PeerConnection
                              Add video track
                              Generate SDP Offer
   <--------------------------
   | Receive Offer
   |
2. Send Answer
   | POST /api/stream/answer
   |-------------------------->
                              Process Answer
                              Complete Connection
   <--------------------------
   |
3. 🎉 Direct WebRTC video streaming begins
   <==========================>
```

## 🔑 Key Concepts

### WebRTC Components

- **PeerConnection**: Main connection manager between Go backend and browser
- **SDP (Session Description Protocol)**: Text format describing media capabilities
  - **Offer**: "Here's what I can send" (Go → Browser)
  - **Answer**: "Here's what I can receive" (Browser → Go)
- **ICE Candidates**: Network paths for establishing connection (handles NAT/firewalls)
- **Signaling**: The handshake mechanism (we use HTTP, but could be WebSocket, etc.)

### Why This Architecture?

- Browsers **cannot** play RTSP streams directly (security/compatibility)
- Browsers **can** play WebRTC natively
- Go backend acts as translator: RTSP → WebRTC
- Once connected, video flows with minimal latency

### Signaling is Custom

**WebRTC doesn't specify a signaling pattern and leaves it up to the developer.**

This means the protocol defines how peers communicate once connected, but not how they exchange connection details (SDP offers/answers and ICE candidates).

We chose **HTTP endpoints** for signaling, but you could use:
- WebSocket (real-time, bidirectional)
- Firebase/Supabase (managed real-time DB)
- Email (slow but technically works!)
- Manual copy/paste (for testing)

The signaling mechanism is **separate** from WebRTC - it's just the handshake. Once connected, media flows directly via WebRTC.

## 📦 Tech Stack

- **Backend**: Go 1.23+
  - `pion/webrtc` - WebRTC implementation
  - `bluenviron/gortsplib` - RTSP client
  - `gorilla/websocket` - WebSocket support
  - `joho/godotenv` - Environment variable management
  
- **Frontend**: React (coming soon)
