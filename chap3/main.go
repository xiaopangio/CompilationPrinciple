// Package main  @Author xiaobaiio 2023/3/25 14:30:00
package main

import (
	"chap3/analyzer"
	"chap3/lexer"
	"log"
	"strconv"
)

func main() {
	sourcePrefix := "test/source"
	sourceSuffix := ".txt"
	targetPrefix := "test/target"
	targetSuffix := ".txt"
	for i := 1; i < 9; i++ {
		source := sourcePrefix + strconv.Itoa(i) + sourceSuffix
		target := targetPrefix + strconv.Itoa(i) + targetSuffix
		Run(source, target)
	}
	//source := sourcePrefix + "8" + sourceSuffix
	//target := targetPrefix + "8" + targetSuffix
	//Run(source, target)
}
func Run(readFile, writeFile string) {
	lexer.InitLexer()
	lexer := lexer.NewLexer()
	err := lexer.ReadFromFile(readFile)
	if err != nil {
		log.Fatal(err)
	}
	lexer.Run()
	//lexer.Print()
	//lexer.Print()
	analyzer.InitAnalyzer()
	analyzer := analyzer.NewAnalyzer(lexer.Target())
	analyzer.Analyse()
	//analyzer.PrintTree()
	analyzer.PrintToFile(writeFile)
}
