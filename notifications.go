package main

func putNotification(note string, x uint8, y uint8) {
	for _, i := range note {
		world[convertToRealPosition(x, y)] = byte(i)
		x++
	}
}
