package stream

import (
	"fmt"
	"log"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
)

type RTSPStream struct {
	URL string
	client *gortsplib.Client // pointer to the RTSP client object. It's a complex object and therefore should be a pointer.
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

// Connect establishes connection to the RTSP camera
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

	// Connect to the camera
	err = s.client.Start(parsedURL.Scheme, parsedURL.Host)
	if err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// Read the stream description (what formats are available)
	session, _, err := s.client.Describe(parsedURL)
	if err != nil {
		return fmt.Errorf("failed to describe stream: %w", err)
	}

	log.Printf("Connected to camera, found %d tracks", len(session.Medias))

	// Setup all tracks (usually video + audio)
	err = s.client.SetupAll(parsedURL, session.Medias)
	if err != nil {
		return fmt.Errorf("failed to setup tracks: %w", err)
	}

	// Start playing the stream
	_, err = s.client.Play(nil)
	if err != nil {
		return fmt.Errorf("failed to play: %w", err)
	}

	log.Println("RTSP stream is now playing!")

	// Go doesn't have exception handling, so we return an error if something goes wrong.
	return nil

}

// ReadPackets starts reading RTP packets from the stream
func (s *RTSPStream) Close() error {
	if s.client != nil {
		s.client.Close()
	}
	return nil
}