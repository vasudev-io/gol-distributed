package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type GameofLifeOperations struct{}

//executes CalculateNextState and returns the new state via response to client
func (s *GameofLifeOperations) Process(req stubs.Request, res *stubs.Response) (err error) {
	// take the parameters from the req util thingy
	var world = req.World
	world = calculateNextState(req.P, world)

	// send the next turn stRequestuff thru to the response struct
	res.World = world

	return
}

//executes calculatealivecell and returns the new alive cells slice via response to client

func (s *GameofLifeOperations) GetAlivers(req stubs.Request, res *stubs.AliveResp) (err error) {
	res.Alive_Cells = calculateAliveCells(req.P, req.World)
	return
}

//compares world with new world and returns cells flipped via response
func (s *GameofLifeOperations) GetCellsFlipped(req stubs.Request2, res *stubs.AliveResp) (err error) {
	newWorldData := req.NewWorld
	world := req.World
	returnable := make([]util.Cell, 0)
	for row := 0; row < req.P.ImageHeight; row++ {
		for col := 0; col < req.P.ImageWidth; col++ {
			if newWorldData[row][col] != world[row][col] {
				cell := util.Cell{X: row, Y: col}
				returnable = append(returnable, cell)
			}
		}
	}
	//c.events <- TurnComplete{CompletedTurns: turn}
	res.Alive_Cells = returnable
	return
}

//take in empty request and terminates server
func (s *GameofLifeOperations) CancelServer(req stubs.EmptyReq, res *stubs.ServerCancelled) (err error) {
	os.Exit(0)
	return
}

func main() {
	//initializes the listening port
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	//registers the GoL operations for them to be used by client/distributor
	rpc.Register(&GameofLifeOperations{})
	//listens to port for any input from client
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)
	//accepts connection to listener and serves requests for each incoming connection
	rpc.Accept(listener)
}

//calculates the next state by giving in the world and params to get the image height and width
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

//calculates the alive cells and returns as []util.Cell with the co-ordinates of them
func calculateAliveCells(p stubs.Params, world [][]byte) []util.Cell {

	var alivecells []util.Cell

	for row := 0; row < p.ImageWidth; row++ {
		for col := 0; col < p.ImageHeight; col++ {

			pair := util.Cell{}
			currentCell := world[row][col]

			if currentCell == 255 {
				pair.X = col
				pair.Y = row
				alivecells = append(alivecells, pair)

			}
		}
	}

	return alivecells
}
