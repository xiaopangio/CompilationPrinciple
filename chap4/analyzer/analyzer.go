// Package analyzer  @Author xiaobaiio 2023/3/25 14:50:00
package analyzer

import (
	"chap4/lexer"
	"fmt"
	"os"
	"strings"
)

const (
	// EXPR 算术表达式文法
	EXPR = iota + 1
	EXPR1
	TERM
	TERM1
	ADDOP
	MULOP
	Empty
	FACTOR
	Id
	Number
	Operator
	LeftBracket
	RightBracket
	// BOOL BOOL表达式文法
	BOOL
	JOIN
	NOT
	REL
	ROP
	PROG
	DECLS
	STMTS
	DECL
	NAMES
	NAME
	STMT
	Int
	Bool
	Separator
	If
	While
	Then
	Else
	Do
	Read
	Write
	NEGA
)

var ConstMap = make(map[int]string)
var (
	EndErr = fmt.Errorf("analysis of the end")
)

type Analyzer struct {
	source     []*lexer.Token
	index      int
	token      *lexer.Token
	treeSource []string
	root       *Node
	err        error
}

func NewAnalyzer(source []*lexer.Token) *Analyzer {
	return &Analyzer{
		source: source,
	}
}
func (a *Analyzer) GetRoot() *Node {
	return a.root
}
func InitAnalyzer() {
	ConstMap[EXPR] = "<EXPR>"
	ConstMap[EXPR1] = "<EXPR1>"
	ConstMap[TERM] = "<TERM>"
	ConstMap[TERM1] = "<TERM1>"
	ConstMap[ADDOP] = "<ADDOP>"
	ConstMap[MULOP] = "<MULOP>"
	ConstMap[FACTOR] = "<FACTOR>"
	ConstMap[Empty] = "empty"
	ConstMap[BOOL] = "<BOOL>"
	ConstMap[JOIN] = "<JOIN>"
	ConstMap[NOT] = "<NOT>"
	ConstMap[REL] = "<REL>"
	ConstMap[ROP] = "<ROP>"
	ConstMap[PROG] = "<PROG>"
	ConstMap[DECLS] = "<DECLS>"
	ConstMap[STMTS] = "<STMTS>"
	ConstMap[DECL] = "<DECL>"
	ConstMap[NAMES] = "<NAMES>"
	ConstMap[NAME] = "<NAME>"
	ConstMap[STMT] = "<STMT>"
	ConstMap[NEGA] = "<NEGA>"
}
func classError(class, index int) error {
	return fmt.Errorf("unexpected Token Class: %v,Token index: %d", class, index)
}
func valueError(value string, index int) error {
	return fmt.Errorf("unexpected Token value: %s,Token index: %d", value, index)
}
func (a *Analyzer) Analyse() {
	node, err := a.PROG()
	a.root = node
	a.err = err
}
func (a *Analyzer) GetToken() bool {
	if a.index >= len(a.source) {
		return false
	}
	a.token = a.source[a.index]
	a.index++
	return true
}

//算法表达式文法
//E->T E1
//E1->ADDOP T E1
//E1-> 空
//T-> NEGA T1
//T1->MULOP NEGA T1
//T1-> 空
// NEGA -> - F | F
//F->id | number | (E)

// E E->T E1
func (a *Analyzer) E() (*Node, error) {
	node := &Node{
		Class: EXPR,
	}
	t, err := a.T()
	if err != nil {
		return nil, err
	}
	e1, err := a.E1()
	if err != nil {
		return nil, err
	}
	node.LeftChild = t
	t.RightBro = e1
	return node, nil
}

// E1 E1->ADDOP T E1 、E1-> 空
func (a *Analyzer) E1() (*Node, error) {
	node := &Node{
		Class: EXPR1,
	}
	addOp, err := a.AddOp()
	if err != nil {
		node.LeftChild = &Node{Class: Empty}
		return node, nil
	}
	t, err := a.T()
	if err != nil {
		return nil, err
	}
	e1, err := a.E1()
	if err != nil {
		return nil, err
	}
	node.LeftChild = addOp
	addOp.RightBro = t
	t.RightBro = e1
	return node, nil
}

