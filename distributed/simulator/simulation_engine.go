/*Package distconssim with several files to offer a distributed conservative simulation
PROPOSITO: Tipo abstracto para realizar la simulacion de una (sub)RdP.
COMENTARIOS:
	- El resultado de una simulacion local sera un slice dinamico de
	componentes, de forma que cada una de ella sera una structura estatica de
	dos enteros, el primero de ellos sera el codigo de la transiciondisparada y
	el segundo sera el valor del reloj local para el que se disparo.
*/
package simulator

import (
	"fmt"
	"time"
)

// TypeClock defines integer size for holding time.
type TypeClock int64
const LookAhead = 1

// ResultadoTransition holds fired transition id and time of firing
type ResultadoTransition struct {
	CodTransition     IndTrans
	ValorRelojDisparo TypeClock
}

// SimulationEngine is the basic data type for simulation execution
type SimulationEngine struct {
	Node               Node      // Estructura para la comunicación con el resto de nodos
	ilMisLefs          Lefs      // Estructura de datos del simulador
	iiRelojLocal       TypeClock // Valor de mi reloj local
	iiFinClk           TypeClock
	IlEventos      	   EventList //Lista de eventos a procesar
	ivTransResults     []ResultadoTransition // slice dinámico con los resultados
	EventNumber        float64   // cantidad de eventos ejecutados
	MapTransitionsNode MapTransitionNode     // diccionario con el nombre del nodo en el que se encuentra cada transición
	ChanReciveEvent    chan Event
	FinishChannel      chan bool
	Log                *LogStruct
}

// MakeMotorSimulation : inicializar SimulationEngine struct
func MakeMotorSimulation(node *Node, alLaLef Lefs, transDistr MapTransitionNode, finClk TypeClock, Log *LogStruct) *SimulationEngine {
	m := SimulationEngine{}
	m.Node = *node
	m.iiRelojLocal = 0
	m.iiFinClk = finClk
	m.ilMisLefs = alLaLef
	m.IlEventos = MakeEventList(100) //aun siendo dinámicos...
	m.ivTransResults = make([]ResultadoTransition, 0, 100)
	m.EventNumber = 0
	m.MapTransitionsNode = transDistr
	m.Log = Log
	m.ChanReciveEvent = make(chan Event, 0)
	m.FinishChannel = make(chan bool, 1)

	go m.Node.ReciveEvent(m.ChanReciveEvent, m.FinishChannel)
	return &m
}

// disparar una transicion. Esto es, generar todos los eventos
//	   ocurridos por el disparo de una transicion
//   RECIBE: Indice en el vector de la transicion a disparar
func (se *SimulationEngine) dispararTransicion(ilTr IndTrans) {
	// Prepare 5 local variables
	trList := se.ilMisLefs.IaRed              // transition list
	timeTrans := trList[ilTr].IiTiempo        // time to spread to new events
	timeDur := trList[ilTr].IiDuracionDisparo // firing time length
	listPul := trList[ilTr].TransConstPul     // Pul list of pairs Trans, Ctes
	listIul := trList[ilTr].TransConstIul     // Iul list of pairs Trans, Ctes

	// First apply Iul propagations (Inmediate : 0 propagation time)
	for _, trCo := range listIul {
		trIndLocal := trList.getLocalIndTrans(IndTrans(trCo[0]))
		// check if IUL transition belongs to local transition
		if trIndLocal == TRANS_IND_ERROR {
			//TODO: it always should be true, delete when it works
			se.Log.Error.Printf("Error: IUL transition not belongs to local transition."+
				"Node: [%s] - IUL trans: %v\n", se.Node.Name, trCo)
		}
		trList[trIndLocal].updateFuncValue(TypeConst(trCo[1]))
	}
	// Generamos eventos ocurridos por disparo de transicion ilTr
	for _, trCo := range listPul {
		// tiempo = tiempo de la transicion + coste disparo
		EventNode := se.MapTransitionsNode[IndTrans(trCo[0])]
		ev := Event{timeTrans + timeDur,
			IndTrans(trCo[0]),
			TypeConst(trCo[1]), 0, se.Node.Name}
		//Remote Transition
		if EventNode != se.Node.Name {
			se.Log.Info.Printf("Sending event %s to node [%s]\n", ev, EventNode)
			se.Node.sendEvent(&ev, EventNode)
		} else {
			se.Log.Trace.Printf("Adding local event: %s\n", ev)
			se.IlEventos.inserta(ev)
		}
	}
}

