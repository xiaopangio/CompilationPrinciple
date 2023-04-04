// Package analyzer  @Author xiaobaiio 2023/3/25 14:48:00
package analyzer

import "chap3/lexer"

type Node struct {
	class      int
	token      *lexer.Token
	isTerminal bool
	leftChild  *Node
	rightBro   *Node
}
