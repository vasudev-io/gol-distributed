package stubs

var Processsor = "GameofLifeOperations.Process"
var GetAlive = "GameofLifeOperations.GetAlivers"

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

//type EmptyReq struct {
//}

type AliveResp struct {
	Alive_Cells int
	Turns       int
}
type WorldChecker struct {
	world [][]byte
}
