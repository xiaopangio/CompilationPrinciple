// Package analyzer  @Author xiaobaiio 2023/3/25 14:48:00
package analyzer

import "chap4/lexer"

type Node struct {
	Class      int
	Token      *lexer.Token
	IsTerminal bool
	LeftChild  *Node
	RightBro   *Node
	Attr       map[string]any
}
