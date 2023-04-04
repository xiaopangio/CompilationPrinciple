// Package main  @Author xiaobaiio 2023/3/25 14:30:00
package main

import (
	"chap3/analyzer"
	"chap3/lexer"
	"log"
)

func main() {
	lexer.InitLexer()
	lexer := lexer.NewLexer()
	err := lexer.ReadFromFile("source.txt")
	if err != nil {
		log.Fatal(err)
	}
	lexer.Run()

	//lexer.Print()
	lexer.Print()
	analyzer := analyzer.NewAnalyzer(lexer.Target())
	analyzer.Analyse()
	analyzer.PrintTree()
	analyzer.PrintToFile("tree.txt")
}
