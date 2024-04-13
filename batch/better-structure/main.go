package main

import (
	"fmt"
	"sync"
	"time"
)

type process struct {
	Record interface{}
	Result string
	Index  int
}

type processor interface {
	Process(record interface{}) string
}

type processorImpl struct{}

func NewProcessor() processor {
	return processorImpl{}
}

func (p processorImpl) Process(record interface{}) string {
	time.Sleep(1 * time.Second)
	return fmt.Sprintf("processed: %+v", record)
}

// START MAIN OMIT
func main() {
	records := make([]int, 100)
	t := time.Now()
	fmt.Println("starting at", t)
	ps := make([]process, len(records))
	for i := range records {
		ps[i] = process{Record: i, Index: i}
	}
	total := len(records)
	fmt.Println("Total records: ", total)
	spawners := 40
	if total < spawners {
		spawners = total
	}
	pro := NewProcessor()

	executor := NewBatchExecutor(pro, total, spawners)
	executor.Initialize()
	executor.Process(ps)
	fmt.Println("Finished at: ", time.Since(t))
}

// END MAIN OMIT

type BatchExecutor struct {
	wg            *sync.WaitGroup
	poc           processor
	collectorChan chan process
	resultChan    chan []process
	noOfBatches   int
	total         int
	buffer        []chan process
}

// START EXECUTOR OMIT
type Executor interface {
	Initialize()
	Process(records []process) []process
}

func NewBatchExecutor(p processor, total int, noOfBatches int) Executor {
	wg := sync.WaitGroup{}
	wg.Add(total)

	buffer := make([]chan process, noOfBatches)
	return &BatchExecutor{poc: p, wg: &wg, collectorChan: make(chan process),
		resultChan: make(chan []process), noOfBatches: noOfBatches, buffer: buffer, total: total}
}

// END EXECUTOR OMIT

// START EXECUTORIMP OMIT
func (b *BatchExecutor) Initialize() {
	go b.collector()
	for i := range b.buffer {
		b.buffer[i] = make(chan process)
		go b.spawner(b.buffer[i])
	}
}

func (b *BatchExecutor) Process(records []process) []process {
	for i := range records {
		b.buffer[i%b.noOfBatches] <- process{Record: i, Index: i}
		if i == b.total-1 {
			break
		}
	}
	b.wg.Wait()
	for _, bufCh := range b.buffer {
		close(bufCh)
	}
	close(b.collectorChan)
	return <-b.resultChan
}

// END EXECUTORIMP OMIT

func (b BatchExecutor) collector() {
	list := []process{}
	for dt := range b.collectorChan {
		list = append(list, dt)
	}
	b.resultChan <- list
}

func (b *BatchExecutor) spawner(ch chan process) {
	for rec := range ch {
		time.Sleep(100 * time.Millisecond)
		resp := b.poc.Process(rec.Record)
		fmt.Println("Processed: ", rec.Index+1)
		rec.Result = resp
		b.collectorChan <- rec
		b.wg.Done()
	}
}
