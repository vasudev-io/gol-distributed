package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

//moved calcAlive to server.go

func makeCall(client rpc.Client, world [][]byte, params Params) *stubs.Response {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request{World: world, P: stubs.Params(params)}
	response := new(stubs.Response)
	client.Call(stubs.Processsor, request, response)
	return response
}

/// basically make new method for rpc call and then put that in th go func before the resval thing.

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	//string conversion for the filename

	height := strconv.Itoa(p.ImageHeight)
	width := strconv.Itoa(p.ImageWidth)

	FileName := width + "x" + height
	c.ioCommand <- ioInput

	c.ioFilename <- FileName

	// TODO: Create a 2D slice to store the world.

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	//updating it with the bytes sent from io.go
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput //not sure if it works
		}
	}

	// TODO: Execute all turns of the Game of Life.

	// the following would iterate calculating next state till done with turns
	turn := 0

	//server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	server := "127.0.0.1:8030"
	flag.Parse()
	client, b := rpc.Dial("tcp", server)
	if b != nil {
		fmt.Println(b)
	}
	defer client.Close()

	reciever := make(chan *stubs.Response)

	go func() {
		resval := makeCall(*client, world, p)

		reciever <- resval

	}()

	tk := time.NewTicker(2 * time.Second)

	rec := <-reciever
	//fmt.Println(rec)

	turn = rec.P.Turns
	world = rec.World

	go func() {
		for i := 0; i < 10; i++ {

			c.events <- TurnComplete{turn}
			fmt.Println(AliveCellsCount{turn, len(calculateAliveCells(p, world))})

			select {

			case <-tk.C:

				//client.Call(stubs.GetAlive, stubs.EmptyReq{}, res)
				//print(res.Turns)
				c.events <- AliveCellsCount{turn, len(calculateAliveCells(p, world))}

				//default:

			}
		}
	}()

	//}

	//}()

	//var alivers = resval.Alivers
	//stage 3
	c.ioCommand <- ioOutput
	turnstr := strconv.Itoa(p.Turns)
	name := height + "x" + width + "x" + turnstr
	c.ioFilename <- name
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col]
		}
	}
	c.events <- ImageOutputComplete{p.Turns, name}

	// report how many turns are over and how many alivers are remaining after each turn
	// and send that to a channel which goes into the FinalTurnComplete

	// TODO: Report the final st	resval := makeCall(*client, world, p)

	alivers := calculateAliveCells(p, world)

	final := FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivers}
	c.events <- final

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {

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
