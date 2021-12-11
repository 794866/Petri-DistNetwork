package utils

import "time"

const RelTestDataPath = "testdata/"
const MaxAttempsConnect = 5
const PeriodRetry = 2 * time.Second // second between connection retries
const BinFilePath = "distributed/distributed"
const MaxEventsQueueCap = 1000
const TimeWaitStop = 3 * time.Second
