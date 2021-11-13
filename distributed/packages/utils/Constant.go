package utils

import "time"

const HomePath = "/home/uri/"
const AbsWorkPath = "/home/uri/go/src/uri/Petri-DistNetwork/distributed/"
const RelOutputPath = "results/"
const LoggerPath = AbsWorkPath + "results/"
const RelTestDataPath = "testdata/"
const MaxAttempsConnect = 5
const PeriodRetry = 2 * time.Second // second between connection retries
const BinFilePath = "distributed/distributed"
const MaxEventsQueueCap = 1000
const TimeWaitStop = 3 * time.Second
