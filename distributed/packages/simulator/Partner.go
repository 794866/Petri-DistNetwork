package simulator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const networkPath = "/home/uri/go/src/uri/Petri-DistNetwork/distributed/testdata/"
type Partners map[string]Partner // Partner name is the key of the map with all partners

type Partner struct {
	Username       string `json:"User"`
	IP             string `json:"IP"`
	Port           int    `json:"Port"`
	IncomingEvFIFO EventList
	LookAhead      TypeClock
	RemoteSafeTime TypeClock // for incoming events
	LastTimeSent   TypeClock // for outcoming events
}

type MapTransitionNode map[IndTrans]string
type Network struct {
	Nodes        Partners          `json:"Nodes"`
	MapTransNode MapTransitionNode `json:"TransitionsMapping"`
}

func ReadPartners(networkFile string) *Network {

	// read network file json with PLs List - IP && ports
	data, err := ioutil.ReadFile(networkPath + networkFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("JSON Network file OK!")

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

func (p Partners) String() string {
	res := fmt.Sprint("Partners:")
	for k, v := range p {
		res += fmt.Sprintf("\n\t\t[%s]\t\tIP: %s\t\tPort: %d", k, v.IP, v.Port)
	}
	return res
}
func (p Partners) StringFIFO() string {
	res := fmt.Sprint("Partners FIFO --> ")
	for name, pi := range p {
		res += fmt.Sprintf("\t%s: %s, ", name, pi.IncomingEvFIFO)
	}
	return res
}

func (p Partner) String() string {
	return fmt.Sprintf("IP: %s\t\tPort: %d\t\tFIFO: %s\n", p.IP, p.Port, p.IncomingEvFIFO)
}
