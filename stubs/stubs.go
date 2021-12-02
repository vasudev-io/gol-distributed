package stubs

import "uk.ac.bris.cs/gameoflife/util"

var Processsor = "GameofLifeOperations.Process"
var GetAlive = "GameofLifeOperations.GetAlivers"
var CancelServer = "GameofLifeOperations.CancelServer"
var GetCellsFlipped = "GameofLifeOperations.GetCellsFlipped"

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
type Response struct {
	World [][]byte
	//	Alivers []util.Cell //not sure
	//P Params

	//gol.State
}

type Request struct {
	World [][]byte
	P     Params

	//gol.State
}

type EmptyReq struct {
}
type Request2 struct {
	World    [][]byte
	P        Params
	NewWorld [][]byte
}

type AliveResp struct {
	Alive_Cells []util.Cell
}
type WorldChecker struct {
	world [][]byte
}
type ServerCancelled struct {
	yes bool
}
