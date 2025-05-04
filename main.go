package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type player struct {
	xPosition uint8
	yPosition uint8
	name      byte
	step      uint8
	ready     uint8
}

const (
	emptySpace      = ' '
	wallObject      = '#'
	movebleObject   = 'X'
	preWallObject   = '0'
	goodPlayer      = '@'
	badPlayer       = '&'
	dropYummyTime   = 3
	dropWallTime    = 1
	LEFT            = 4
	RIGHT           = 6
	UP              = 8
	DOWN            = 2
	maxPlayersCount = 5
	fistStageTime   = 10
	mainStageTIme   = 20
)

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
	x, y, step := currentPlayer.xPosition, currentPlayer.yPosition, currentPlayer.step

	switch dir {
	case LEFT:
		if world[convertToRealPosition(x-step, y)] != emptySpace {
			currentPlayer.xPosition += step
			return 1
		}
		world[convertToRealPosition(x-step, y)] = object
	case RIGHT:
		if world[convertToRealPosition(x+step, y)] != emptySpace {
			currentPlayer.xPosition -= step
			return 1
		}
		world[convertToRealPosition(x+step, y)] = object
	case UP:
		if world[convertToRealPosition(x, y-step)] != emptySpace {
			currentPlayer.yPosition += step
			return 1
		}
		world[convertToRealPosition(x, y-step)] = object
	case DOWN:
		if world[convertToRealPosition(x, y+step)] != emptySpace {
			currentPlayer.yPosition -= step
			return 1
		}
		world[convertToRealPosition(x, y+step)] = object
	}
	return 0
}

func throwOver(p *player) {
	objectOnTheRight := world[convertToRealPosition(p.xPosition+1, p.yPosition)]
	objectOnTheLeft := world[convertToRealPosition(p.xPosition-1, p.yPosition)]
	objectOnTheUp := world[convertToRealPosition(p.xPosition, p.yPosition-1)]
	objectOnTheDown := world[convertToRealPosition(p.xPosition, p.yPosition+1)]

	if objectOnTheRight == movebleObject {
		if objectOnTheLeft == emptySpace {
			world[convertToRealPosition(p.xPosition-1, p.yPosition)] = objectOnTheRight
			world[convertToRealPosition(p.xPosition+1, p.yPosition)] = emptySpace
			return
		}
	}
	if objectOnTheLeft == movebleObject {
		if objectOnTheRight == emptySpace {
			world[convertToRealPosition(p.xPosition+1, p.yPosition)] = objectOnTheLeft
			world[convertToRealPosition(p.xPosition-1, p.yPosition)] = emptySpace
			return
		}
	}
	if objectOnTheUp == movebleObject {
		if objectOnTheDown == emptySpace {
			world[convertToRealPosition(p.xPosition, p.yPosition+1)] = objectOnTheUp
			world[convertToRealPosition(p.xPosition, p.yPosition-1)] = emptySpace
			return
		}
	}
	if objectOnTheDown == movebleObject {
		if objectOnTheUp == emptySpace {
			world[convertToRealPosition(p.xPosition, p.yPosition-1)] = objectOnTheDown
			world[convertToRealPosition(p.xPosition, p.yPosition+1)] = emptySpace
			return
		}
	}
}

func moveLeft(p *player) uint8 {
	p.xPosition -= p.step
	nextSpace := world[convertToRealPosition(p.xPosition, p.yPosition)]
	switch nextSpace {
	case emptySpace:
	case movebleObject:
		return tryMoveObject(p, LEFT, movebleObject)
	default:
		p.xPosition += p.step
		return 1
	}
	return 0
}
func moveRight(p *player) uint8 {
	p.xPosition += p.step
	nextSpace := world[convertToRealPosition(p.xPosition, p.yPosition)]
	switch nextSpace {
	case emptySpace:
	case movebleObject:
		return tryMoveObject(p, RIGHT, movebleObject)
	default:
		p.xPosition -= p.step
		return 1
	}
	return 0
}

func moveUp(p *player) uint8 {
	p.yPosition -= p.step
	nextSpace := world[convertToRealPosition(p.xPosition, p.yPosition)]
	switch nextSpace {
	case emptySpace:
	case movebleObject:
		return tryMoveObject(p, UP, movebleObject)
	default:
		p.yPosition += p.step
		return 1
	}
	return 0
}

func moveDown(p *player) uint8 {
	p.yPosition += p.step
	nextSpace := world[convertToRealPosition(p.xPosition, p.yPosition)]
	switch nextSpace {
	case emptySpace:
	case movebleObject:
		return tryMoveObject(p, DOWN, movebleObject)
	default:
		p.yPosition -= p.step
		return 1
	}
	return 0
}