// T -> NEGA T1
func (a *Analyzer) T() (*Node, error) {
	node := &Node{
		Class: TERM,
	}
	nega, err := a.NEGA()
	if err != nil {
		return nil, err
	}
	t1, err := a.T1()
	if err != nil {
		return nil, err
	}
	node.LeftChild = nega
	nega.RightBro = t1
	return node, nil
}

// NEGA -> - F | F
func (a *Analyzer) NEGA() (*Node, error) {
	node := &Node{Class: NEGA}
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Operator:
		if a.token.Value == "-" {
			minus := &Node{Class: Operator, Token: a.token, IsTerminal: true}
			f, err := a.F()
			if err != nil {
				return nil, err
			}
			node.LeftChild = minus
			minus.RightBro = f
			return node, nil
		} else {
			a.index = tempIndex
		}
	default:
		a.index = tempIndex
	}
	f, err := a.F()
	if err != nil {
		return nil, err
	}
	node.LeftChild = f
	return node, nil
}

// T1 ->MULOP NEGA T1 | empty
func (a *Analyzer) T1() (*Node, error) {
	node := &Node{
		Class: TERM1,
	}
	mulOp, err := a.MulOp()
	if err != nil {
		node.LeftChild = &Node{Class: Empty}
		return node, nil
	}
	nega, err := a.NEGA()
	if err != nil {
		return nil, err
	}
	t1, err := a.T1()
	if err != nil {
		return nil, err
	}
	node.LeftChild = mulOp
	mulOp.RightBro = nega
	nega.RightBro = t1
	return node, nil
}

//F->id | number | (E)

func (a *Analyzer) F() (*Node, error) {
	lastIndex := a.index
	node := &Node{
		Class: FACTOR,
	}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Identifier:
		node.LeftChild = &Node{
			Class:      Id,
			Token:      a.token,
			IsTerminal: true,
		}
		return node, nil
	case lexer.IntConst:
		node.LeftChild = &Node{
			Class:      Number,
			Token:      a.token,
			IsTerminal: true,
		}
		return node, nil
	case lexer.Separator:
		if a.token.Value == "(" {
			leftBracket := &Node{Class: LeftBracket, Token: a.token, IsTerminal: true}
			e, err := a.E()
			if err != nil {
				a.index = lastIndex
				return nil, err
			}
			if ok := a.GetToken(); !ok {
				return nil, EndErr

			}
			if a.token.Class == lexer.Separator && a.token.Value == ")" {
				rightBracket := &Node{Class: RightBracket, Token: a.token, IsTerminal: true}
				node.LeftChild = leftBracket
				leftBracket.RightBro = e
				e.RightBro = rightBracket
				return node, nil
			} else {
				a.index = lastIndex
				return nil, valueError(a.token.Value, a.index)
			}
		} else {
			a.index = lastIndex
			return nil, valueError(a.token.Value, a.index)
		}
	default:
		a.index = lastIndex
		return nil, classError(a.token.Class, a.index)
	}
}

// AddOp ADDOP -> + | -
func (a *Analyzer) AddOp() (*Node, error) {
	lastIndex := a.index
	node := &Node{
		Class: ADDOP,
	}
	if ok := a.GetToken(); !ok {
		return nil, EndErr

	}
	if a.token.Class != lexer.Operator {
		a.index = lastIndex
		return nil, classError(a.token.Class, a.index)
	}
	if len(a.token.Value) != 1 {
		a.index = lastIndex
		return nil, valueError(a.token.Value, a.index)
	}
	switch a.token.Value {
	case "+", "-":
		node.LeftChild = &Node{
			Class:      Operator,
			Token:      a.token,
			IsTerminal: true,
		}
		return node, nil
	default:
		a.index = lastIndex
		return nil, valueError(a.token.Value, a.index)
	}
}
func (a *Analyzer) MulOp() (*Node, error) {
	lastIndex := a.index
	node := &Node{
		Class: MULOP,
	}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	if a.token.Class != lexer.Operator {
		a.index = lastIndex
		return nil, classError(a.token.Class, a.index)
	}
	if len(a.token.Value) != 1 {
		a.index = lastIndex
		return nil, valueError(a.token.Value, a.index)
	}
	switch a.token.Value {
	case "*", "/":
		node.LeftChild = &Node{
			Class:      Operator,
			Token:      a.token,
			IsTerminal: true,
		}
		return node, nil
	default:
		a.index = lastIndex
		return nil, valueError(a.token.Value, a.index)
	}
}

