package main

import (
	"fmt"
)

// START EVERYTHING OMIT
func main() {
	s := secretData{pan: "ABCDE1234F"}
	fmt.Println("Secret", s)
	fmt.Printf("Secret %+v\n", s)
}

type secretData struct {
	pan string
}

func (s secretData) GetPan() string {
	return s.pan
}

// END EVERYTHING OMIT
