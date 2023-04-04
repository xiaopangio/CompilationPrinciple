// Package analyzer  @Author xiaobaiio 2023/3/25 14:50:00
package analyzer

import (
	"chap3/lexer"
	"fmt"
	"log"
	"os"
	"strings"
)

type Analyzer struct {
	source     []*lexer.Token
	index      int
	token      *lexer.Token
	treeSource []string
	root       *Node
}

func NewAnalyzer(source []*lexer.Token) *Analyzer {
	return &Analyzer{
		source: source,
	}
}

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

func init() {
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

//算法表达式文法
//E->T E1
//E1->ADDOP T E1
//E1-> 空
//T-> NEGA T1
//T1->MULOP NEGA T1
//T1-> 空
// NEGA -> - F | F
//F->id | number | (E)

func (a *Analyzer) Analyse() {
	node, err := a.PROG()
	if err != nil {
		log.Printf("error :%w", err)
		return
	}
	a.root = node
}
func (a *Analyzer) GetToken() bool {
	if a.index >= len(a.source) {
		return false
	}
	a.token = a.source[a.index]
	a.index++
	return true
}

// E E->T E1
func (a *Analyzer) E() (*Node, error) {
	node := &Node{
		class: EXPR,
	}
	t, err := a.T()
	if err != nil {
		return nil, err
	}
	e1, err := a.E1()
	if err != nil {
		return nil, err
	}
	node.leftChild = t
	t.rightBro = e1
	return node, nil
}

// E1 E1->ADDOP T E1 、E1-> 空
func (a *Analyzer) E1() (*Node, error) {
	node := &Node{
		class: EXPR1,
	}
	addOp, err := a.AddOp()
	if err != nil {
		node.leftChild = &Node{class: Empty}
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
	node.leftChild = addOp
	addOp.rightBro = t
	t.rightBro = e1
	return node, nil
}

// T -> NEGA T1
func (a *Analyzer) T() (*Node, error) {
	node := &Node{
		class: TERM,
	}
	nega, err := a.NEGA()
	if err != nil {
		return nil, err
	}
	t1, err := a.T1()
	if err != nil {
		return nil, err
	}
	node.leftChild = nega
	nega.rightBro = t1
	return node, nil
}

// NEGA -> - F | F
func (a *Analyzer) NEGA() (*Node, error) {
	node := &Node{class: NEGA}
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Operator:
		if string(a.token.Value) == "-" {
			minus := &Node{class: Operator, token: a.token, isTerminal: true}
			f, err := a.F()
			if err != nil {
				return nil, err
			}
			node.leftChild = minus
			minus.rightBro = f
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
	node.leftChild = f
	return node, nil
}

// T1 ->MULOP NEGA T1 | empty
func (a *Analyzer) T1() (*Node, error) {
	node := &Node{
		class: TERM1,
	}
	mulOp, err := a.MulOp()
	if err != nil {
		node.leftChild = &Node{class: Empty}
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
	node.leftChild = mulOp
	mulOp.rightBro = nega
	nega.rightBro = t1
	return node, nil
}

//F->id | number | (E)

func (a *Analyzer) F() (*Node, error) {
	lastIndex := a.index
	node := &Node{
		class: FACTOR,
	}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Identifier:
		node.leftChild = &Node{
			class:      Id,
			token:      a.token,
			isTerminal: true,
		}
		return node, nil
	case lexer.IntConst:
		node.leftChild = &Node{
			class:      Number,
			token:      a.token,
			isTerminal: true,
		}
		return node, nil
	case lexer.Separator:
		if a.token.Value[0] == byte('(') {
			leftBracket := &Node{class: LeftBracket, token: a.token, isTerminal: true}
			e, err := a.E()
			if err != nil {
				a.index = lastIndex
				return nil, err
			}
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")

			}
			if a.token.Class == lexer.Separator && a.token.Value[0] == byte(')') {
				rightBracket := &Node{class: RightBracket, token: a.token, isTerminal: true}
				node.leftChild = leftBracket
				leftBracket.rightBro = e
				e.rightBro = rightBracket
				return node, nil
			} else {
				a.index = lastIndex
				return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
			}
		} else {
			a.index = lastIndex
			return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
		}
	default:
		a.index = lastIndex
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
}

// AddOp ADDOP -> + | -
func (a *Analyzer) AddOp() (*Node, error) {
	lastIndex := a.index
	node := &Node{
		class: ADDOP,
	}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")

	}
	if a.token.Class != lexer.Operator {
		a.index = lastIndex
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
	if len(a.token.Value) != 1 {
		a.index = lastIndex
		return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
	}
	switch a.token.Value[0] {
	case byte('+'), byte('-'):
		node.leftChild = &Node{
			class:      Operator,
			token:      a.token,
			isTerminal: true,
		}
		return node, nil
	default:
		a.index = lastIndex
		return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
	}
}
func (a *Analyzer) MulOp() (*Node, error) {
	lastIndex := a.index
	node := &Node{
		class: MULOP,
	}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	if a.token.Class != lexer.Operator {
		a.index = lastIndex
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
	if len(a.token.Value) != 1 {
		a.index = lastIndex
		return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
	}
	switch a.token.Value[0] {
	case byte('*'), byte('/'):
		node.leftChild = &Node{
			class:      Operator,
			token:      a.token,
			isTerminal: true,
		}
		return node, nil
	default:
		a.index = lastIndex
		return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
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
	node := &Node{class: BOOL}
	join, err := a.JOIN()
	if err != nil {
		return nil, err
	}
	node.leftChild = join
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		a.index = tempIndex
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Operator:
		if string(a.token.Value) == "||" {
			or := &Node{class: Operator, token: a.token, isTerminal: true}
			boolvar, err := a.BOOL()
			if err != nil {
				a.index = tempIndex
				return node, nil
			}
			join.rightBro = or
			or.rightBro = boolvar
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
	node := &Node{class: JOIN}
	not, err := a.NOT()
	if err != nil {
		return nil, err
	}
	node.leftChild = not
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Operator:
		if string(a.token.Value) == "&&" {
			and := &Node{class: Operator, token: a.token, isTerminal: true}
			join, err := a.JOIN()
			if err != nil {
				a.index = tempIndex
				return node, nil
			}
			not.rightBro = and
			and.rightBro = join
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

// NOT      →    REL   |  ! REL
func (a *Analyzer) NOT() (*Node, error) {
	node := &Node{class: NOT}
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Operator:
		if string(a.token.Value) == "!" {
			node.leftChild = &Node{class: Operator, token: a.token, isTerminal: true}
			rel, err := a.REL()
			if err != nil {
				a.index = tempIndex
				return nil, err
			}
			node.leftChild.rightBro = rel
			return node, nil
		} else {
			a.index = tempIndex
			rel, err := a.REL()
			if err != nil {
				return nil, err
			}
			node.leftChild = rel
			return node, nil
		}
	default:
		a.index = tempIndex
		rel, err := a.REL()
		if err != nil {
			return nil, err
		}
		node.leftChild = rel
		return node, nil
	}
}

// REL       →    EXPR   ROP  EXPR
func (a *Analyzer) REL() (*Node, error) {
	node := &Node{class: REL}
	expr, err := a.E()
	if err != nil {
		return nil, err
	}
	node.leftChild = expr
	rop, err := a.ROP()
	if err != nil {
		return nil, err
	}
	node.leftChild.rightBro = rop
	expr, err = a.E()
	if err != nil {
		return nil, err
	}
	rop.rightBro = expr
	return node, nil
}

// ROP      →     >  |  >=  |  <  |  <=  |  ==  |   !=
func (a *Analyzer) ROP() (*Node, error) {
	node := &Node{class: ROP}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Operator:
		isBoolOperator := lexer.IsBoolOperator(string(a.token.Value))
		if !isBoolOperator {
			return nil, fmt.Errorf("unexpeted token value: %s", a.token.Value)
		}
		node.leftChild = &Node{class: Operator, token: a.token, isTerminal: true}
		return node, nil
	default:
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
}

//PROG        →    {  DECLS  STMTS  }
//DECLS       →    DECL  DECLS    |   empty
//DECL         →    int  NAMES  ;  |  bool  NAMES  ;
//NAMES     →    NAME ,  NAMES  |  NAME
//NAME       →    id
//STMTS    →    STMT  STMTS  |   STMT
//STMT      →    id  =  EXPR ;    |   id := BOOL ;
//STMT      →    if  id   then  STMT
//STMT      →    if   id   then  STMT  else STMT
//STMT      →    while   id  do  STMT
//STMT      →    {  STMTS   STMT  }
//STMT      →    read  id  ;
//STMT      →    write  id  ;

// PROG   →    {  DECLS  STMTS  }
func (a *Analyzer) PROG() (*Node, error) {
	node := &Node{class: PROG}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Separator:
		if string(a.token.Value) == "{" {
			leftBracket := &Node{class: LeftBracket, token: a.token, isTerminal: true}
			decls, err := a.DECLS()
			if err != nil {
				return nil, err
			}
			stmts, err := a.STMTS()
			if err != nil {
				return nil, err
			}
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if string(a.token.Value) != "}" {
				return nil, fmt.Errorf("unexpected token value: %s", a.token.Value)
			}
			rightBracket := &Node{class: RightBracket, token: a.token, isTerminal: true}
			leftBracket.rightBro = decls
			decls.rightBro = stmts
			stmts.rightBro = rightBracket
			node.leftChild = leftBracket
			return node, nil
		} else {
			return nil, fmt.Errorf("unexpected token value: %s", a.token.Value)
		}
	default:
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
}

// STMTS    →    STMT  STMTS  |   empty
func (a *Analyzer) STMTS() (*Node, error) {
	node := &Node{class: STMTS}
	tempIndex := a.index
	stmt, err := a.STMT()
	if err != nil {
		a.index = tempIndex
		node.leftChild = &Node{class: Empty}
		return node, nil
	}
	node.leftChild = stmt
	stmts, err := a.STMTS()
	if err != nil {
		return node, nil
	}
	stmt.rightBro = stmts
	return node, nil
}

// STMT      →    id  =  EXPR ;    |   id := BOOL ;
// STMT      →    if  id   then  STMT
// STMT      →    if   id   then  STMT  else STMT
// STMT      →    while   id  do  STMT
// STMT      →    {  STMTS   STMT  }
// STMT      →    read  id  ;
// STMT      →    write  id  ;
func (a *Analyzer) STMT() (*Node, error) {
	node := &Node{class: STMT}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Identifier:
		id := &Node{class: Id, token: a.token, isTerminal: true}
		node.leftChild = id
		if ok := a.GetToken(); !ok {
			return nil, fmt.Errorf("analysis of the end")
		}
		if a.token.Class != lexer.Operator {
			return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
		}
		switch string(a.token.Value) {
		case "=":
			equal := &Node{class: Operator, token: a.token, isTerminal: true}
			id.rightBro = equal
			expr, err := a.E()
			if err != nil {
				return nil, err
			}
			equal.rightBro = expr
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Separator || string(a.token.Value) != ";" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			semicolon := &Node{class: Separator, token: a.token, isTerminal: true}
			expr.rightBro = semicolon
			return node, nil
		case ":=":
			assign := &Node{class: Operator, token: a.token, isTerminal: true}
			id.rightBro = assign
			boolvar, err := a.BOOL()
			if err != nil {
				return nil, err
			}
			assign.rightBro = boolvar
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Separator || string(a.token.Value) != ";" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			semicolon := &Node{class: Separator, token: a.token, isTerminal: true}
			boolvar.rightBro = semicolon
			return node, nil
		default:
			return nil, fmt.Errorf("unexpected token value: %s", a.token.Value)
		}
	case lexer.Keyword:
		switch string(a.token.Value) {
		case "if":
			ifvar := &Node{class: If, token: a.token, isTerminal: true}
			node.leftChild = ifvar
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Identifier {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			id := &Node{class: Id, token: a.token, isTerminal: true}
			ifvar.rightBro = id
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Keyword || string(a.token.Value) != "then" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			then := &Node{class: Then, token: a.token, isTerminal: true}
			id.rightBro = then
			stmt, err := a.STMT()
			if err != nil {
				return nil, err
			}
			then.rightBro = stmt
			tempIndex := a.index
			if ok := a.GetToken(); !ok {
				a.index = tempIndex
				return node, nil
			}
			if a.token.Class != lexer.Keyword || string(a.token.Value) != "else" {
				a.index = tempIndex
				return node, nil
			}
			elsevar := &Node{class: Else, token: a.token, isTerminal: true}
			stmt.rightBro = elsevar
			stmt, err = a.STMT()
			if err != nil {
				return nil, err
			}
			elsevar.rightBro = stmt
			return node, nil
		case "while":
			whilevar := &Node{class: While, token: a.token, isTerminal: true}
			node.leftChild = whilevar
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Identifier {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			id := &Node{class: Id, token: a.token, isTerminal: true}
			whilevar.rightBro = id
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Keyword || string(a.token.Value) != "do" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			do := &Node{class: Do, token: a.token, isTerminal: true}
			id.rightBro = do
			stmt, err := a.STMT()
			if err != nil {
				return nil, err
			}
			do.rightBro = stmt
			return node, nil
		case "read":
			read := &Node{class: Read, token: a.token, isTerminal: true}
			node.leftChild = read
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			id := &Node{class: Id, token: a.token, isTerminal: true}
			read.rightBro = id
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Separator || string(a.token.Value) != ";" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			id.rightBro = &Node{class: Separator, token: a.token, isTerminal: true}
			return node, nil
		case "write":
			write := &Node{class: Write, token: a.token, isTerminal: true}
			node.leftChild = write
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			id := &Node{class: Id, token: a.token, isTerminal: true}
			write.rightBro = id
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Separator || string(a.token.Value) != ";" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			id.rightBro = &Node{class: Separator, token: a.token, isTerminal: true}
			return node, nil
		default:
			return nil, fmt.Errorf("unexpected token class: %v", a.token.Value)
		}
	case lexer.Separator:
		switch string(a.token.Value) {
		case "{":
			leftBracket := &Node{class: LeftBracket, token: a.token, isTerminal: true}
			node.leftChild = leftBracket
			stmts, err := a.STMTS()
			if err != nil {
				return nil, err
			}
			leftBracket.rightBro = stmts
			if ok := a.GetToken(); !ok {
				return nil, fmt.Errorf("analysis of the end")
			}
			if a.token.Class != lexer.Separator || string(a.token.Value) != "}" {
				return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
			}
			rightBracket := &Node{class: RightBracket, token: a.token, isTerminal: true}
			stmts.rightBro = rightBracket
			return node, nil
		default:
			return nil, fmt.Errorf("unexpected token class: %v", a.token.Value)
		}
	default:
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
}

// DECLS       →    DECL  DECLS    |   empty
func (a *Analyzer) DECLS() (*Node, error) {
	node := &Node{class: DECLS}
	tempIndex := a.index
	decl, err := a.DECL()
	if err != nil {
		a.index = tempIndex
		node.leftChild = &Node{class: Empty}
		return node, nil
	}
	node.leftChild = decl
	decls, err := a.DECLS()
	if err != nil {
		return nil, err
	}
	decl.rightBro = decls
	return node, nil
}

// DECL         →    int  NAMES  ;  |  bool  NAMES  ;
func (a *Analyzer) DECL() (*Node, error) {
	node := &Node{class: DECL}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	if a.token.Class != lexer.Keyword {
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
	switch string(a.token.Value) {
	case "int":
		node.leftChild = &Node{class: Int, token: a.token, isTerminal: true}
	case "bool":
		node.leftChild = &Node{class: Bool, token: a.token, isTerminal: true}
	default:
		return nil, fmt.Errorf("unexpected token value: %s", a.token.Value)
	}
	names, err := a.NAMES()
	if err != nil {
		return nil, err
	}
	node.leftChild.rightBro = names
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	if a.token.Class != lexer.Separator || string(a.token.Value) != ";" {
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
	names.rightBro = &Node{class: Separator, token: a.token, isTerminal: true}
	return node, nil
}

// NAMES     →    NAME ,  NAMES  |  NAME
func (a *Analyzer) NAMES() (*Node, error) {
	node := &Node{class: NAMES}
	name, err := a.NAME()
	if err != nil {
		return nil, err
	}
	node.leftChild = name
	tempIndex := a.index
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	if a.token.Class != lexer.Separator || string(a.token.Value) != "," {
		a.index = tempIndex
		return node, nil
	}
	colon := &Node{class: Separator, token: a.token, isTerminal: true}
	names, err := a.NAMES()
	if err != nil {
		return node, nil
	}
	name.rightBro = colon
	colon.rightBro = names
	return node, nil
}

// NAME      →    id
func (a *Analyzer) NAME() (*Node, error) {
	node := &Node{class: NAME}
	if ok := a.GetToken(); !ok {
		return nil, fmt.Errorf("analysis of the end")
	}
	switch a.token.Class {
	case lexer.Identifier:
		node.leftChild = &Node{class: Id, isTerminal: true, token: a.token}
		return node, nil
	default:
		return nil, fmt.Errorf("unexpected token class: %v", a.token.Class)
	}
}
func (a *Analyzer) PrintTree() {
	a.genTreeString()
	for _, v := range a.treeSource {
		fmt.Println(v)
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
	if node.isTerminal {
		a.treeSource[row] += string(node.token.Value)
	} else {
		a.treeSource[row] += ConstMap[node.class]
	}
	//打印当前节点下的竖线
	row++
	column += len(ConstMap[node.class]) / 2
	if node.leftChild != nil {
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
		if node.leftChild.isTerminal {
			column -= 1
		} else {
			column -= len(ConstMap[node.class]) / 2
		}
		column = a.deal(node.leftChild, row, column)
	}
	//打印当前节点的右兄弟节点
	row = tempRow
	if node.rightBro == nil {
		return max(column, len(a.treeSource[row]))
	}
	if len(a.treeSource[row]) < column {
		a.treeSource[row] += strings.Repeat("-", column-len(a.treeSource[row])+1)
	} else {
		a.treeSource[row] += strings.Repeat("-", 5)
	}
	column = len(a.treeSource[row])
	column = a.deal(node.rightBro, row, column)
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
	defer f.Close()
	for _, v := range a.treeSource {
		f.WriteString(v + "\n")
	}
	return nil
}
