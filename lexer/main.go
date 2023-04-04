// Package main  @Author xiaobaiio 2023/3/18 21:28:00
package main

import (
	"fmt"
	"log"
)

func main() {
	err := InitLexer()
	if err != nil {
		log.Fatal(err)
	}
	lexer := NewLexer()
	for i := 0; i < 6; i++ {
		filename := fmt.Sprintf("test/source/source%d.txt", i+1)
		targetFilename := fmt.Sprintf("test/target/target%d.txt", i+1)
		lexer.ReadFromFile(filename)
		lexer.Run()
		lexer.WriteToFile(targetFilename)
	}

}