//布尔表达式的赋值
//BOOL    →    JOIN  ||  BOOL    |    JOIN
//JOIN     →    NOT   &&   JOIN  |   NOT
//NOT      →    REL   |  ! REL
//REL       →    EXPR   ROP  EXPR
//ROP      →     >  |  >=  |  <  |  <=  |  ==  |   !=

// BOOL    →    JOIN  ||  BOOL    |    JOIN
func (a *Analyzer) BOOL() (*Node, error) {
	node := &Node{Class: BOOL}
	join, err := a.JOIN()
	if err != nil {
		return nil, err
	}
	node.LeftChild = join
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		a.index = tempIndex
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Operator:
		if a.token.Value == "||" {
			or := &Node{Class: Operator, Token: a.token, IsTerminal: true}
			boolvar, err := a.BOOL()
			if err != nil {
				a.index = tempIndex
				return node, nil
			}
			join.RightBro = or
			or.RightBro = boolvar
			return node, nil
		} else {
			a.index = tempIndex
			return node, nil
		}
	default:
		a.index = tempIndex
		return node, nil
	}
}

// JOIN     →    NOT   &&   JOIN  |   NOT
func (a *Analyzer) JOIN() (*Node, error) {
	node := &Node{Class: JOIN}
	not, err := a.NOT()
	if err != nil {
		return nil, err
	}
	node.LeftChild = not
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Operator:
		if a.token.Value == "&&" {
			and := &Node{Class: Operator, Token: a.token, IsTerminal: true}
			join, err := a.JOIN()
			if err != nil {
				a.index = tempIndex
				return node, nil
			}
			not.RightBro = and
			and.RightBro = join
			return node, nil
		} else {
			a.index = tempIndex
			return node, nil
		}
	default:
		a.index = tempIndex
		return node, nil
	}
}

// NOT      →    REL   |  ! REL | ! id | id
func (a *Analyzer) NOT() (*Node, error) {
	node := &Node{Class: NOT}
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Operator:
		if a.token.Value == "!" {
			node.LeftChild = &Node{Class: Operator, Token: a.token, IsTerminal: true}
			tIndex := a.index
			rel, err := a.REL()
			if err != nil {
				a.index = tIndex
				if ok := a.GetToken(); !ok {
					return nil, EndErr
				}
				if a.token.Class != lexer.Identifier {
					a.index = tempIndex
					return nil, err
				} else {
					node.LeftChild.RightBro = &Node{Class: Id, Token: a.token, IsTerminal: true}
					return node, nil
				}
			}
			node.LeftChild.RightBro = rel
			return node, nil
		} else {
			a.index = tempIndex
			rel, err := a.REL()
			if err != nil {
				return nil, err
			}
			node.LeftChild = rel
			return node, nil
		}
	default:
		a.index = tempIndex
		rel, err := a.REL()
		if err != nil {
			a.index = tempIndex
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Identifier {
				return nil, err
			} else {
				node.LeftChild = &Node{Class: Id, Token: a.token, IsTerminal: true}
				return node, nil
			}
		}
		node.LeftChild = rel
		return node, nil
	}
}

// REL       →    EXPR   ROP  EXPR
func (a *Analyzer) REL() (*Node, error) {
	node := &Node{Class: REL}
	expr, err := a.E()
	if err != nil {
		return nil, err
	}
	node.LeftChild = expr
	rop, err := a.ROP()
	if err != nil {
		return nil, err
	}
	node.LeftChild.RightBro = rop
	expr, err = a.E()
	if err != nil {
		return nil, err
	}
	rop.RightBro = expr
	return node, nil
}

// ROP      →     >  |  >=  |  <  |  <=  |  ==  |   !=
func (a *Analyzer) ROP() (*Node, error) {
	node := &Node{Class: ROP}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Operator:
		isBoolOperator := lexer.IsBoolOperator(a.token.Value)
		if !isBoolOperator {
			return nil, valueError(a.token.Value, a.index)
		}
		node.LeftChild = &Node{Class: Operator, Token: a.token, IsTerminal: true}
		return node, nil
	default:
		return nil, classError(a.token.Class, a.index)
	}
}

