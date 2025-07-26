package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("=== MINIMAL TEST ===")
	fmt.Printf("Args: %v\n", os.Args)
	fmt.Printf("PORT env: %s\n", os.Getenv("PORT"))
	fmt.Printf("PWD: %s\n", os.Getenv("PWD"))
	
	// Test if we can create a file
	if err := os.WriteFile("/tmp/test", []byte("hello"), 0644); err != nil {
		log.Printf("Cannot write to /tmp: %v", err)
	} else {
		fmt.Println("File write test: OK")
	}
	
	fmt.Println("=== END TEST ===")
}