package stream

import (
	"fmt"
	"log"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
)

type RTSPStream struct {
	URL string
	client *gortsplib.Client // pointer to the RTSP client object. It's a complex object and therefore should be a pointer.
	onPacketHandler func(*rtp.Packet) // Callback function to handle incoming RTP packets
	detectedCodec string // The codec type detected from the stream (H264 or H265)
}

// All these methods need to be exported so they are pascal case and therefore public.

// NewRTSPStream creates a new RTSP stream connection
// This returns a pointer to the RTSPStream object. I.e., pointer to the same object in memory.
func NewRTSPStream(rtspURL string) *RTSPStream {
	// & means "address of" - it's a memory address of the RTSPStream object.
	return &RTSPStream{
		URL: rtspURL,
	}
}

// Connect establishes connection to the RTSP camera and sets up packet handlers
// This function is called a receiver function - it's a function that is called when the object is used.
// You call it like this: stream.Connect()
// The s is the receiver of the function. It's a pointer to the RTSPStream type.
// The error is the return value of the function. It's a error object.
// It is a pointer to that type so that the original object is modified.
func (s *RTSPStream) Connect() error {
	// parse the URL
	parsedURL, err := base.ParseURL(s.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// create a new RTSP client
	// We use the & to get the address of the RTSPStream object.
	// Therefore, we are creating a pointer
	s.client = &gortsplib.Client{}

	// Connect to the camera using Start(scheme, host) for v4
	err = s.client.Start(parsedURL.Scheme, parsedURL.Host)
	if err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// Read the stream description (what formats are available)
	// session is a pointer but Go automatically dereferences it for us.
	session, _, err := s.client.Describe(parsedURL)
	if err != nil {
		return fmt.Errorf("failed to describe stream: %w", err)
	}

	log.Printf("Connected to camera, found %d tracks", len(session.Medias))

	// Setup packet handlers for each media track
	// This is the new callback-based approach in gortsplib v4
	// Iterates through each media track in the session.
	// _ is a blank identifier. It is used to ignore the index of the loop.
	var setupCount int
	for _, media := range session.Medias {
		log.Printf("Processing media track with %d formats", len(media.Formats))
		
		// Find video format (H264 or H265)
		for _, forma := range media.Formats {
			// Debug: log what format type we're checking
			log.Printf("Checking format type: %T", forma)
			// Type assertion to check if this is an H264 format
            // forma is an interface type (could be any format)
            // (*format.H264) - we're asking "is this specifically an H264 format?
            // h264Format, ok := - this returns TWO values:
            // h264Format - the value converted to *format.H264 type (if successful)
            // ok - a boolean: true if the conversion worked, false if not
            // ; ok - only enters the if block if ok is true
			// Try H264 format first
			if h264Format, ok := forma.(*format.H264); ok {
				log.Printf("Found H264 video format - setting up...")
				
				// Setup this media track (port 0, 0 means auto-select)
				_, err = s.client.Setup(session.BaseURL, media, 0, 0)
				if err != nil {
					return fmt.Errorf("failed to setup media: %w", err)
				}
				
				log.Printf("Successfully set up H264 media track")
				s.detectedCodec = "H264"
				setupCount++
				
				// Set up the OnPacketRTP handler for this media
				// This callback is called automatically when packets arrive
				s.client.OnPacketRTP(media, h264Format, func(pkt *rtp.Packet) {
					// Call our custom handler if it's set
					if s.onPacketHandler != nil {
						s.onPacketHandler(pkt)
					}
				})
				
				// Break after setting up the first video track
				break
			}
			
			// Try H265 format if H264 wasn't found
			if h265Format, ok := forma.(*format.H265); ok {
				log.Printf("Found H265 video format - setting up...")
				
				// Setup this media track (port 0, 0 means auto-select)
				_, err = s.client.Setup(session.BaseURL, media, 0, 0)
				if err != nil {
					return fmt.Errorf("failed to setup media: %w", err)
				}
				
				log.Printf("Successfully set up H265 media track")
				s.detectedCodec = "H265"
				setupCount++
				
				// Set up the OnPacketRTP handler for this media
				// This callback is called automatically when packets arrive
				s.client.OnPacketRTP(media, h265Format, func(pkt *rtp.Packet) {
					// Call our custom handler if it's set
					if s.onPacketHandler != nil {
						s.onPacketHandler(pkt)
					}
				})
				
				// Break after setting up the first video track
				break
			}
		}
	}
	
	if setupCount == 0 {
		return fmt.Errorf("no H264 or H265 video format found in stream - check camera codec settings")
	}
	
	log.Printf("Set up %d media track(s)", setupCount)

	// Start playing the stream
	// After this, packets will start arriving via the OnPacketRTP callbacks
	_, err = s.client.Play(nil)
	if err != nil {
		return fmt.Errorf("failed to play: %w", err)
	}

	log.Println("RTSP stream is now playing!")

	// Go doesn't have exception handling, so we return an error if something goes wrong.
	return nil

}

// SetPacketHandler sets the callback function that will be called for each RTP packet
// This must be called before Connect() to receive packets
func (s *RTSPStream) SetPacketHandler(handler func(*rtp.Packet)) {
	s.onPacketHandler = handler
}

// GetCodec returns the detected video codec (H264 or H265)
// This should be called after Connect() to get the actual codec used
func (s *RTSPStream) GetCodec() string {
	return s.detectedCodec
}

// Close closes the RTSP client connection
func (s *RTSPStream) Close() error {
	if s.client != nil {
		s.client.Close()
	}
	return nil
}