/* fireEnabledTransitions dispara todas las transiciones sensibilizadas
   		PROPOSITO: Accede a lista de transiciones sensibilizadas y procede con
	   	su disparo, lo que generara nuevos eventos y modificara el marcado de
		transicion disparada. Igualmente anotara en resultados el disparo de
		cada transicion para el reloj actual dado
*/
func (se *SimulationEngine) fireEnabledTransitions(aiLocalClock TypeClock) {
	for se.ilMisLefs.haySensibilizadas() { //while
		liCodTrans := se.ilMisLefs.getSensibilizada()
		se.dispararTransicion(liCodTrans)

		// Anotar el Resultado que disparo la liCodTrans en tiempoaiLocalClock
		se.ivTransResults = append(se.ivTransResults,
			ResultadoTransition{liCodTrans, aiLocalClock})
	}
}

// tratarEventos : Accede a lista eventos y trata todos con tiempo aiTiempo
func (se *SimulationEngine) tratarEvento(ev *Event) {
	if ev.getTiempo() == se.iiRelojLocal {
		se.Log.Info.Printf("Processing next event: %s .....\n", ev)
		trList := se.ilMisLefs.IaRed // obtener lista de transiciones de Lefs
		localIndTrans := trList.getLocalIndTrans(ev.IiTransicion)
		if localIndTrans == TRANS_IND_ERROR {
			se.Log.Error.Panicf("Cannot find transition id [%d] in >> %v <<", ev.IiTransicion, trList)
		}
		// Establecer nuevo valor de la funcion
		trList[localIndTrans].updateFuncValue(ev.IiCte)
		// Establecer nuevo valor del tiempo
		trList[localIndTrans].actualizaTiempo(ev.IiTiempo)

		se.EventNumber++
	} else {
		se.Log.Error.Panicf("Processing event in other time: event time: %d - local time: %d", ev.IiTiempo, se.iiRelojLocal)
	}
}

// avanzarTiempo : Modifica reloj local con minimo tiempo de entre
//	   recibidos del exterior o del primer evento en lista de eventos
func (se *SimulationEngine) avanzarTiempo() TypeClock {
	nextTime := se.IlEventos.tiempoPrimerEvento()
	fmt.Println("NEXT CLOCK...... : ", nextTime)
	return nextTime
}

// devolverResultados : Mostrar los resultados de la simulacion
func (se SimulationEngine) devolverResultados() {
	se.Log.Info.Println("----------------------------------------")
	se.Log.Info.Println("Resultados del simulador local")
	se.Log.Info.Println("----------------------------------------")
	if len(se.ivTransResults) == 0 {
		se.Log.Info.Println("No esperes ningun resultado...")
	}

	for _, liResult := range se.ivTransResults {
		se.Log.Info.Printf("TIEMPO: %d  -> TRANSICION: %d\n", liResult.ValorRelojDisparo, liResult.CodTransition)
	}

	se.Log.Info.Printf("========== TOTAL DE TRANSICIONES DISPARADAS = %d\n", len(se.ivTransResults))
}

/* Devuelve el evento con menor tiempo, nil si no hay eventos pendientes en ese momento
		true = local - false = remote*/
