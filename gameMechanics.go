package main

import (
	"time"
)

func allReady() {
	for {
		if readyCount == playersCount && playersCount != 0 {
			break
		}
		time.Sleep(time.Second * 2)
	}

	for _, value := range players {
		value.ready = false
		world[convertToRealPosition(value.xPosition, value.yPosition)] = value.name
	}
	readyCount = 0
	currentStage = MainStage
	go stageController()

}

func stageController() {
	commonTime := 0
	clearWorld()
	for {
		switch currentStage {
		case MainStage:
			currentTime := time.Now()
			if currentTime.Sub(lastDropTime) > DropYummyTime*time.Second {
				dropMovebaleObject()
				lastDropTime = currentTime
				commonTime += DropYummyTime
				if commonTime >= MainStageTime {
					currentStage = FinishStage
					commonTime = 0
				}
			}
			time.Sleep(1000 * time.Millisecond)
		case FinishStage:
			putNotification("!game over!", 50, 18)

			checkScore()

			time.Sleep(time.Second * 10)

			clearWorld()

			currentStage = PrepareStage
			lastDropTime = time.Now()
			go allReady()
			return
		}

	}

}

func tryMovePlayer(x int8, y int8, dir int8, p *player) bool {

	p.xPosition += int16(x * dir)
	p.yPosition += int16(y * dir)

	nextSpace := world[convertToRealPosition(p.xPosition, p.yPosition)]
	switch nextSpace {
	case EmptySpace:
		return true
	case MovebleObject:
		if tryMoveObject(x, y, dir, p) {
			return true
		}
	}

	p.xPosition -= int16(x * dir)
	p.yPosition -= int16(y * dir)

	return false

}

func tryMoveObject(x int8, y int8, dir int8, p *player) bool {
	objX := p.xPosition + int16(x*dir)
	objY := p.yPosition + int16(y*dir)

	if world[convertToRealPosition(objX, objY)] != EmptySpace {
		return false
	}
	world[convertToRealPosition(objX, objY)] = MovebleObject
	return true
}

func tryThrowOver(p *player) {

	objectOnTheRight := world[convertToRealPosition(p.xPosition+1, p.yPosition)]
	objectOnTheLeft := world[convertToRealPosition(p.xPosition-1, p.yPosition)]
	objectOnTheUp := world[convertToRealPosition(p.xPosition, p.yPosition-1)]
	objectOnTheDown := world[convertToRealPosition(p.xPosition, p.yPosition+1)]

	if objectOnTheRight == MovebleObject && objectOnTheLeft == EmptySpace {
		world[convertToRealPosition(p.xPosition-1, p.yPosition)] = MovebleObject
		world[convertToRealPosition(p.xPosition+1, p.yPosition)] = EmptySpace
	} else if objectOnTheLeft == MovebleObject && objectOnTheRight == EmptySpace {
		world[convertToRealPosition(p.xPosition+1, p.yPosition)] = MovebleObject
		world[convertToRealPosition(p.xPosition-1, p.yPosition)] = EmptySpace
	} else if objectOnTheUp == MovebleObject && objectOnTheDown == EmptySpace {
		world[convertToRealPosition(p.xPosition, p.yPosition+1)] = MovebleObject
		world[convertToRealPosition(p.xPosition, p.yPosition-1)] = EmptySpace
	} else if objectOnTheDown == MovebleObject && objectOnTheUp == EmptySpace {
		world[convertToRealPosition(p.xPosition, p.yPosition-1)] = MovebleObject
		world[convertToRealPosition(p.xPosition, p.yPosition+1)] = EmptySpace
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
