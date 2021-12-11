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
const commandPath = "/home/uri/go/src/uri/Petri-DistNetwork/distributed/distributed"

func CreateMotorSimulation() *simulator.SimulationEngine {
	// init Log, create files and build files names
	err := os.MkdirAll("Logs/6subredes", os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating log dir: %s\n", err)
	}
	NetPL, lefsFile := simulator.ParseFilesNames("P0")
	Log := utils.InitLogs("P0")

	// read partners and transition mapping to them
	net := simulator.ReadPartners(NetPL)
	partners := net.Nodes
	myNode := partners["P0"]
	delete(partners, "P0")
	Log.Info.Printf("[%s] Reading partners: \n%s", "P0", partners)

	// Create local node
	node := simulator.MakeNode("P0", myNode.Port, partners, Log)

	// Carga de la subred
	lefs, err := simulator.Load(lefsFile, Log)
	if err != nil {
		println("Couln't load the Petri Net file !")
	}
	return simulator.MakeMotorSimulation(node, lefs, net.MapTransNode, 100, Log)
}

var PLConnect map[string]*ssh.Client
func terminate() {
	for _, conn := range PLConnect {
		_ = conn.Close()
	}
}

// Source: http://networkbit.ch/golang-ssh-client/
func startNodes(partners simulator.Partners, wg *sync.WaitGroup) {
	PLConnect = make(map[string]*ssh.Client, 0)
	for name, proc := range partners {
		PLConnect[name] = utils.ConnectSSH(proc.Username, proc.IP)

		// Execute program
		fmt.Println("Starting: " + name)
		var cmd = commandPath + fmt.Sprintf(" %s %s %d", name, "6subredes", 100)
		fmt.Printf("Node [%s]->[%s]:$ %s\n", name, proc.IP, cmd)
		go utils.RunCommandSSH(cmd, PLConnect[name], wg)
	}
}

func TestFuncTest(t *testing.T) {
	// WaitGroup for synchronisation goroutines, test wait until each net finish
	var wg sync.WaitGroup
	wg.Add(5)

	// Setup Motor Simulation of root net
	ms := CreateMotorSimulation()
	startNodes(ms.Node.Partners, &wg)
	fmt.Println("[P0] Simulating net...")
	ms.SimularPeriodo()
	fmt.Printf("Simulación terminada\n")
	wg.Wait()
	terminate()
}
