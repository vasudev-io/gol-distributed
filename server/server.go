package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type GameofLifeOperations struct{}

func (s *GameofLifeOperations) Process(req stubs.Request, res *stubs.Response) (err error) {

	// for loop to take the request and run the distributor code thru it and then send this code off to
	// the response

	// take the parameters from the req util thingy
	turn := 0

	for turn < req.P.Turns {
		req.World = calculateNextState(req.P, req.World)

		turn++
	}

	// send the next turn stuff thru to the response struct
	res.World = req.World
	res.P.Turns = turn

	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameofLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}

/* Super-Secret `reversing a string' method we can't allow clients to see.
func ReverseString(s string, i int) string {
	time.Sleep(time.Duration(rand.Intn(i))* time.Second)
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
} */

func calculateNextState(p stubs.Params, world [][]byte) [][]byte {

	//making a separate world to check without disturbing the actual world
	testerworld := make([][]byte, len(world))
	for i := range world {
		testerworld[i] = make([]byte, len(world[i]))
		copy(testerworld[i], world[i])
	}

	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {

			alivemeter := 0
			for i := row - 1; i <= row+1; i++ {
				for j := col - 1; j <= col+1; j++ {

					if i == row && j == col {
						continue
					}

					if world[((i + p.ImageWidth) % p.ImageWidth)][(j+p.ImageHeight)%p.ImageHeight] == 255 {
						alivemeter++

					}
				}
			}

			// game of life conditions
			if alivemeter < 2 || alivemeter > 3 {
				testerworld[row][col] = 0
			}
			if alivemeter == 3 {
				testerworld[row][col] = 255
			}
		}
	}

	return testerworld
}