//PROG        →    {  DECLS  STMTS  }
//DECLS       →    DECL  DECLS    |   empty
//DECL         →    int  NAMES  ;  |  bool  NAMES  ;
//NAMES     →    NAME ,  NAMES  |  NAME
//NAME       →    id
//STMTS    →    STMT  STMTS  |   empty
//STMT      →    id  =  EXPR ;    |   id := BOOL ;
//STMT      →    if  id   then  STMT
//STMT      →    if   id   then  STMT  else STMT
//STMT      →    while   id  do  STMT
//STMT      →    {  STMTS   STMT  }
//STMT      →    read  id  ;
//STMT      →    write  id  ;

// PROG   →    {  DECLS  STMTS  }
func (a *Analyzer) PROG() (*Node, error) {
	node := &Node{Class: PROG}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Separator:
		if a.token.Value == "{" {
			leftBracket := &Node{Class: LeftBracket, Token: a.token, IsTerminal: true}
			decls, err := a.DECLS()
			if err != nil {
				return nil, err
			}
			stmts, err := a.STMTS()
			if err != nil {
				return nil, err
			}
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Value != "}" {
				return nil, valueError(a.token.Value, a.index)
			}
			rightBracket := &Node{Class: RightBracket, Token: a.token, IsTerminal: true}
			leftBracket.RightBro = decls
			decls.RightBro = stmts
			stmts.RightBro = rightBracket
			node.LeftChild = leftBracket
			return node, nil
		} else {
			return nil, valueError(a.token.Value, a.index)
		}
	default:
		return nil, classError(a.token.Class, a.index)
	}
}

// STMTS    →    STMT  STMTS  |   empty
func (a *Analyzer) STMTS() (*Node, error) {
	node := &Node{Class: STMTS}
	tempIndex := a.index
	stmt, err := a.STMT()
	if err != nil {
		a.index = tempIndex
		node.LeftChild = &Node{Class: Empty}
		return node, nil
	}
	node.LeftChild = stmt
	stmts, err := a.STMTS()
	if err != nil {
		return node, nil
	}
	stmt.RightBro = stmts
	return node, nil
}