func (se *SimulationEngine) getFirstEvent() (*Event, bool) {
	_, lowestTimeNode := se.Node.nextStackTime()
	// There is any local event
	if se.IlEventos.emptyEventList() && lowestTimeNode.IncomingEvFIFO.emptyEventList() {
		// No events from retarded node
		return nil, false //no hay eventos pendientes en este momento
	}
	if se.IlEventos.emptyEventList() {
		ev := lowestTimeNode.IncomingEvFIFO.leePrimerEvento()
		return &ev, false //evento remoto
	}

	// Get local event
	localEv := se.IlEventos.leePrimerEvento()
	se.Log.Trace.Printf("getFirstEvent: LOCAL->%s, lowerFIFOtime: %d\n", localEv, lowestTimeNode.RemoteSafeTime)
	// No events in lazy node FIFO
	if lowestTimeNode.IncomingEvFIFO.emptyEventList() && localEv.IiTiempo > lowestTimeNode.RemoteSafeTime {
		return nil, false //no hay eventos pendientes en este momento
	}
	if lowestTimeNode.IncomingEvFIFO.emptyEventList(){
		return &localEv, true //local event
	}

	// Events in lazy node FIFO
	remoteEvent := lowestTimeNode.IncomingEvFIFO.leePrimerEvento()
	if localEv.IiTiempo <= remoteEvent.IiTiempo {
		return &localEv, true
	}
	return &remoteEvent, false
}

// Get the lowerEvent. If it not exists, blocks until new event arrive. If it is an event and is the lowest return it.
// If not, blocks again until receive the lowset and all FIFO have at least one event. If the recv event is null with
// lower time, blocks again until receives the correct one.

/* */

func (se *SimulationEngine) getNextEvent() *Event {
	// Iterate until get a processable event or finish event
	for {
		ev, isLocalEv := se.getFirstEvent()
		// Not blocked, I've get an event to process
		if ev != nil {
			// delete event for list
			if isLocalEv {
				se.IlEventos.eliminaPrimerEvento()
				se.Log.Info.Printf("Lower event is local: %s\n", ev)
			} else {
				name, remoteNode := se.Node.nextStackTime()
				remoteNode.IncomingEvFIFO.eliminaPrimerEvento()
				se.Node.LPs[name] = remoteNode
				se.Log.Info.Printf("Lower event is remote: %s\n", ev)
			}
			return ev
		}

		// I'm gonna to block, send before it an NULL event
		_, lowestNodeTime := se.Node.nextStackTime()
		lowestTime := lowestNodeTime.RemoteSafeTime
		clkFstEvLocal := se.IlEventos.leePrimerEvento().IiTiempo
		if lowestTime > clkFstEvLocal && !se.IlEventos.emptyEventList() { // Time on NULL event depend on local events
			lowestTime = clkFstEvLocal
		}
		// Send null event or finish
		nextTime := lowestTime + LookAhead
		if lowestTime >= se.iiFinClk || nextTime >= se.iiFinClk {
			return se.FinishSim()
		}
		nullEv := Event{
			IiTransicion: 0,
			IiCte:        0,
			Is_Sender:    se.Node.Name,
			IiTiempo:     nextTime,
			Ib_IsNULL:    1,
		}
		se.Log.Trace.Printf("Sending NULL event: %s\n", nullEv)
		if nullEv.getTiempo() <= se.iiFinClk {
			se.Node.sendEventNetworkProcess(&nullEv)
		}

		// Wait for a event or a null message
		se.Log.Trace.Printf("	Waiting for incoming event...\n")
		recvEv := <-se.ChanReciveEvent

		// Process event
		if recvEv.validateCloseEvent() {
			se.Log.Warning.Printf("Received closing event %s\n", recvEv)
			time.Sleep(time.Second*3)
			se.FinishChannel <- true
			return &recvEv
		} else if recvEv.validateNullEvent() {
			// Update RemoteSafeTime
			sender := se.Node.LPs[recvEv.getSource()]
			sender.RemoteSafeTime = recvEv.IiTiempo
			se.Node.LPs[recvEv.getSource()] = sender

			// Check if any event has been unblocked
			se.Log.Trace.Printf("NULL received: %s\n", recvEv)
		} else { // Event can be processed
			// Insert event in remote node FIFO
			se.Log.Trace.Printf("Adding received event to incFIFO: %s\n", recvEv)
			senderNode := se.Node.LPs[recvEv.getSource()]
			senderNode.RemoteSafeTime = recvEv.IiTiempo
			senderNode.IncomingEvFIFO.inserta(recvEv)
			se.Node.LPs[recvEv.getSource()] = senderNode
		}
	}
}

