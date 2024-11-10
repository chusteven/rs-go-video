package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

// ----------------------------------------------------------------------------
// 	Structs
//

type App struct {
	Displays      map[int]Display
	VideoRunners  map[int]*VideoRunner
	Videos        []string
	VLCBinaryPath string
}

type Message struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Available struct {
	Videos   []string  `json:"videos"`
	Displays []Display `json:"displays"`
}
type Display struct {
	ScreenID int     `json:"screen_id"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	OriginX  float64 `json:"origin_x"`
	OriginY  float64 `json:"origin_y"`
	Index    int     `json:"index"`
}

type VideoRequest struct {
	Video    string `json:"video"`
	ScreenID int    `json:"screen_id"`
}

// ----------------------------------------------------------------------------
// 	Handlers
//

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	displays := make([]Display, 0, len(a.Displays))
	for _, d := range a.Displays {
		displays = append(displays, d)
	}
	available := Available{
		Videos:   a.Videos,
		Displays: displays,
	}

	if err := json.NewEncoder(w).Encode(available); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) playVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var videoReq VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&videoReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	screenId := videoReq.ScreenID
	display := a.Displays[screenId]
	if _, ok := a.VideoRunners[screenId]; !ok {
		http.Error(w, fmt.Sprintf("Locks for screen %d do not exist. Danger!", screenId), http.StatusInternalServerError)
		return
	}

	lock, _ := a.VideoRunners[screenId]
	if !lock.playVideo(a.VLCBinaryPath, videoReq.Video, display) {
		http.Error(w, fmt.Sprintf("Screen %d is currently playing a video", screenId), http.StatusBadRequest)
		return
	}

	msg := fmt.Sprintf("playing video %s on screen %d", videoReq.Video, videoReq.ScreenID)
	response := Message{
		Status:  "success",
		Message: msg,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ----------------------------------------------------------------------------
// 	Entrypoint
//

func main() {
	cmd := exec.Command("python", "./py-scripts/displays_mac.py")
	if runtime.GOOS != "darwin" {
		cmd = exec.Command("python", "./py-scripts/displays_linux.py")
	}

	vlcBinaryPath := "/Applications/VLC.app/Contents/MacOS/VLC"
	if runtime.GOOS != "darwin" {
		vlcBinaryPath = "/usr/bin/vlc"
	}

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error running Python script: %v", err)
	}
	var displaySlice []Display
	if err := json.Unmarshal(output, &displaySlice); err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}
	displays := make(map[int]Display)
	videoRunners := make(map[int]*VideoRunner)
	for _, d := range displaySlice {
		displays[d.ScreenID] = d
		videoRunners[d.ScreenID] = &VideoRunner{}
	}

	dirPath := "./videos"
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Error running directory: %v", err)
	}

	videos := make([]string, 0, len(entries))
	for _, entry := range entries {
		videos = append(videos, fmt.Sprintf("./videos/%s", entry.Name()))
	}

	state := &App{
		Displays:      displays,
		VideoRunners:  videoRunners,
		Videos:        videos,
		VLCBinaryPath: vlcBinaryPath,
	}

	http.HandleFunc("/", state.indexHandler)
	http.HandleFunc("/video", state.playVideo)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// ----------------------------------------------------------------------------
// 	Utility structs
//

type VideoRunner struct {
	mx      sync.Mutex
	running bool
}

func (t *VideoRunner) playVideo(vlcBinaryPath string, video string, display Display) bool {
	t.mx.Lock()
	if t.running {
		t.mx.Unlock()
		return false
	}
	t.running = true
	t.mx.Unlock()

	go func() {
		cmd := exec.Command(
			vlcBinaryPath,
			video,
			fmt.Sprintf("--video-x=%f", display.OriginX),
			fmt.Sprintf("--video-y=%f", display.OriginY),
			fmt.Sprintf("--width=%f", display.Width),
			fmt.Sprintf("--height=%f", display.Height),
			"--video-on-top",
			"--fullscreen",
			"--no-video-deco",
			"--key-intf-show=false",
			"--play-and-exit",
		)
		if err := cmd.Run(); err != nil {
			log.Printf("Error running VLC: %v", err)
		}
		t.mx.Lock()
		t.running = false
		t.mx.Unlock()
	}()

	return true
}
