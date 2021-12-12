package simulator

import (
	"encoding/gob"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

type Node struct {
	Name	string
	Listen 	net.Listener
	Port    int
	LPs 	LogicalProcess
	Log     *LogStruct
}

type ErrorNode struct {
	Detail string
}

func (e *ErrorNode) Error() string {
	return e.Detail
}

// MakeNode : inicializar MakeNode struct
func MakeNode(name string, port int, LPs LogicalProcess, Log *LogStruct) *Node {
	gob.Register(Event{})
	// Open Listen port
	Listen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		Log.Error.Fatalf("ERROR: unable to open port: %s. Error: %s.", strconv.Itoa(port), err)
	}

	n := Node{name, Listen, port, LPs, Log}
	return &n
}

func connect(p *LPStruct, ch chan bool) {
	netAddr := fmt.Sprint(p.IP + ":" + strconv.Itoa(p.Port))
	conn, err := net.Dial("tcp", netAddr)
	for err != nil {
		time.Sleep(time.Second*3)
		conn, err = net.Dial("tcp", netAddr)
	}
	ch <- true
	_ = conn.Close()
}

func ParseFilesNames(nodeName string) (string, string) {
	nodeInd, _ := strconv.Atoi(nodeName[len(nodeName)-1:])
	lefsFile := fmt.Sprintf("6subredes.subred%d.json", nodeInd)
	netFile := fmt.Sprintf("6subredes.network.json")
	return netFile, lefsFile
}

func (n *Node) accept(ch chan bool) {
	conn, _ := n.Listen.Accept()
	//_, _ = conn.Read(nil)
	ch <- true
	_ = conn.Close()
}

/* Llamada bloqueante hasta recibir el evento algún proceso remoto */
func (n *Node) ReciveEvent(chEvent chan Event, FinishChannel chan bool) {
	//Capacidad 1000 eventos
	var sliceEvent = make([]Event, 1000)
	var conn net.Conn
	var err error
	var i = 0

	for {
		select {
		case <-FinishChannel:
			err := n.Listen.Close()
			if err != nil {
				n.Log.Error.Println("Error closing Listen")
			}
			return
		default:
		}
		if conn, err = n.Listen.Accept(); err != nil {
			n.Log.Error.Panicf("Server accept connection error: %s\n", err)
		}

		decoder := gob.NewDecoder(conn)
		err = decoder.Decode(&sliceEvent[i])
		if err != nil {
			n.Log.Error.Printf("Error decoding event: %s\n", err)
			panic(err)
		}
		n.Log.Trace.Printf("Received event: %s", sliceEvent[i])
		_ = conn.Close() //TODO: que hacer con esto
		chEvent <- sliceEvent[i]
		i = (i + 1) % 1000 //Capacidad de envio de eventos
	}
}

func (n *Node) sendEvent(e *Event, dstNodeName string) {

	var conn net.Conn
	var err error

	dstNode := n.LPs[dstNodeName]
	netAddr := fmt.Sprint(dstNode.IP + ":" + strconv.Itoa(dstNode.Port))
	conn, err = net.Dial("tcp", netAddr)
	n.Log.Trace.Printf("Sending event to: %s\n", netAddr)
	var i int
	for i = 0; err != nil && i < 5; i++ {
		n.Log.Warning.Printf("Remote connection error: %s. Retrying in %d...", err, time.Second*3)
		time.Sleep(time.Second*3)
		conn, err = net.Dial("tcp", netAddr)
	}
	if conn != nil {
		defer conn.Close()
	}
	if err != nil || conn == nil {
		if e.validateCloseEvent() {
			n.Log.Warning.Printf("Sending close event - Process [%s] gonna be closed\n", dstNodeName)
			return
		}
		n.Log.Warning.Fatalf("Event -> %s. Remote node connection error: %v\n", e, err)
		return
	}

	// Update lastTimeSent used to ensure the order and not send repetitive null events
	dstNode.LastTimeSent = e.IiTiempo
	n.LPs[dstNodeName] = dstNode

	e.Is_Sender = n.Name
	enc := gob.NewEncoder(conn)
	err = enc.Encode(e)
	for i = 0; err != nil && i < 5; i++ {
		n.Log.Warning.Printf("Error when sending event: %v. Retrying in %d...", err,time.Second*3)
		time.Sleep(time.Second*3)
		err = enc.Encode(e)
	}
	if err != nil {
		n.Log.Error.Panicf("Error when sending event: %v", err)
	}
	//n.Log.Trace.Printf("Event: %s sent to [%s]\n", e, dstNodeName)
}

func (n *Node) sendEventNetworkProcess(e *Event) {
	for nodeName, p := range n.LPs {
		if e.IiTiempo > p.LastTimeSent || (e.IiTiempo == p.LastTimeSent && !e.validateNullEvent()) || e.validateCloseEvent() {
			// Send event only if time is bigger as last sent or is equal and not a NULL event
			//n.Log.Trace.Printf("Sending ev %s to node [%s]\n", e, nodeName)
			n.sendEvent(e, nodeName)
		} else {
			if !e.validateNullEvent() {
				n.Log.Error.Panicf("Event %s out of order with LastTimeSent: [%d]\n", e, p.LastTimeSent)
			}
		}
	}
}

// Return the process with lower remote time
func (n *Node) nextStackTime() (string, LPStruct) {

	var lowerLastSafeTime TypeClock
	lowerLastSafeTime = math.MinInt32
	var lowerProc LPStruct
	var name string
	first := true
	previousHasEvents := false
	for key, proc := range n.LPs {
		partSafeTime := proc.RemoteSafeTime
		if first {
			lowerProc = proc
			lowerLastSafeTime = partSafeTime
			name = key
			first = false
			previousHasEvents = !proc.IncomingEvFIFO.emptyEventList()
		}
		/*Actualiza el tiempo si es menor que los tiempos recibidos por
		procesos remotos || si es igual pero con eventos en satck
		*/
		if partSafeTime < lowerLastSafeTime || (partSafeTime == lowerLastSafeTime && !previousHasEvents) {
			lowerProc = proc
			lowerLastSafeTime = partSafeTime
			name = key
			previousHasEvents = !proc.IncomingEvFIFO.emptyEventList()
		}
	}
	return name, lowerProc
}
