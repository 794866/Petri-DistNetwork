package main

import (
	"distributed/packages/simulator"
	"distributed/packages/utils"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
	"testing"
)

const FIN_CLK simulator.TypeClock = 100

var sshConn map[string]*ssh.Client

func CreateMotorSimulation(nodeName string, filesPrefix string, finClk simulator.TypeClock) *simulator.SimulationEngine {
	// init logger, create files and build files names
	err := os.Mkdir(utils.RelOutputPath+filesPrefix, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating log dir: %s\n", err)
	}
	netFile, lefsFile := utils.ParseFilesNames(nodeName, filesPrefix)
	logger := utils.InitLoggers(filesPrefix, nodeName)

	// read partners and transition mapping to them
	net := simulator.ReadPartners(netFile)
	partners := net.Nodes
	myNode := partners[nodeName]
	delete(partners, nodeName)
	logger.Info.Printf("[%s] Reading partners: \n%s", nodeName, partners)

	// Create local node
	node := simulator.MakeNode(nodeName, myNode.Port, partners, logger)

	// Carga de la subred
	lefs, err := simulator.LoadLefs(lefsFile, logger)
	if err != nil {
		println("Couln't load the Petri Net file !")
	}
	return simulator.MakeMotorSimulation(node, lefs, net.MapTransNode, finClk, logger)
}

func terminate(se *simulator.SimulationEngine) {
	for _, conn := range sshConn {
		_ = conn.Close()
	}
}

// Source: http://networkbit.ch/golang-ssh-client/
func startNodes(filesPrefix string, finishClk simulator.TypeClock, partners simulator.Partners, wg *sync.WaitGroup) {
	sshConn = make(map[string]*ssh.Client, 0)
	for nodeName, node := range partners {
		sshConn[nodeName] = utils.ConnectSSH(node.Username, node.IP)

		// Execute program
		fmt.Println("Starting: " + nodeName)
		var cmd = "/home/uri/go/src/uri/Petri-DistNetwork/distributed/distributed" + fmt.Sprintf(" %s %s %d", nodeName, filesPrefix, finishClk)
		fmt.Printf("Node [%s]->[%s]:$ %s\n", nodeName, node.IP, cmd)
		go utils.RunCommandSSH(cmd, sshConn[nodeName], wg)
	}
}

func TestFuncTest(t *testing.T) {
	// WaitGroup for synchronisation goroutines, test wait until each net finish
	var wg sync.WaitGroup
	wg.Add(5)

	// Setup Motor Simulation of root net
	ms := CreateMotorSimulation("P0", "6subredes", 100)
	startNodes("6subredes", 100, ms.Node.Partners, &wg)
	fmt.Println("[P0] Simulating net...")
	ms.SimularPeriodo()
	fmt.Printf("[%s] Simulación terminada\n", ms.Node.Name)
	wg.Wait()
	terminate(ms)
}


//
//func testXSubnets(nSubnets int, filesPrefix string, finalClk simulator.TypeClock) {
//	// WaitGroup for synchronisation goroutines, test wait until each net finish
//	var wg sync.WaitGroup
//	wg.Add(nSubnets - 1)
//
//	// Setup Motor Simulation of root net
//	ms := CreateMotorSimulation("P0", filesPrefix, finalClk)
//	startNodes(filesPrefix, finalClk, ms.Node.Partners, &wg)
//	fmt.Println("[P0] Simulating net...")
//	ms.SimularPeriodo()
//	fmt.Printf("[%s] Simulación terminada\n", ms.Node.Name)
//	wg.Wait()
//	terminate(ms)
//}
//
//func Test2SubNets2Br(t *testing.T) {
//	testXSubnets(2, "2subredes", FIN_CLK)
//}
//
////func Test3SubNets2Br(t *testing.T) {
////	testXSubnets(3, "3subredes", FIN_CLK)
////}
//
//func Test6SubNets5BrHomogen(t *testing.T) {
//	testXSubnets(6, "6subredes", FIN_CLK)
//}
//
//func Test6SubNets5Br1BrSlow(t *testing.T) {
//	testXSubnets(6, "6subredes1lenta", FIN_CLK)
//}
//
//func Test6SubNets5BrLA(t *testing.T) {
//	testXSubnets(6, "6subredesLA", FIN_CLK)
//}
//
