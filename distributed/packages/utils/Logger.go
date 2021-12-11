package utils

import (
	"log"
	"os"
)

const LogPath = "/home/uri/go/src/uri/Petri-DistNetwork/distributed/Logs/6subredes/"

type LogStruct struct {
	// Logs
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func InitLogs(processName string) *LogStruct {
	//Reading PLs logs file
	logFile, err := os.OpenFile(LogPath+"Log"+processName+".log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}
	//Initialization log
	Logs := LogStruct{}
	Logs.Trace = log.New(logFile,"TRACE :["+processName+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Logs.Info = log.New(logFile,"INFO :["+processName+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Logs.Warning = log.New(logFile,"WARNING :["+processName+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Logs.Error = log.New(logFile,"ERROR :["+processName+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	return &Logs
}
