package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type player struct {
	xPosition uint8
	yPosition uint8
	name      byte
	ready     uint8
}

var (
	currentStage       uint8
	playersCount       uint8
	lastPlayer         byte
	lastDropTime       time.Time
	world              []byte
	forgottenPositions = make([][2]uint8, 0, 10)
	players            = make(map[string]*player) // Пул игроков
	playersMu          sync.Mutex                 // Мьютекс для безопасного доступа к пулу
)

func tryMoveObject(currentPlayer *player, dir uint8, object byte) uint8 {
	x, y := currentPlayer.xPosition, currentPlayer.yPosition

	switch dir {
	case Left:
		if world[convertToRealPosition(x-1, y)] != EmptySpace {
			currentPlayer.xPosition += 1
			return 1
		}
		world[convertToRealPosition(x-1, y)] = object
	case Right:
		if world[convertToRealPosition(x+1, y)] != EmptySpace {
			currentPlayer.xPosition -= 1
			return 1
		}
		world[convertToRealPosition(x+1, y)] = object
	case Up:
		if world[convertToRealPosition(x, y-1)] != EmptySpace {
			currentPlayer.yPosition += 1
			return 1
		}
		world[convertToRealPosition(x, y-1)] = object
	case Down:
		if world[convertToRealPosition(x, y+1)] != EmptySpace {
			currentPlayer.yPosition -= 1
			return 1
		}
		world[convertToRealPosition(x, y+1)] = object
	}
	return 0
}

func throwOver(p *player) {
	objectOnTheRight := world[convertToRealPosition(p.xPosition+1, p.yPosition)]
	objectOnTheLeft := world[convertToRealPosition(p.xPosition-1, p.yPosition)]
	objectOnTheUp := world[convertToRealPosition(p.xPosition, p.yPosition-1)]
	objectOnTheDown := world[convertToRealPosition(p.xPosition, p.yPosition+1)]

	if objectOnTheRight == MovebleObject {
		if objectOnTheLeft == EmptySpace {
			world[convertToRealPosition(p.xPosition-1, p.yPosition)] = objectOnTheRight
			world[convertToRealPosition(p.xPosition+1, p.yPosition)] = EmptySpace
			return
		}
	}
	if objectOnTheLeft == MovebleObject {
		if objectOnTheRight == EmptySpace {
			world[convertToRealPosition(p.xPosition+1, p.yPosition)] = objectOnTheLeft
			world[convertToRealPosition(p.xPosition-1, p.yPosition)] = EmptySpace
			return
		}
	}
	if objectOnTheUp == MovebleObject {
		if objectOnTheDown == EmptySpace {
			world[convertToRealPosition(p.xPosition, p.yPosition+1)] = objectOnTheUp
			world[convertToRealPosition(p.xPosition, p.yPosition-1)] = EmptySpace
			return
		}
	}
	if objectOnTheDown == MovebleObject {
		if objectOnTheUp == EmptySpace {
			world[convertToRealPosition(p.xPosition, p.yPosition-1)] = objectOnTheDown
			world[convertToRealPosition(p.xPosition, p.yPosition+1)] = EmptySpace
			return
		}
	}
}

func movePlayer(x uint8, y uint8, dir uint8, p *player) uint8 {
	switch dir {
	case 1:
		p.xPosition += x
		p.yPosition += y
	case 2:
		p.xPosition -= x
		p.yPosition -= y
	}
	nextSpace := world[convertToRealPosition(p.xPosition, p.yPosition)]
	switch nextSpace {
	case EmptySpace:
	case MovebleObject:
		return tryMoveObject(p, Left, MovebleObject)
	default:
		switch dir {
		case 1:
			p.xPosition -= x
			p.yPosition -= y
		case 2:
			p.xPosition += x
			p.yPosition += y
		}
		return 1
	}
	return 0
}

func handlePlayerMovement(key byte, p *player) {
	prevPss := [2]uint8{p.xPosition, p.yPosition}
	var err uint8 = 0
	var x uint8 = 0
	var y uint8 = 0
	var d uint8 = 0
	switch key {
	case 'a':
		x = 1
		d = 2
	case 'd':
		x = 1
		d = 1
	case 'w':
		y = 1
		d = 2
	case 's':
		y = 1
		d = 1
	case 'p':
		throwOver(p)
	case 'r':
		if currentStage == 0 {
			p.ready = 1
		}

	default:
		return
	}

	err = movePlayer(x, y, d, p)
	if err != 0 {
		return
	}
	forgottenPositions = append(forgottenPositions, prevPss)
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
					ready:     0,
				}
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

func dropMovebaleObject() {
	x := returnRandomNumber(2, 100)
	y := returnRandomNumber(2, 40)
	if world[convertToRealPosition(x, y)] != EmptySpace {
		dropMovebaleObject()
	} else {
		world[convertToRealPosition(x, y)] = MovebleObject
	}
}

func drop() {
	commonTime := 0
	for {
		switch currentStage {
		case 1:
			currentTime := time.Now()
			if currentTime.Sub(lastDropTime) > DropYummyTime*time.Second {
				dropMovebaleObject()
				lastDropTime = currentTime
				commonTime += DropYummyTime
				if commonTime >= MainStageTIme {
					currentStage = 2
					commonTime = 0
					fmt.Println("the main stage is over!")
				}
			}
			time.Sleep(1000 * time.Millisecond)
		case 2: // change change change!!!!!!!!!!!!!!!!!
			fmt.Println("!!!!game over!!!!!")
			fmt.Println("since 10 seconds game will restart")
			checkScore()
			time.Sleep(time.Second * 10)
			playersMu.Lock()
			world = world[:0]
			fillWorld()
			buildBorders()
			buildHouse()
			playersMu.Unlock()
			currentStage = 0
			lastDropTime = time.Now()
			go allReady()
			return
		}

	}

}

func allReady() {
	count := 0
	for {
		for _, value := range players {
			if value.ready != 1 {
				time.Sleep(2000 * time.Millisecond)
				count = 0
				break
			}
			count++
		}
		if count > 0 {
			fmt.Println("all ready!")
			for _, value := range players {
				value.ready = 0
				world[convertToRealPosition(value.xPosition, value.yPosition)] = value.name // TODO: need to replace on better solution!
			}
			go drop()
			break
		}
	}
	fmt.Println("prepare finished!")
	currentStage = 1
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	lastDropTime = time.Now()
	lastPlayer = GoodPlayer
	playersCount = 0
	fillWorld()
	buildBorders()
	buildHouse()
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
