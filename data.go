package main

import (
	"sync"
	"time"
)

const (
	EmptySpace      = ' '
	WallObject      = '#'
	MovebleObject   = 'X'
	GoodPlayer      = '@'
	BadPlayer       = '&'
	DropYummyTime   = 7
	MaxPlayersCount = 5
	MainStageTime   = 100
	PrepareStage    = 0
	MainStage       = 1
	FinishStage     = 2
)

var (
	currentStage       uint8
	playersCount       uint8
	readyCount         uint8
	lastPlayer         byte
	lastDropTime       time.Time
	world              []byte
	forgottenPositions = make([][2]int16, 0, 10)
	players            = make(map[string]*player) // Пул игроков
	playersMu          sync.Mutex                 // Мьютекс для безопасного доступа к пулу
)

type player struct {
	xPosition int16
	yPosition int16
	name      byte
	ready     bool
	lastSeen  time.Time
}
