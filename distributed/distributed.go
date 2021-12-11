// Este programa requiere 2 parámetros de entrada :
//      - Nombre fichero json de Lefs
//        - Número de ciclo final
//
// Ejemplo : censim  testdata/PrimerEjemplo.rdp.subred0.json  5
package main

import (
	"distributed/packages/simulator"
	"distributed/packages/utils"
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
	Log := utils.InitLogs(nodeName)

	// read partners and transition mapping to them
	net := simulator.ReadPartners(netFile)
	partners := net.Nodes
	myNode := partners[nodeName]
	delete(partners, nodeName)
	Log.Info.Printf("[%s] Reading partners: \n%s", nodeName, partners)

	// Create local node
	node := simulator.MakeNode(nodeName, myNode.Port, partners, Log)

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
