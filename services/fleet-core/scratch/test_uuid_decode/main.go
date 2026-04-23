package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type Test struct {
	ID *uuid.UUID `json:"id"`
}

func main() {
	var t Test
	err := json.Unmarshal([]byte(`{"id":""}`), &t)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", t.ID)
	}
}
