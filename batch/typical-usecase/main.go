package main

import (
	"fmt"
	"time"
)

// START EVERYTHING OMIT
func main() {
	records := make([]int, 5)
	for i := 0; i < len(records); i++ {
		processRecord(i)
	}
}

func processRecord(payload interface{}) {
	time.Sleep(1 * time.Second)
	fmt.Println("Record processed", payload)
}

// END EVERYTHING OMIT
