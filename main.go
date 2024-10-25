package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
)

// ----------------------------------------------------------------------------
// 	Constants
//

//const VLC_BINARY = "/Applications/VLC.app/Contents/MacOS/VLC"
const VLC_BINARY = "/usr/bin/vlc"

// ----------------------------------------------------------------------------
// 	Structs
//

type App struct {
	Videos       []string
	Displays     map[int]Display
	DisplayLocks map[int]*Mutex
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
	if _, ok := a.DisplayLocks[screenId]; !ok {
		http.Error(w, fmt.Sprintf("Locks for screen %d do not exist. Danger!", screenId), http.StatusInternalServerError)
		return
	}

	lock, _ := a.DisplayLocks[screenId]
	if lock.IsLocked() { // TODO (stevenchu): I think I'm not supposed to do this because of race conditions
		http.Error(w, fmt.Sprintf("Screen %d is currently playing a video", screenId), http.StatusBadRequest)
		return
	}

	lock.Lock()
	go func(video string, display Display, mut *Mutex) error {
		cmd := exec.Command(
			VLC_BINARY,
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
		err := cmd.Run()
		if err != nil {
			log.Printf("Error running VLC: %v", err)
			return err
		}
		mut.Unlock()
		return nil
	}(videoReq.Video, display, lock)

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
	cmd := exec.Command("python", "./py-scripts/display_linux.py")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error running Python script: %v", err)
	}
	var displaySlice []Display
	if err := json.Unmarshal(output, &displaySlice); err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}
	displays := make(map[int]Display)
	displayLocks := make(map[int]*Mutex)
	for _, d := range displaySlice {
		displays[d.ScreenID] = d
		displayLocks[d.ScreenID] = &Mutex{}
	}
	state := &App{
		Displays:     displays,
		DisplayLocks: displayLocks,
		Videos:       []string{"./jesus.webm"},
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

type Mutex struct {
	mu     sync.Mutex
	locked bool
}

func (m *Mutex) Lock() {
	m.mu.Lock()
	m.locked = true
}

func (m *Mutex) Unlock() {
	m.locked = false
	m.mu.Unlock()
}

func (m *Mutex) IsLocked() bool {
	return m.locked
}
