package main

func putNotification(note string, x int16, y int16) {
	for _, i := range note {
		world[convertToRealPosition(x, y)] = byte(i)
		x++
	}
}
