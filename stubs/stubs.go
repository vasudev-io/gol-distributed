package stubs

var Processsor = "GameofLifeOperations.Process"

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
type Response struct {
	World [][]byte
	P     Params

	//gol.State
}

type Request struct {
	World [][]byte
	P     Params

	//gol.State
}
