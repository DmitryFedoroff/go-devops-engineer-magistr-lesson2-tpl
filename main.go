package main

import (
	"fmt"
	"os"

	"github.com/DmitryFedoroff/go-devops-engineer-magistr-lesson2-tpl/yamlvalidator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: yamlvalidator <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("%s does not exist\n", filename)
		os.Exit(1)
	}

	v, err := yamlvalidator.NewValidator(filename)
	if err != nil {
		fmt.Printf("Error initializing validator: %v\n", err)
		os.Exit(1)
	}

	errors := v.Validate()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Println("YAML file is valid")
}