// STMT      →    id  =  EXPR ;    |   id := BOOL ;
// STMT      →    if  id   then  STMT
// STMT      →    if   id   then  STMT  else STMT
// STMT      →    while   id  do  STMT
// STMT      →    {  STMTS }
// STMT      →    read  id  ;
// STMT      →    write  id  ;
func (a *Analyzer) STMT() (*Node, error) {
	node := &Node{Class: STMT}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Identifier:
		id := &Node{Class: Id, Token: a.token, IsTerminal: true}
		node.LeftChild = id
		if ok := a.GetToken(); !ok {
			return nil, EndErr
		}
		if a.token.Class != lexer.Operator {
			return nil, classError(a.token.Class, a.index)
		}
		switch a.token.Value {
		case "=":
			equal := &Node{Class: Operator, Token: a.token, IsTerminal: true}
			id.RightBro = equal
			expr, err := a.E()
			if err != nil {
				return nil, err
			}
			equal.RightBro = expr
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Separator || a.token.Value != ";" {
				return nil, classError(a.token.Class, a.index)
			}
			semicolon := &Node{Class: Separator, Token: a.token, IsTerminal: true}
			expr.RightBro = semicolon
			return node, nil
		case ":=":
			assign := &Node{Class: Operator, Token: a.token, IsTerminal: true}
			id.RightBro = assign
			boolvar, err := a.BOOL()
			if err != nil {
				return nil, err
			}
			assign.RightBro = boolvar
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Separator || a.token.Value != ";" {
				return nil, classError(a.token.Class, a.index)
			}
			semicolon := &Node{Class: Separator, Token: a.token, IsTerminal: true}
			boolvar.RightBro = semicolon
			return node, nil
		default:
			return nil, valueError(a.token.Value, a.index)
		}
	case lexer.Keyword:
		switch a.token.Value {
		case "if":
			ifvar := &Node{Class: If, Token: a.token, IsTerminal: true}
			node.LeftChild = ifvar
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Identifier {
				return nil, classError(a.token.Class, a.index)
			}
			id := &Node{Class: Id, Token: a.token, IsTerminal: true}
			ifvar.RightBro = id
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Keyword || a.token.Value != "then" {
				return nil, classError(a.token.Class, a.index)
			}
			then := &Node{Class: Then, Token: a.token, IsTerminal: true}
			id.RightBro = then
			stmt, err := a.STMT()
			if err != nil {
				return nil, err
			}
			then.RightBro = stmt
			tempIndex := a.index
			if ok := a.GetToken(); !ok {
				a.index = tempIndex
				return node, nil
			}
			if a.token.Class != lexer.Keyword || a.token.Value != "else" {
				a.index = tempIndex
				return node, nil
			}
			elsevar := &Node{Class: Else, Token: a.token, IsTerminal: true}
			stmt.RightBro = elsevar
			stmt, err = a.STMT()
			if err != nil {
				return nil, err
			}
			elsevar.RightBro = stmt
			return node, nil
		case "while":
			whilevar := &Node{Class: While, Token: a.token, IsTerminal: true}
			node.LeftChild = whilevar
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Identifier {
				return nil, classError(a.token.Class, a.index)
			}
			id := &Node{Class: Id, Token: a.token, IsTerminal: true}
			whilevar.RightBro = id
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Keyword || a.token.Value != "do" {
				return nil, classError(a.token.Class, a.index)
			}
			do := &Node{Class: Do, Token: a.token, IsTerminal: true}
			id.RightBro = do
			stmt, err := a.STMT()
			if err != nil {
				return nil, err
			}
			do.RightBro = stmt
			return node, nil
		case "read":
			read := &Node{Class: Read, Token: a.token, IsTerminal: true}
			node.LeftChild = read
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Identifier {
				return nil, classError(a.token.Class, a.index)
			}
			id := &Node{Class: Id, Token: a.token, IsTerminal: true}
			read.RightBro = id
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Separator || a.token.Value != ";" {
				return nil, classError(a.token.Class, a.index)
			}
			id.RightBro = &Node{Class: Separator, Token: a.token, IsTerminal: true}
			return node, nil
		case "write":
			write := &Node{Class: Write, Token: a.token, IsTerminal: true}
			node.LeftChild = write
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			id := &Node{Class: Id, Token: a.token, IsTerminal: true}
			write.RightBro = id
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Separator || a.token.Value != ";" {
				return nil, classError(a.token.Class, a.index)
			}
			id.RightBro = &Node{Class: Separator, Token: a.token, IsTerminal: true}
			return node, nil
		default:
			return nil, valueError(a.token.Value, a.index)
		}
	case lexer.Separator:
		switch a.token.Value {
		case "{":
			leftBracket := &Node{Class: LeftBracket, Token: a.token, IsTerminal: true}
			node.LeftChild = leftBracket
			stmts, err := a.STMTS()
			if err != nil {
				return nil, err
			}
			leftBracket.RightBro = stmts
			if ok := a.GetToken(); !ok {
				return nil, EndErr
			}
			if a.token.Class != lexer.Separator || a.token.Value != "}" {
				return nil, classError(a.token.Class, a.index)
			}
			rightBracket := &Node{Class: RightBracket, Token: a.token, IsTerminal: true}
			stmts.RightBro = rightBracket
			return node, nil
		default:
			return nil, classError(a.token.Class, a.index)
		}
	default:
		return nil, classError(a.token.Class, a.index)
	}
}

// DECLS       →    DECL  DECLS    |   empty
func (a *Analyzer) DECLS() (*Node, error) {
	node := &Node{Class: DECLS}
	tempIndex := a.index
	decl, err := a.DECL()
	if err != nil {
		a.index = tempIndex
		node.LeftChild = &Node{Class: Empty}
		return node, nil
	}
	node.LeftChild = decl
	decls, err := a.DECLS()
	if err != nil {
		return nil, err
	}
	decl.RightBro = decls
	return node, nil
}

