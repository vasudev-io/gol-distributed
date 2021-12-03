package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
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

// uses an rpc call to get the world of next state by implementing calculateNextState in server
func makeCall(client *rpc.Client, world [][]byte, params Params) [][]byte {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request{World: world, P: stubs.Params(params)}
	response := new(stubs.Response)
	client.Call(stubs.Processsor, request, response)
	return response.World //returns world from response
}

//uses an rpc call to get the alive cells as a []util.Cell
func getalivecall(client *rpc.Client, world [][]byte, params Params) []util.Cell {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	response := new(stubs.AliveResp)
	request := stubs.Request{World: world, P: stubs.Params(params)}
	client.Call(stubs.GetAlive, request, response)
	return response.Alive_Cells
}

//ues an rpc call to get the cells flipped for state to state processing
func cellsflipped(client *rpc.Client, world [][]byte, params Params, newworld [][]byte) []util.Cell {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	response := new(stubs.AliveResp)
	request := stubs.Request2{World: world, P: stubs.Params(params), NewWorld: newworld}
	client.Call(stubs.GetCellsFlipped, request, response)
	return response.Alive_Cells
}

//uses rpc call to terminate server(used with KeyPress 'k)
func cancelserver(client *rpc.Client) bool {
	client.Call(stubs.CancelServer, stubs.EmptyReq{}, stubs.EmptyReq{})
	return true
}

// distributor divides the work between client and server
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	//string conversion for the filename

	//converts height and width to string to be concatenated for the io.Input call
	height := strconv.Itoa(p.ImageHeight)
	width := strconv.Itoa(p.ImageWidth)

	FileName := width + "x" + height
	//passes filename to ioInput to start processing pgm image as series of bytes which will be sent
	c.ioCommand <- ioInput

	c.ioFilename <- FileName

	// TODO: Create a 2D slice to store the world.

	//instantiates 2D world slice to be used later on
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	//the server IP address which we will connect to
	server := "127.0.0.1:8030"
	flag.Parse()
	//dials server to set up connection
	client, b := rpc.Dial("tcp", server)
	//catches error and outputs it if any occurs while dialing
	if b != nil {
		fmt.Println(b)
	}
	defer client.Close()
	//initialising ogworld to be used with getcellflipped for cell flipping processing
	ogworld := world
	//initializes the mutex lock that will be used later on to pause gol with KeyPress 'p'
	var lock sync.Mutex

	//first cell flip processing event
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput
			if world[row][col] == 255 {
				cell := util.Cell{X: row, Y: col}
				c.events <- CellFlipped{CompletedTurns: 0, Cell: cell}
			}
		}

	}
	// TODO: Execute all turns of the Game of Life.
	tk := time.NewTicker(2 * time.Second) //intializing+declaring the ticker to use for step 2 distributed

	//initialising turn to use with the main state change for loop
	var turn = 0
	//initialising ran to use with k as to make server not terminate while processing next state
	var ran = 0
	go func() { //go-routine runs concurrently with the for loop following it (contains step 2 and step 4)
		for {
			select {
			//(step 2) every 2 seconds report alive cells and turns to c.events
			case <-tk.C:
				c.events <- AliveCellsCount{turn, len(getalivecall(client, world, p))}
				//if a keyPress has been inputted process it
			case x := <-keyPresses:
				//if keypress is 'p' output Paused and lock the processing of next state till
				//p is re-input and then unpause
				if x == 'p' {
					fmt.Println("Paused!")
					lock.Lock()
					for {
						if <-keyPresses == 'p' {
							lock.Unlock()
							fmt.Println("Continuing..")
							break
						}
					}
					//if keyPress is 's' then switch to ioOutput and then send file name and world
					//and inform events that output is complete to save pgm image
				} else if x == 's' {
					c.ioCommand <- ioOutput

					fturn := strconv.Itoa(turn)

					FileName = width + "x" + height + "x" + fturn

					c.ioFilename <- FileName

					for row := 0; row < p.ImageHeight; row++ {
						for col := 0; col < p.ImageWidth; col++ {
							c.ioOutput <- world[row][col]
						}
					}

					c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: FileName}
					//if KeyPress is 'q' then switch to ioOutput and then send file name and world
					//inform events output is over and also stop the execution from client
				} else if x == 'q' {
					c.ioCommand <- ioOutput
					fturn := strconv.Itoa(turn)
					FileName = width + "x" + height + "x" + fturn
					c.ioFilename <- FileName
					for row := 0; row < p.ImageHeight; row++ {
						for col := 0; col < p.ImageWidth; col++ {
							c.ioOutput <- world[row][col]
						}
					}
					c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: FileName}
					c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: getalivecall(client, world, p)}
					fmt.Println("Terminated.")
					//if keyPress is 'k' then make turn 1 which should cancel server from within later in the next for loop
				} else if x == 'k' {
					ran = 1
				}

			}
		}

	}()
	//loop iterates for every turn
	for turn < p.Turns {
		//mutex locks that 'p' changes to halt execution till un-pause
		lock.Lock()
		lock.Unlock()
		//ogworld is updated to previous world for cell flipping later, as compares previous and current world
		ogworld = world
		//getting new world state and then cell flipping + reporting to events about cells flipped
		world = makeCall(client, world, p)
		cd := cellsflipped(client, world, p, ogworld)
		for _, s := range cd {
			c.events <- CellFlipped{CompletedTurns: turn, Cell: s}
		}
		//updating events turncomplete and incrementing turn (could get rid of)
		c.events <- TurnComplete{turn}
		turn++
		//for keyPress 'k' and if pressed then terminates server from within and then outputs image as saved pgm file and
		//stops/terminates the execution
		if ran == 1 {
			cancelserver(client)
			c.ioCommand <- ioOutput
			fturn := strconv.Itoa(turn)
			FileName = width + "x" + height + "x" + fturn
			c.ioFilename <- FileName
			for row := 0; row < p.ImageHeight; row++ {
				for col := 0; col < p.ImageWidth; col++ {
					c.ioOutput <- world[row][col]
				}
			}
			c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: FileName}
			c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: getalivecall(client, world, p)}
			fmt.Println("Terminated.")
		}
	}
	//after finishing every turn stop the ticker
	tk.Stop()
	//stage 3
	//after finishing every turn takes remaining world and outputs it as a pgm file
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

	// TODO: Report the final st   resval := makeCall(*client, world, p)
	//report alive cells and stop events by sending back final turn
	alivers := getalivecall(client, world, p)

	final := FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivers}
	c.events <- final

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	//changes state to quitting
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
