// Package main  @Author xiaobaiio 2023/3/25 14:30:00
package main

import (
	"chap4/analyzer"
	"chap4/lexer"
	"chap4/semantic"
)

func main() {
	source := "text/source.txt"
	target := "text/target.txt"
	Run(source, target)
}
func Run(readFile, writeFile string) {
	if err := lexer.InitLexer(); err != nil {
		return
	}

	lexer := lexer.NewLexer()
	if err := lexer.ReadFromFile(readFile); err != nil {
		return
	}
	lexer.Run()
	//lexer.Print()
	analyzer.InitAnalyzer()
	analyzer := analyzer.NewAnalyzer(lexer.Target())
	analyzer.Analyse()
	//analyzer.PrintTree()
	//analyzer.PrintToFile(writeFile)
	semanticAnalyzer := semantic.NewSemanticAnalyzer(analyzer.GetRoot())
	semanticAnalyzer.Run()
	semanticAnalyzer.PrintToFile(writeFile)
}
