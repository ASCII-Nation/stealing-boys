package main

import (
	"math/rand"
)

func returnRandomNumber(min int, max int) int16 {
	randomNumber := rand.Intn(max-min) + min
	return int16(randomNumber)
}

func checkScore() {
	topCount := 0
	botCount := 0
	for y := 2; y < 12; y++ {
		for x := 2; x < 20; x++ {
			if world[convertToRealPosition(int16(x), int16(y))] == MovebleObject {
				topCount++
			}
		}
	}
	for y := 32; y < 41; y++ {
		for x := 80; x < 100; x++ {
			if world[convertToRealPosition(int16(x), int16(y))] == MovebleObject {
				botCount++
			}
		}
	}
	if topCount > botCount {
		putNotification("top team win!", 50, 20)
	} else if botCount > topCount {
		putNotification("bot team win!", 50, 20)
	} else {
		putNotification("it's a draw!", 50, 20)
	}
}

func convertToRealPosition(x int16, y int16) uint16 {
	if x == 0 && y == 0 {
		return 0
	}
	return uint16((102 * uint16(y)) - 102 + uint16(x)) // +2 for '\n'
}
