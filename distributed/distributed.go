// Este programa requiere 2 parámetros de entrada :
//      - Nombre fichero json de Lefs
//        - Número de ciclo final
//
// Ejemplo : censim  testdata/PrimerEjemplo.rdp.subred0.json  5
package main

import (
	"distributed/simulator"
	"fmt"
	"os"
	"strconv"
)

func main() {

	if len(os.Args) != 4 {
		panic("bad usage: distim <nodeName> <files_prefix> <finalClk>")
	}

	var nodeName string
	nodeName = os.Args[1]
	//filesPrefix = os.Args[2]
	netFile, lefsFile := simulator.ParseFilesNames(nodeName)

	// init log
	Log := simulator.LogInitialization(nodeName)

	// read LogicalProcess and transition mapping to them
	net := simulator.ReadLogicalProcess(netFile)
	LogicalProcess := net.Nodes
	myNode := LogicalProcess[nodeName]
	delete(LogicalProcess, nodeName)
	Log.Info.Printf("[%s] Reading LogicalProcess: \n%s", nodeName, LogicalProcess)

	// Create local node
	node := simulator.MakeNode(nodeName, myNode.Port, LogicalProcess, Log)

	// Carga de la subred
	lefs, err := simulator.Load(lefsFile, Log)
	if err != nil {
		println("Couln't load the Petri Net file !")
	}
	cicloFinal, _ := strconv.Atoi(os.Args[3])
	ms := simulator.MakeMotorSimulation(node, lefs, net.MapTransNode, simulator.TypeClock(cicloFinal), Log)

	fmt.Printf("Simulating [%s]\n", nodeName)
	ms.SimularPeriodo()
}
