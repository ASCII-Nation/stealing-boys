package main

func buildHouse() {

	for i := 31; i < 41; i++ {
		if i == 36 {
			world[convertToRealPosition(80, int16(i))] = EmptySpace
		} else {
			world[convertToRealPosition(80, int16(i))] = WallObject
		}
	}

	for i := 80; i < 100; i++ {
		world[convertToRealPosition(int16(i), 31)] = WallObject
	}

	for i := 1; i < 12; i++ {
		if i == 6 {
			world[convertToRealPosition(21, int16(i))] = EmptySpace
		} else {
			world[convertToRealPosition(21, int16(i))] = WallObject
		}

	}

	for i := 1; i < 21; i++ {
		world[convertToRealPosition(int16(i), 11)] = WallObject
	}

}

func fillWorld() {
	for y := 0; y < 41; y++ {
		for x := 0; x < 101; x++ {
			world = append(world, EmptySpace)
		}
		world = append(world, '\n')
	}
}

func buildBorders() {
	for i := 2; i < 41; i++ {
		world[convertToRealPosition(0, int16(i))] = '#'
	}

	for i := 2; i < 41; i++ {
		world[convertToRealPosition(100, int16(i))] = '#'
	}

	for i := 0; i < 101; i++ {
		world[convertToRealPosition(int16(i), 1)] = '#'
	}

	for i := 0; i < 101; i++ {
		world[convertToRealPosition(int16(i), 41)] = '#'
	}

}

func render() string {

	for _, value := range forgottenPositions {
		world[convertToRealPosition(value[0], value[1])] = EmptySpace
	}
	for _, player := range players {
		if player.ready {
			world[convertToRealPosition(player.xPosition, player.yPosition)] = 'R'
		} else {
			world[convertToRealPosition(player.xPosition, player.yPosition)] = player.name
		}

	}

	forgottenPositions = forgottenPositions[:0]

	return string(world)
}

func clearWorld() {
	playersMu.Lock()
	world = world[:0]
	fillWorld()
	buildBorders()
	buildHouse()
	for _, value := range players {
		world[convertToRealPosition(value.xPosition, value.yPosition)] = value.name // TODO: need to replace on better solution!
	}
	playersMu.Unlock()
}
