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
			xx := 0
			yy := 0
			if playersCount < MaxPlayersCount && currentStage == PrepareStage {
				if lastPlayer == GoodPlayer {
					lastPlayer = BadPlayer
					xx = 90
					yy = 35
				} else {
					lastPlayer = GoodPlayer
					xx = 10
					yy = 10
				}
				players[userID] = &player{
					xPosition: int16(xx),
					yPosition: int16(yy),
					name:      lastPlayer,
					ready:     false,
				}
				world[convertToRealPosition(int16(xx), int16(yy))] = lastPlayer
				playersCount++
			}

		}

		playersMu.Unlock()
		r = r.WithContext(r.Context())
		next.ServeHTTP(w, r)
	}
}

func deletePlayer() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		playersMu.Lock()
		for id, p := range players {
			if time.Since(p.lastSeen) > 10*time.Second {
				delete(players, id)
				playersCount--
				world[convertToRealPosition(p.xPosition, p.yPosition)] = EmptySpace
			}
		}
		playersMu.Unlock()
	}
}

func handlePlayer() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()
		userID := r.Header.Get("X-User-ID")
		playersMu.Lock()
		p, exists := players[userID]
		playersMu.Unlock()
		p.lastSeen = time.Now()
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

	lastDropTime = time.Now()
	lastPlayer = GoodPlayer
	playersCount = 0
	clearWorld()
	currentStage = 0
	educationNotification()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
	http.HandleFunc("/loh", authenticate(handlePlayer()))
	port := "0.0.0.0:7075"
	fmt.Println("Starting server on port", port)

	go allReady()
	go deletePlayer()
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}
