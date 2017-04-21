package main

import (
	"fmt"
	"github.com/ajesler/playbulb-candle"
)

const (
	candleID          = "7d61a760051544f4b83307bc316fada6" // Change this to your candle id
)

var (
	done          = make(chan bool, 1)
)

func main() {
	p := playbulb.NewCandle(candleID)

	red, _ := playbulb.ColourFromHexString("00FF0000")
	green, _ := playbulb.ColourFromHexString("8000FF00")

	currentColour := green

	fmt.Println("Connecting to", candleID)
	err := p.Connect()
	if err != nil {
		fmt.Println("Failed to connect", err)
	}

	p.OnToggle(func (isOn bool) {
		if currentColour == green {
			currentColour = red
		} else {
			currentColour = green
		}

		p.SetColour(currentColour)
	})

	<-done
}