func handlePlayerMovement(key byte, p *player) {
	prevPss := [2]uint8{p.xPosition, p.yPosition}
	var err uint8 = 0
	switch key {
	case 'a':
		err = moveLeft(p)
	case 'd':
		err = moveRight(p)
	case 'w':
		err = moveUp(p)
	case 's':
		err = moveDown(p)
	case 'p':
		throwOver(p)
	case 'r':
		if currentStage == 0 {
			p.ready = 1
		}

	default:
		return
	}
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
			if playersCount < maxPlayersCount {
				if lastPlayer == goodPlayer {
					lastPlayer = badPlayer
				} else {
					lastPlayer = goodPlayer
				}
				players[userID] = &player{
					xPosition: 10,
					yPosition: 10,
					name:      lastPlayer,
					step:      1,
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

func convertToRealPosition(x uint8, y uint8) uint16 {
	if x == 0 && y == 0 {
		return 0
	}
	return uint16((102 * uint16(y)) - 102 + uint16(x)) // +2 for '\n'
}

func render() string {

	for _, value := range forgottenPositions {
		world[convertToRealPosition(value[0], value[1])] = emptySpace
	}
	for _, value := range players {
		if value.ready == 1 {
			world[convertToRealPosition(value.xPosition, value.yPosition)] = 'R'
		} else {
			world[convertToRealPosition(value.xPosition, value.yPosition)] = value.name
		}

	}

	forgottenPositions = forgottenPositions[:0]

	return string(world)
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

func fillWorld() {
	for y := 0; y < 41; y++ {
		for x := 0; x < 101; x++ {
			world = append(world, emptySpace)
		}
		world = append(world, '\n')
	}
}

func buildBorders() {
	for i := 2; i < 41; i++ {
		world[convertToRealPosition(0, uint8(i))] = '#'
	}

	for i := 2; i < 41; i++ {
		world[convertToRealPosition(100, uint8(i))] = '#'
	}

	for i := 0; i < 101; i++ {
		world[convertToRealPosition(uint8(i), 1)] = '#'
	}

	for i := 0; i < 101; i++ {
		world[convertToRealPosition(uint8(i), 41)] = '#'
	}

}

func returnRandomNumber(min int, max int) uint8 {
	randomNumber := rand.Intn(max-min) + min
	return uint8(randomNumber)
}

func dropMovebaleObject() {
	x := returnRandomNumber(2, 100)
	y := returnRandomNumber(2, 40)
	if world[convertToRealPosition(x, y)] != emptySpace {
		dropMovebaleObject()
	} else {
		world[convertToRealPosition(x, y)] = movebleObject
	}
}

func drop() {
	commonTime := 0
	for {
		switch currentStage {
		case 1:
			currentTime := time.Now()
			if currentTime.Sub(lastDropTime) > dropYummyTime*time.Second {
				dropMovebaleObject()
				lastDropTime = currentTime
				commonTime += dropYummyTime
				if commonTime >= mainStageTIme {
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

func buildHouse() {

	for i := 31; i < 41; i++ {
		if i == 36 {
			world[convertToRealPosition(80, uint8(i))] = emptySpace
		} else {
			world[convertToRealPosition(80, uint8(i))] = wallObject
		}
	}

	for i := 80; i < 100; i++ {
		world[convertToRealPosition(uint8(i), 31)] = wallObject
	}

	for i := 1; i < 12; i++ {
		if i == 6 {
			world[convertToRealPosition(21, uint8(i))] = emptySpace
		} else {
			world[convertToRealPosition(21, uint8(i))] = wallObject
		}

	}

	for i := 1; i < 21; i++ {
		world[convertToRealPosition(uint8(i), 11)] = wallObject
	}

}

func checkScore() {
	topCount := 0
	botCount := 0
	for y := 2; y < 12; y++ {
		for x := 2; x < 20; x++ {
			if world[convertToRealPosition(uint8(x), uint8(y))] == movebleObject {
				topCount++
			}
		}
	}
	for y := 32; y < 41; y++ {
		for x := 80; x < 100; x++ {
			if world[convertToRealPosition(uint8(x), uint8(y))] == movebleObject {
				botCount++
			}
		}
	}
	fmt.Println(topCount)
	fmt.Println(botCount)
}

func putNotification(note string, x uint8, y uint8) {
	for _, i := range note {
		world[convertToRealPosition(x, y)] = byte(i)
		x++
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	lastDropTime = time.Now()
	lastPlayer = goodPlayer
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
