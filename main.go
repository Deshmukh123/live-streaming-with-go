package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	videoDir  = "videos"
	playlist  = "stream.m3u8"
	ffmpegCmd = "ffmpeg -f dshow -i video=\"HP TrueVision HD Camera\" -c:v libx264 -preset ultrafast -tune zerolatency -f hls -hls_time 2 -hls_list_size 5 -hls_flags delete_segments videos/stream.m3u8"
)

func startStream() {
	cmd := exec.Command("sh", "-c", ffmpegCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}
	log.Println("Streaming started...")
}

func servePlaylist(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(videoDir, playlist))
}

func serveSegments(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[len("/segments/"):] // Extract filename
	http.ServeFile(w, r, filepath.Join(videoDir, file))
}

func getVideoDetails(w http.ResponseWriter, r *http.Request) {
	videoPath := filepath.Join(videoDir, playlist)
	info, err := os.Stat(videoPath)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	details := fmt.Sprintf(`{"resolution": "1280x720", "duration": "Live", "size": "%d bytes"}`, info.Size())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(details))
}

func main() {
	os.MkdirAll(videoDir, os.ModePerm)
	go startStream()

	http.HandleFunc("/stream.m3u8", servePlaylist)
	http.HandleFunc("/segments/", serveSegments)
	http.HandleFunc("/video-details", getVideoDetails)
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
