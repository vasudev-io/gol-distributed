package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris/gameoflife/sdl"
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

	var alivecount []util.Cell
	go func() {//step 2 distributed
		tk := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-tk.C:
				res := new(stubs.AliveResp)
				client.Call(stubs.GetAlive, stubs.EmptyReq{}, res)
				//print(res.Turns)
				alivecount = res.Alive_Cells
				c.events <- AliveCellsCount{res.Turns, res.Alive_Cells}
			}
		}

	}()
	var input rune = <-keyPresses
	switch input{
	case 's':
			c.ioCommand <- ioOutput
			turnstr := strconv.Itoa(p.Turns)
			name := height + "x" + width + "x with s" + turnstr//change this later
			c.ioFilename <- name
			for row := 0; row < p.ImageHeight; row++ {
				for col := 0; col < p.ImageWidth; col++ {
					c.ioOutput <- world[row][col]
				}
			}
		case 'q':
			final := FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivecount}
			finished := ImageOutputComplete{turn, FileName}
			c.events <- finished
			c.events <- final
		case 'k':
			c.ioCommand <- ioOutput
			turnstr := strconv.Itoa(p.Turns)
			name := height + "x" + width + "x with k" + turnstr//change this later
			c.ioFilename <- name
			for row := 0; row < p.ImageHeight; row++ {
				for col := 0; col < p.ImageWidth; col++ {
					c.ioOutput <- world[row][col]
				}
			}
			final := FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivecount}
			finished := ImageOutputComplete{turn, FileName}
			c.events <- finished
			c.events <- final
			case 'p'://not sure about this as it's concurrent with the next part
			c := true
			for c{
				var update = <- KeyPresses
				if update == 'p'{
					fmt.Println("Continuing")
					c = false
				}
			}


		}
	resval := makeCall(*client, world, p)
	world = resval.World
	turn = resval.P.Turns
	var alivers = resval.Alivers
	//stage 3 distributed
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

	final := FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivers}
	c.events <- final

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
