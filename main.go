// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// )

// const (
// 	videoDir  = "videos"
// 	playlist  = "stream.m3u8"
// 	ffmpegCmd = "ffmpeg -f dshow -i video=\"HP TrueVision HD Camera\" -c:v libx264 -preset ultrafast -tune zerolatency -f hls -hls_time 2 -hls_list_size 5 -hls_flags delete_segments videos/stream.m3u8"
// )

// func startStream() {
// 	cmd := exec.Command("sh", "-c", ffmpegCmd)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Start(); err != nil {
// 		log.Fatalf("Failed to start FFmpeg: %v", err)
// 	}
// 	log.Println("Streaming started...")
// }

// func servePlaylist(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, filepath.Join(videoDir, playlist))
// }

// func serveSegments(w http.ResponseWriter, r *http.Request) {
// 	file := r.URL.Path[len("/segments/"):] // Extract filename
// 	http.ServeFile(w, r, filepath.Join(videoDir, file))
// }

// func getVideoDetails(w http.ResponseWriter, r *http.Request) {
// 	videoPath := filepath.Join(videoDir, playlist)
// 	info, err := os.Stat(videoPath)
// 	if err != nil {
// 		http.Error(w, "Video not found", http.StatusNotFound)
// 		return
// 	}

// 	details := fmt.Sprintf(`{"resolution": "1280x720", "duration": "Live", "size": "%d bytes"}`, info.Size())
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write([]byte(details))
// }

// func main() {
// 	os.MkdirAll(videoDir, os.ModePerm)
// 	go startStream()

// 	http.HandleFunc("/stream.m3u8", servePlaylist)
// 	http.HandleFunc("/segments/", serveSegments)
// 	http.HandleFunc("/video-details", getVideoDetails)
// 	http.Handle("/", http.FileServer(http.Dir(".")))

// 	fmt.Println("Server started at http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

var (
	ffmpegCmd *exec.Cmd
	mu        sync.Mutex
)

func main() {
	// Ensure the videos directory exists
	os.MkdirAll("./videos", 0755)

	// Serve HLS files from /hls/
	http.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir("./videos"))))

	// Serve the HTML player
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle live streaming control
	http.HandleFunc("/start-stream", startStreamHandler)
	http.HandleFunc("/stop-stream", stopStreamHandler)

	port := 8080
	fmt.Printf("Starting server at http://localhost:%d/\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Start FFmpeg live stream
func startStreamHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if ffmpegCmd != nil {
		http.Error(w, "Stream already running", http.StatusConflict)
		return
	}

	// FFmpeg command for Windows (uncomment if using Windows)
	ffmpegCmd = exec.Command("ffmpeg", "-f", "dshow" , "-i", "video=HP TrueVision HD Camera",
		"-c:v", "libx264", "-f", "hls", "-hls_time", "2", "-hls_list_size", "3", "videos/live.m3u8")

	ffmpegCmd.Stdout = os.Stdout
	ffmpegCmd.Stderr = os.Stderr

	if err := ffmpegCmd.Start(); err != nil {
		http.Error(w, "Failed to start stream", http.StatusInternalServerError)
		ffmpegCmd = nil
		return
	}

	fmt.Fprintln(w, "Live stream started")
}

// Stop FFmpeg live stream
func stopStreamHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if ffmpegCmd == nil {
		http.Error(w, "No stream running", http.StatusNotFound)
		return
	}

	if err := ffmpegCmd.Process.Kill(); err != nil {
		http.Error(w, "Failed to stop stream", http.StatusInternalServerError)
		return
	}

	ffmpegCmd = nil
	fmt.Fprintln(w, "Live stream stopped")
}