// SimularUnpaso de una RdP con duración disparo >= 1. Devuelve true si se ha procesado el ultimo evento
func (se *SimulationEngine) simularUnpaso() bool {

	se.Log.Info.Printf("-----------TIME: %d \n", se.iiRelojLocal)
	se.ilMisLefs.ImprimeLefs()
	se.ilMisLefs.actualizaSensibilizadas(se.iiRelojLocal)
	se.Log.Trace.Println("-----------Stack de transiciones sensibilizadas---------")
	se.ilMisLefs.TransSensib.ImprimeTransStack(se.Log)
	se.Log.Trace.Println("-----------Final Stack de transiciones-----------")

	// Fire enabled transitions and produce events
	if se.ilMisLefs.haySensibilizadas() {
		se.fireEnabledTransitions(se.iiRelojLocal)
	}

	se.Log.Trace.Println("-----------Lista eventos después de disparos-----------")
	se.Log.Trace.Printf("Eventos locales: %s\n", se.IlEventos)
	se.Log.Trace.Println(se.Node.LPs.StringFIFO())
	se.Log.Trace.Println("-----------Final lista eventos-----------")

	ev := se.getNextEvent()
	if ev.validateCloseEvent() {
		return true
	} else if ev != nil {
		// advance local clock to soonest available event
		se.iiRelojLocal = ev.IiTiempo
		se.Log.Trace.Printf("NEXT CLOCK...... : %d\n", se.iiRelojLocal)

		// if events exist for current local clock, process them
		se.tratarEvento(ev)
		return false
	} else {
		se.Log.Error.Panicf("Simulating nil event\n")
		return true
	}
}

func (se *SimulationEngine) FinishSim() *Event {
	// Send closing event to LogicalProcess
	ev := Event{
		IiTiempo:     se.iiRelojLocal,
		IiTransicion: FINISH_EVENT,
		IiCte:        0,
		Ib_IsNULL:    0,
	}
	se.Log.Info.Println("Close Event")
	se.Node.sendEventNetworkProcess(&ev)
	time.Sleep(time.Second*3)
	se.FinishChannel <- true
	return &ev
}

// SimularPeriodo de una RdP
// RECIBE: - Ciclo inicial (por si marcado recibido no se corresponde al
//				inicial sino a uno obtenido tras simular ai_cicloinicial ciclos)
//		   - Ciclo con el que terminamos
func (se *SimulationEngine) SimularPeriodo() {
	ldIni := time.Now()
	finish := false
	initClk := se.iiRelojLocal
	for se.iiRelojLocal < se.iiFinClk && !finish {
		finish = se.simularUnpaso()
	}

	if !finish {
		se.FinishSim()
	}
	elapsedTime := time.Since(ldIni)

	se.Log.Info.Printf("Eventos por segundo = %.4f\n",
		se.EventNumber/elapsedTime.Seconds())

	// Devolver los resultados de la simulacion
	se.devolverResultados()
	se.Log.Info.Println("---------------------")
	se.Log.Info.Printf("TIEMPO SIMULADO en ciclos: %d\n", se.iiFinClk-initClk)
	se.Log.Info.Printf("TIEMPO ejecución REAL simulación: %s\n", elapsedTime.String())
}
