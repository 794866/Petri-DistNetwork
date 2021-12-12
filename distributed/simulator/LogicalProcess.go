package simulator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const networkPath = "/home/uri/go/src/uri/Petri-DistNetwork/distributed/testdata/"
type LogicalProcess map[string]LPStruct // NameProcess with transitionMapping for all process

type LPStruct struct {
	Username       string `json:"User"`
	IP             string `json:"IP"`
	Port           int    `json:"Port"`
	IncomingEvFIFO EventList
	RemoteSafeTime TypeClock // for incoming events
	LastTimeSent   TypeClock // for outcoming events
}

type MapTransitionNode map[IndTrans]string
type Network struct {
	Nodes        LogicalProcess          `json:"Nodes"`
	MapTransNode MapTransitionNode `json:"TransitionsMapping"`
}

func ReadLogicalProcess(networkFile string) *Network {

	// read network file json with PLs List - IP && ports
	data, err := ioutil.ReadFile(networkPath + networkFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("Network file OK!")

	var network Network
	// parse content of json file to Config struct
	err = json.Unmarshal(data, &network)
	if err != nil  {
		panic(err)
	}
	for name, p := range network.Nodes {
		p.RemoteSafeTime = 0
		p.LastTimeSent = 0
		p.IncomingEvFIFO = MakeEventList(5)
		network.Nodes[name] = p
	}
	return &network
}

func (p LogicalProcess) String() string {
	res := fmt.Sprint("LogicalProcess:")
	for k, v := range p {
		res += fmt.Sprintf("\n\t\t[%s]\t\tIP: %s\t\tPort: %d", k, v.IP, v.Port)
	}
	return res
}
func (p LogicalProcess) StringFIFO() string {
	res := fmt.Sprint("LogicalProcess FIFO --> ")
	for name, pi := range p {
		res += fmt.Sprintf("\t%s: %s, ", name, pi.IncomingEvFIFO)
	}
	return res
}

func (p LPStruct) String() string {
	return fmt.Sprintf("IP: %s\t\tPort: %d\t\tFIFO: %s\n", p.IP, p.Port, p.IncomingEvFIFO)
}