// DECL         →    int  NAMES  ;  |  bool  NAMES  ;
func (a *Analyzer) DECL() (*Node, error) {
	node := &Node{Class: DECL}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	if a.token.Class != lexer.Keyword {
		return nil, classError(a.token.Class, a.index)
	}
	switch a.token.Value {
	case "int":
		node.LeftChild = &Node{Class: Int, Token: a.token, IsTerminal: true}
	case "bool":
		node.LeftChild = &Node{Class: Bool, Token: a.token, IsTerminal: true}
	default:
		return nil, valueError(a.token.Value, a.index)
	}
	names, err := a.NAMES()
	if err != nil {
		return nil, err
	}
	node.LeftChild.RightBro = names
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	if a.token.Class != lexer.Separator || a.token.Value != ";" {
		return nil, classError(a.token.Class, a.index)
	}
	names.RightBro = &Node{Class: Separator, Token: a.token, IsTerminal: true}
	return node, nil
}

// NAMES     →    NAME ,  NAMES  |  NAME
func (a *Analyzer) NAMES() (*Node, error) {
	node := &Node{Class: NAMES}
	name, err := a.NAME()
	if err != nil {
		return nil, err
	}
	node.LeftChild = name
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	if a.token.Class != lexer.Separator || a.token.Value != "," {
		a.index = tempIndex
		return node, nil
	}
	colon := &Node{Class: Separator, Token: a.token, IsTerminal: true}
	names, err := a.NAMES()
	if err != nil {
		return node, nil
	}
	name.RightBro = colon
	colon.RightBro = names
	return node, nil
}

// NAME      →    id
func (a *Analyzer) NAME() (*Node, error) {
	node := &Node{Class: NAME}
	if ok := a.GetToken(); !ok {
		return nil, EndErr
	}
	switch a.token.Class {
	case lexer.Identifier:
		node.LeftChild = &Node{Class: Id, IsTerminal: true, Token: a.token}
		return node, nil
	default:
		return nil, classError(a.token.Class, a.index)
	}
}
func (a *Analyzer) PrintTree() {
	a.genTreeString()
	for _, v := range a.treeSource {
		fmt.Println(v)
	}
	if a.err != nil {
		fmt.Println(a.err.Error())
	}

}
func (a *Analyzer) genTreeString() {
	a.treeSource = make([]string, 0)
	row := 0
	column := 0
	a.deal(a.root, row, column)
}
func (a *Analyzer) deal(node *Node, row, column int) int {
	tempRow := row
	if node == nil {
		return column
	}
	if len(a.treeSource) <= row {
		a.treeSource = append(a.treeSource, "")
	}
	//打印当前节点
	if len(a.treeSource[row]) < column {
		a.treeSource[row] += strings.Repeat(" ", column-len(a.treeSource[row])+1)
	}
	if node.IsTerminal {
		a.treeSource[row] += node.Token.Value
	} else {
		a.treeSource[row] += ConstMap[node.Class]
	}
	//打印当前节点下的竖线
	row++
	column += len(ConstMap[node.Class]) / 2
	if node.LeftChild != nil {
		if len(a.treeSource) <= row {
			a.treeSource = append(a.treeSource, "")
		}
		if len(a.treeSource[row]) < column {
			count := column - len(a.treeSource[row])
			a.treeSource[row] += strings.Repeat(" ", count)
			a.treeSource[row] += "|"
		} else {
			a.treeSource[row] += "|"
		}
		//打印当前节点的左儿子节点
		row++
		if node.LeftChild.IsTerminal {
			column -= 1
		} else {
			column -= len(ConstMap[node.Class]) / 2
		}
		column = a.deal(node.LeftChild, row, column)
	}
	//打印当前节点的右兄弟节点
	row = tempRow
	if node.RightBro == nil {
		return max(column, len(a.treeSource[row]))
	}
	if len(a.treeSource[row]) < column {
		a.treeSource[row] += strings.Repeat("-", column-len(a.treeSource[row])+1)
	} else {
		a.treeSource[row] += strings.Repeat("-", 5)
	}
	column = len(a.treeSource[row])
	column = a.deal(node.RightBro, row, column)
	return column
}
func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func (a *Analyzer) PrintToFile(filename string) error {
	a.genTreeString()
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)
	for _, v := range a.treeSource {
		_, err := f.WriteString(v + "\n")
		if err != nil {
			return err
		}
	}
	if a.err != nil {
		_, err := f.WriteString(a.err.Error())
		if err != nil {
			return err
		}
	}
	return nil
}
