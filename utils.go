package main

import (
	"fmt"
	"math/rand"
)

func returnRandomNumber(min int, max int) uint8 {
	randomNumber := rand.Intn(max-min) + min
	return uint8(randomNumber)
}

func checkScore() {
	topCount := 0
	botCount := 0
	for y := 2; y < 12; y++ {
		for x := 2; x < 20; x++ {
			if world[convertToRealPosition(uint8(x), uint8(y))] == MovebleObject {
				topCount++
			}
		}
	}
	for y := 32; y < 41; y++ {
		for x := 80; x < 100; x++ {
			if world[convertToRealPosition(uint8(x), uint8(y))] == MovebleObject {
				botCount++
			}
		}
	}
	fmt.Println(topCount)
	fmt.Println(botCount)
}

func convertToRealPosition(x uint8, y uint8) uint16 {
	if x == 0 && y == 0 {
		return 0
	}
	return uint16((102 * uint16(y)) - 102 + uint16(x)) // +2 for '\n'
}
