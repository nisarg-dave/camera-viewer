package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"camera-viewer/stream"

	"github.com/joho/godotenv"
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
	
	rtspStream := stream.NewRTSPStream(rtspUrl)
	err = rtspStream.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to RTSP stream: %v", err)
	}

	log.Println("Connected to RTSP stream")

	// Defer is used to close the RTSP stream after the main function exits.
	defer rtspStream.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/cameras", camerasHandler)

	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the camera viewer!")
}

// Passing a pointer to the http.Request type since it is a complex object and therefore should be a pointer.
// So the second param is a pointer of http.Request type.
// ResponseWriter is an interface and by default interface are passed by reference and therefore we don't need to pass a pointer.
func camerasHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "List of cameras:")
}