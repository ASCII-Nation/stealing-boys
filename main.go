package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func handlePlayerMovement(key byte, p *player) {
	prevPss := [2]int16{p.xPosition, p.yPosition}
	var x int8 = 0
	var y int8 = 0
	var d int8 = 0
	switch key {
	case 'a':
		x = 1
		d = -1
	case 'd':
		x = 1
		d = 1
	case 'w':
		y = 1
		d = -1
	case 's':
		y = 1
		d = 1
	case 'p':
		tryThrowOver(p)
	case 'r':
		if currentStage == PrepareStage && !p.ready {
			p.ready = true
			readyCount++
		}

	default:
		return
	}

	if tryMovePlayer(x, y, d, p) {
		forgottenPositions = append(forgottenPositions, prevPss)
	}

}

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Missing X-User-ID header", http.StatusUnauthorized)
			return
		}
		playersMu.Lock()
		_, exists := players[userID]
		if !exists {
			if playersCount < MaxPlayersCount {
				if lastPlayer == GoodPlayer {
					lastPlayer = BadPlayer
				} else {
					lastPlayer = GoodPlayer
				}
				players[userID] = &player{
					xPosition: 10,
					yPosition: 10,
					name:      lastPlayer,
					ready:     false,
				}
				world[convertToRealPosition(10, 10)] = lastPlayer
				playersCount++
			}

		}

		playersMu.Unlock()
		r = r.WithContext(r.Context())
		next.ServeHTTP(w, r)
	}
}

func handlePlayer() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()
		userID := r.Header.Get("X-User-ID")
		playersMu.Lock()
		p, exists := players[userID]
		playersMu.Unlock()
		if !exists {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		outcome := string(world)
		if len(body) > 0 {
			handlePlayerMovement(body[0], p)
			outcome = render()
		}
		w.Header().Set("Content-Type", "text/plain")
		// playersMu.Unlock()
		fmt.Fprint(w, outcome)
		// fmt.Println(time.Since(start))
		// playersMu.Lock()
	}

}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	lastDropTime = time.Now()
	lastPlayer = GoodPlayer
	playersCount = 0
	clearWorld()
	currentStage = 0
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("File server request:", r.URL.Path)
		fs.ServeHTTP(w, r)
	}))

	http.HandleFunc("/loh", authenticate(handlePlayer()))
	putNotification("| hello! u can move with 'wasd' |", 40, 16)
	putNotification("| 'X' - stuff will drops |", 40, 18)
	putNotification("| steal it and collect in your storage |", 40, 20)
	putNotification("| u can throw stuff over yourself with 'p' |", 40, 22)
	putNotification("| when you will be ready to play, just press 'r |", 40, 24)
	port := "0.0.0.0:8080"
	fmt.Println("Starting server on port", port)
	go allReady()
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
