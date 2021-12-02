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

func makeCall(client *rpc.Client, world [][]byte, params Params) [][]byte {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request{World: world, P: stubs.Params(params)}
	response := new(stubs.Response)
	client.Call(stubs.Processsor, request, response)
	return response.World
}

/// basically make new method for rpc call and then put that in th go func before the resval thing.
//this is my implementation of it at least for the moment
//func cancelserver(client rpc.Client) *stubs.E

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	//, keyPresses <-chan rune
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

	//server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	server := "127.0.0.1:8030"
	flag.Parse()
	client, b := rpc.Dial("tcp", server)
	if b != nil {
		fmt.Println(b)
	}
	defer client.Close()

	tk := time.NewTicker(2 * time.Second)
	var turn = 0
	//mutex := sync.Mutex{}

	go func() {
		for {
			select {

			case <-tk.C:
				//fmt.Println("it is iterating")
				//mutex.Lock()
				c.events <- AliveCellsCount{turn, len(calculateAliveCells(p, world))}
				//mutex.Unlock()
				//if turn == p.Turns {
				//case turn == p.Turns:
				//
				//

			}
		}

	}()
	for turn < p.Turns {

		world = makeCall(client, world, p)

		c.events <- TurnComplete{turn}

		turn++

	}
	tk.Stop()

	//}

	//select {
	//case <-tk.C:
	//fmt.Println("it is iterating")
	//c.events <- AliveCellsCount{turn, len(calculateAliveCells(p, world))}
	/*case key := <- keyPresses:
	if key == 's'{
		c.ioCommand <- ioOutput
		turnstr := strconv.Itoa(p.Turns)
		name := height + "x" + width + "x" + turnstr
		c.ioFilename <- name
		for row := 0; row < p.ImageHeight; row++ {
			for col := 0; col < p.ImageWidth; col++ {
				c.ioOutput <- world[row][col]
			}
		}

	}
	if key == 'q'{
		c.ioCommand <- ioOutput
		turnstr := strconv.Itoa(p.Turns)
		name := height + "x" + width + "x" + turnstr
		c.ioFilename <- name
		for row := 0; row < p.ImageHeight; row++ {
			for col := 0; col < p.ImageWidth; col++ {
				c.ioOutput <- world[row][col]
			}
		}
		c.events <- StateChange{turn, Quitting}
	}
	if key == 'p'{
		c.events <- StateChange{turn, Paused}
		fmt.Println("Pausing")
		var c = true
		for c{
			if <-keyPresses == 'p'{
				c = false
			}
		}
	}
	if key == 'k'{
		//send something to server to kill it with os.Exit()
	}*/
	//default:

	//c.events <- TurnComplete{turn}
	//}

	//	//added this
	//	//reciever := make(chan *stubs.Response)
	//
	//	/*alivereceiver := make(chan int)
	//	turnreceiver := make(chan int)
	//	go func(){//adding this
	//		c := true
	//		for c {
	//			sender := makecallforalivecells(*client, world, p)
	//			alivereceiver <- sender.Alive_Cells
	//			turnreceiver <- sender.Turns		leavalone.Lock()
	//
	//			if sender.Turns == p.Turns{
	//				c = false
	//			}
	//		}
	//
	//	}()//up to here
	//*/
	//	var leavalone sync.Mutex
	//	tk := time.NewTicker(2 * time.Second)
	//	var turn = 0//[ut a lock on this so either it gets world or world alive gets set to events but not both at the same time
	//	var worldd [][]byte
	//
	//	//go func(){//step 2 gol
	//		//for i := 0; i < 10; i++ {
	//			//fmt.Println(AliveCellsCount{turn , len(calculateAliveCells(p, worldd))})//adding this or changing
	//
	//
	//
	//
	//		//}
	//	//}()
	//	for turn < p.Turns{//step 1 gol
	//		leavalone.Lock()
	//		resval := makeCall(*client, world, p)
	//		worldd = resval.World
	//		c.events <- TurnComplete{turn}
	//		leavalone.Unlock()
	//		turn++
	//	}
	//
	//	select {
	//
	//	case <-tk.C:
	//
	//		//client.Call(stubs.GetAlive, stubs.EmptyReq{}, res)
	//		//print(res.Turns)
	//		leavalone.Lock()
	//		c.events <- AliveCellsCount{turn, len(calculateAliveCells(p, worldd))}
	//		leavalone.Unlock()
	//
	//	default:
	//		break //keep this
	//	}
	//	//rec := <-reciever
	//	//fmt.Println(rec)
	//
	//	fmt.Println("turn is ", turn)
	//	//world = rec.World
	//
	//	//}
	//
	//	//}()
	//
	//	//var alivers = resval.Alivers
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

//will have to move this to server
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
