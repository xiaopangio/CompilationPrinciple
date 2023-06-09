package semantic

import (
	"chap4/analyzer"
	"chap4/lexer"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Semantic 语义分析器
type Semantic struct {
	root          *analyzer.Node
	SymbolTable   map[string]*lexer.Symbol
	quadrupleList []*Quadruple
	TempVarCount  int
	LabelMap      map[string]*Label
	BoolIndex     int
	JoinIndex     int
	NotIndex      int
	RelIndex      int
}

// Label 跳转需要的标号
type Label struct {
	Name      string
	Addr      int
	BackPatch []int // 回填列表
}

// Quadruple 四元式结构
type Quadruple struct {
	op     string // 操作符
	arg1   string // 第一个操作数
	arg2   string // 第二个操作数
	result string // 结果
}

// 将四元式转换为字符串
func (q *Quadruple) String() string {
	return fmt.Sprintf("(%s, %s, %s, %s)", q.op, q.arg1, q.arg2, q.result)
}

// UpdateResult 更新四元式的结果，一般为添加跳转地址
func (q *Quadruple) UpdateResult(res string) {
	q.result = res
}

// IsTypeInt 检查是否为int类型,用于检查赋值语句，算术表达式
func (s *Semantic) IsTypeInt(id string) bool {
	return s.SymbolTable[id].Type == "int"
}

// IsTypeBool 检查是否为bool类型，用于检查布尔表达式
func (s *Semantic) IsTypeBool(id string) bool {
	return s.SymbolTable[id].Type == "bool"
}

// IsIdDeclared 检查标识符是否已经声明，用户检查赋值语句，算术表达式，布尔表达式
func (s *Semantic) IsIdDeclared(id string) bool {
	_, isDeclared := s.SymbolTable[id]
	return isDeclared
}

// IsIdTyped 检查标识符是否已经声明并且类型为int或bool
func (s *Semantic) IsIdTyped(id string) bool {
	return s.IsTypeInt(id) || s.IsTypeBool(id)
}

// IsIdArray 检查node属性map是否被初始化
func (s *Semantic) isMapMalloced(node *analyzer.Node) bool {
	return node.Attr != nil
}

// MallocAttrMap 为node的属性map分配空间
func (s *Semantic) MallocAttrMap(node *analyzer.Node) {
	if s.isMapMalloced(node) {
		return
	}
	node.Attr = make(map[string]interface{})
}

// TransferAddrOrValueAttr 将source的地址或值转移给target
func (s *Semantic) TransferAddrOrValueAttr(source, target *analyzer.Node) {
	if source.Attr["isAddr"].(bool) {
		target.Attr["isAddr"] = true
		target.Attr["addr"] = source.Attr["addr"]
	} else {
		target.Attr["isAddr"] = false
		target.Attr["value"] = source.Attr["value"]
	}
}

// randomVarName 生成一个随机的变量名
func (s *Semantic) randomVarName() string {
	s.TempVarCount++
	return "t" + strconv.Itoa(s.TempVarCount)
}

// PrintQuadrupleList 打印四元式列表
func (s *Semantic) PrintQuadrupleList() {
	for index, q := range s.quadrupleList {
		fmt.Printf("%d: %s\n", index, q)
	}
}

// PrintToFile 将四元式列表打印到文件中
func (s *Semantic) PrintToFile(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)
	for index, q := range s.quadrupleList {
		_, err := file.WriteString(fmt.Sprintf("%d: %s\n", index, q))
		if err != nil {
			return
		}
	}
}

// 生成一个新的四元式，并将其添加到列表中
func (s *Semantic) generateQuadruple(op, arg1, arg2, result string) *Quadruple {
	quadruple := &Quadruple{op, arg1, arg2, result}
	s.quadrupleList = append(s.quadrupleList, quadruple)
	return quadruple
}

// NewSemanticAnalyzer 创建一个语义分析器实例
func NewSemanticAnalyzer(root *analyzer.Node) *Semantic {
	return &Semantic{
		root:          root,
		SymbolTable:   lexer.SymbolTable,
		quadrupleList: make([]*Quadruple, 0),
		LabelMap:      make(map[string]*Label),
		TempVarCount:  0,
	}
}

// Run 运行语义分析器
func (s *Semantic) Run() {
	s.traverse()
}

// traverse 遍历语法树
func (s *Semantic) traverse() {
	if err := s.traversePROG(s.root); err != nil {
		fmt.Println(err)
		return
	}
	s.generateQuadruple("quit", "_", "_", "_")
}

// traversePROG 遍历PROG   PROG        →    {  DECLS  STMTS  }
func (s *Semantic) traversePROG(node *analyzer.Node) error {
	decls := node.LeftChild.RightBro
	stmts := decls.RightBro
	if err := s.traverseDecls(decls); err != nil {
		return err
	}
	if err := s.traverseSTMTS(stmts); err != nil {
		return err
	}
	return nil
}

// traverseSTMTS 遍历STMTS   STMTS       →    STMT  STMTS | ε
func (s *Semantic) traverseSTMTS(node *analyzer.Node) error {
	if node.LeftChild.Class == analyzer.Empty {
		return nil
	}
	stmt := node.LeftChild
	if err := s.traverseSTMT(stmt); err != nil {
		return err
	}
	stmts := stmt.RightBro
	if err := s.traverseSTMTS(stmts); err != nil {
		return err
	}
	return nil
}

// traverseDecls 遍历DECLs   DECLs        →    DECL  DECLS    |   empty
func (s *Semantic) traverseDecls(node *analyzer.Node) error {
	if node.LeftChild.Class == analyzer.Empty {
		return nil
	}
	decl := node.LeftChild
	s.MallocAttrMap(decl)
	err := s.traverseDecl(decl)
	if err != nil {
		return err
	}
	decls := decl.RightBro
	err = s.traverseDecls(decls)
	if err != nil {
		return err
	}
	return nil
}

// traverseDecl 遍历DECL   DECL        →    int  NAMES  ;  |  bool  NAMES  ;
func (s *Semantic) traverseDecl(node *analyzer.Node) error {
	switch node.LeftChild.Class {
	case analyzer.Int:
		node.Attr["type"] = "int"
	case analyzer.Bool:
		node.Attr["type"] = "bool"
	}
	names := node.LeftChild.RightBro
	s.MallocAttrMap(names)
	names.Attr["type"] = node.Attr["type"]
	err := s.traverseNames(names)
	if err != nil {
		return err
	}
	return nil
}

// traverseNames 遍历NAMES   NAMES       →    NAME  ,  NAMES  |  NAME
func (s *Semantic) traverseNames(node *analyzer.Node) error {
	name := node.LeftChild
	s.MallocAttrMap(name)
	name.Attr["type"] = node.Attr["type"]
	err := s.traverseName(name)
	if err != nil {
		return err
	}
	if name.RightBro != nil {
		names := name.RightBro.RightBro
		s.MallocAttrMap(names)
		names.Attr["type"] = node.Attr["type"]
		err := s.traverseNames(names)
		if err != nil {
			return err
		}
	}
	return nil
}

// traverseName 遍历NAME   NAME        →    id
func (s *Semantic) traverseName(node *analyzer.Node) error {
	id := node.LeftChild.Token.Value
	// 如果id已经被声明过了，那么就报错
	if s.IsIdTyped(id) {
		return errors.New("Error: " + id + " has been declared")
	}
	symbol := s.SymbolTable[id]
	symbol.Type = node.Attr["type"].(string)
	return nil
}
func (s *Semantic) traverseExpr(node *analyzer.Node) (string, error) {
	term := node.LeftChild
	s.MallocAttrMap(term)
	expr1 := node.LeftChild.RightBro
	s.MallocAttrMap(expr1)
	type1, err := s.traverseTerm(term)
	if err != nil {
		return "", err
	}
	//term继承给expr1
	s.TransferAddrOrValueAttr(term, expr1)
	type2, err := s.traverseExpr1(expr1)
	if err != nil {
		return "", err
	}
	if type1 == type2 && type1 == "int" {
		node.Attr["type"] = "int"
		//当term1推导出空时，将term综合给expr
		if expr1.LeftChild.Class == analyzer.Empty {
			s.TransferAddrOrValueAttr(term, node)
		} else {
			//当term1不为空时，将expr1综合给expr
			s.TransferAddrOrValueAttr(expr1, node)
		}
		return "int", nil
	}
	return "", errors.New("error: expr is not an int term")
}
func (s *Semantic) traverseExpr1(node *analyzer.Node) (string, error) {
	if node.LeftChild.Class == analyzer.Empty {
		node.Attr["type"] = "int"
		return "int", nil
	}
	addOp := node.LeftChild
	s.MallocAttrMap(addOp)
	term := node.LeftChild.RightBro
	s.MallocAttrMap(term)
	expr1 := node.LeftChild.RightBro.RightBro
	s.MallocAttrMap(expr1)
	type1, err := s.traverseAddop(addOp)
	if err != nil {
		return "", err
	}
	op := addOp.Attr["op"].(string)
	type2, err := s.traverseTerm(term)
	if err != nil {
		return "", err
	}
	//将term继承给expr1
	s.TransferAddrOrValueAttr(term, expr1)
	type3, err := s.traverseExpr1(expr1)
	if err != nil {
		return "", err
	}
	//处理四元式
	varName := s.randomVarName() //生成一个随机变量名
	operand1 := ""               //操作数1，由node提供
	if node.Attr["isAddr"].(bool) {
		operand1 = node.Attr["addr"].(string)
		if !s.IsIdDeclared(operand1) {
			return "", fmt.Errorf("error: %s is not declared", operand1)
		}
		if !s.IsTypeInt(operand1) {
			return "", fmt.Errorf("error: %s is not an int", operand1)
		}
	} else {
		operand1 = node.Attr["value"].(string)
	}
	operand2 := "" //操作数2，由term1提供
	if expr1.Attr["isAddr"].(bool) {
		operand2 = expr1.Attr["addr"].(string)
		if !s.IsIdDeclared(operand2) {
			return "", fmt.Errorf("error: %s is not declared", operand2)
		}
		if !s.IsTypeInt(operand2) {
			return "", fmt.Errorf("error: %s is not an int", operand2)
		}
	} else {
		operand2 = expr1.Attr["value"].(string)
	}
	//生成一条四元式
	s.generateQuadruple(op, operand1, operand2, varName)
	//进行类型检查
	if type1 == type2 && type1 == type3 && type1 == "int" {
		node.Attr["type"] = "int"
	} else {
		return "", errors.New("error: term1 is not an int term1")
	}
	//varName保存到symbolTable中
	s.SymbolTable[varName] = &lexer.Symbol{
		Name:     []byte(varName),
		Type:     "int",
		IsValued: false,
	}
	//将varName保存到node的attr中
	node.Attr["isAddr"] = true
	node.Attr["addr"] = varName
	node.Attr["type"] = "int"
	return "int", nil
}
func (s *Semantic) traverseTerm(node *analyzer.Node) (string, error) {
	nega := node.LeftChild
	s.MallocAttrMap(nega)
	term1 := node.LeftChild.RightBro
	s.MallocAttrMap(term1)
	type1, err := s.traverseNEGA(nega)
	if err != nil {
		return "", err
	}
	//将nega继承给term1
	s.TransferAddrOrValueAttr(nega, term1)
	type2, err := s.traverseTerm1(term1)
	if err != nil {
		return "", err
	}
	if type1 == type2 && type1 == "int" {
		node.Attr["type"] = "int"
		//当term1推导出空时，将nega综合给term
		if term1.LeftChild.Class == analyzer.Empty {
			s.TransferAddrOrValueAttr(nega, node)
		} else {
			//当term1不为空时，将term1综合给term
			s.TransferAddrOrValueAttr(term1, node)
		}
		return "int", nil
	}
	return "", errors.New("error: term is not an int term")
}
func (s *Semantic) traverseTerm1(node *analyzer.Node) (string, error) {
	if node.LeftChild.Class == analyzer.Empty {
		node.Attr["type"] = "int"
		return "int", nil
	}
	mulOp := node.LeftChild
	s.MallocAttrMap(mulOp)
	nega := node.LeftChild.RightBro
	s.MallocAttrMap(nega)
	term1 := node.LeftChild.RightBro.RightBro
	s.MallocAttrMap(term1)
	type1, err := s.traverseMulop(mulOp)
	op := mulOp.Attr["op"].(string)
	if err != nil {
		return "", err
	}
	type2, err := s.traverseNEGA(nega)
	if err != nil {
		return "", err
	}
	//将nega继承给term1
	s.TransferAddrOrValueAttr(nega, term1)
	type3, err := s.traverseTerm1(term1)
	if err != nil {
		return "", err
	}
	//处理四元式
	varName := s.randomVarName() //生成一个随机变量名
	operand1 := ""               //操作数1，由node提供
	if node.Attr["isAddr"].(bool) {
		operand1 = node.Attr["addr"].(string)
		if !s.IsIdDeclared(operand1) {
			return "", fmt.Errorf("error: %s is not declared", operand1)
		}
		if !s.IsTypeInt(operand1) {
			return "", fmt.Errorf("error: %s is not an int", operand1)
		}
	} else {
		operand1 = node.Attr["value"].(string)
	}
	operand2 := "" //操作数2，由term1提供
	if term1.Attr["isAddr"].(bool) {
		operand2 = term1.Attr["addr"].(string)
		if !s.IsIdDeclared(operand2) {
			return "", fmt.Errorf("error: %s is not declared", operand2)
		}
		if !s.IsTypeInt(operand2) {
			return "", fmt.Errorf("error: %s is not an int", operand2)
		}
	} else {
		operand2 = term1.Attr["value"].(string)
	}
	//生成一条四元式
	s.generateQuadruple(op, operand1, operand2, varName)
	//进行类型检查
	if type1 == type2 && type1 == type3 && type1 == "int" {
		node.Attr["type"] = "int"
	} else {
		return "", errors.New("error: term1 is not an int term1")
	}
	//varName保存到symbolTable中
	s.SymbolTable[varName] = &lexer.Symbol{
		Name:     []byte(varName),
		Type:     "int",
		IsValued: false,
	}
	//将varName保存到node的attr中
	node.Attr["isAddr"] = true
	node.Attr["addr"] = varName
	node.Attr["type"] = "int"
	return "int", nil
}
func (s *Semantic) traverseAddop(node *analyzer.Node) (string, error) {
	node.Attr["type"] = "int"
	child := node.LeftChild
	if child.Class == analyzer.Operator {
		if child.Token.Value == "+" {
			node.Attr["op"] = "+"
		} else {
			node.Attr["op"] = "-"
		}
	}
	return "int", nil
}
func (s *Semantic) traverseMulop(node *analyzer.Node) (string, error) {
	node.Attr["type"] = "int"
	child := node.LeftChild
	if child.Class == analyzer.Operator {
		if child.Token.Value == "*" {
			node.Attr["op"] = "*"
		} else {
			node.Attr["op"] = "/"
		}
	}
	return "int", nil
}

// traverseNEGA NEGA -> FACTOR | - FACTOR
func (s *Semantic) traverseNEGA(node *analyzer.Node) (string, error) {
	var F *analyzer.Node
	if node.LeftChild.Class == analyzer.FACTOR {
		F = node.LeftChild
	} else {
		F = node.LeftChild.RightBro
	}
	s.MallocAttrMap(F)
	type1, err := s.traverseFactor(F)
	if err != nil {
		return "", err
	}
	if type1 == "int" { //类型检查
		node.Attr["type"] = "int"
		if node.LeftChild.Class == analyzer.FACTOR { //NEGA -> FACTOR
			s.TransferAddrOrValueAttr(F, node)
		} else { // NEGA -> - FACTOR
			if F.Attr["isAddr"].(bool) {

				symbolName := F.Attr["addr"].(string) //获取变量名
				if !s.IsIdDeclared(symbolName) {
					return "", fmt.Errorf("error: %s is not declared", symbolName)
				}
				if !s.IsTypeInt(symbolName) {
					return "", fmt.Errorf("error: %s is not an int", symbolName)
				}
				varName := s.randomVarName() //生成一个随机变量名
				//生成中间代码
				s.generateQuadruple("*", symbolName, "-1", varName) // （*,-1,id,temp）
				//将varName保存到symbolTable中
				s.SymbolTable[varName] = &lexer.Symbol{
					Name: []byte(varName),
					Type: "int",
				}
				node.Attr["addr"] = symbolName
				node.Attr["isAddr"] = true
				return "int", nil
			} else {
				value, err := strconv.Atoi(F.Attr["value"].(string))
				if err != nil {
					return "", err
				}
				node.Attr["value"] = strconv.Itoa(-value)
				node.Attr["isAddr"] = false
			}
		}
		return "int", nil
	}
	return "", errors.New("error: nega is not an int nega")
}

// traverseFactor FACTOR -> ID | NUMBER | ( EXP )
func (s *Semantic) traverseFactor(node *analyzer.Node) (string, error) {
	switch node.LeftChild.Class {
	case analyzer.Id:
		id := node.LeftChild.Token.Value
		if !s.IsIdDeclared(id) {
			return "", fmt.Errorf("error: %s is not declared", id)
		}
		if !s.IsTypeInt(id) {
			return "", fmt.Errorf("error: %s is not an int", id)
		}
		symbol := s.SymbolTable[id]
		if symbol.Type == "int" {
			node.Attr["type"] = "int"
			node.Attr["addr"] = string(symbol.Name) //绑定符号表中该变量的名字
			node.Attr["isAddr"] = true              //该Factor是一个地址
			return "int", nil
		}
		return "", fmt.Errorf("error: %s is not an int", id)
	case analyzer.Number:
		node.Attr["type"] = "int"
		node.Attr["value"] = node.LeftChild.Token.Value //绑定该数字的值
		node.Attr["isAddr"] = false                     //该Factor是一个值
		return "int", nil
	case analyzer.LeftBracket:
		E := node.LeftChild.RightBro
		s.MallocAttrMap(E)
		type1, err := s.traverseExpr(E)
		if err != nil {
			return "", err
		}
		if type1 == "int" {
			node.Attr["type"] = "int"
			s.TransferAddrOrValueAttr(E, node)
			return "int", nil
		}
		return "", errors.New("error: expr is not an int expr")
	default:
		return "", errors.New("unknown error")
	}
}

// getBoolIndex 获取一个新的boolIndex
func (s *Semantic) getBoolIndex() int {
	s.BoolIndex++
	return s.BoolIndex
}

// getBoolLabelName 获取一个新的boolLabelName
func (s *Semantic) getBoolLabelName(index int, b bool) string {
	if b {
		return "BOOL-" + strconv.Itoa(index) + "-true"
	}
	return "BOOL-" + strconv.Itoa(index) + "-false"
}

// getBoolLabel 获取boolLabel
func (s *Semantic) getBoolLabel(index int, b bool) *Label {
	labelName := s.getBoolLabelName(index, b)
	return s.LabelMap[labelName]
}

// createBoolLabel 创建一个boolLabel
func (s *Semantic) createBoolLabel(index int, b bool) {
	labelName := s.getBoolLabelName(index, b)
	s.LabelMap[labelName] = &Label{
		Name:      labelName,
		BackPatch: make([]int, 0),
	}
}

// getJoinIndex 获取一个新的joinIndex
func (s *Semantic) getJoinIndex() int {
	s.JoinIndex++
	return s.JoinIndex
}

// getJoinLabelName 获取一个新的joinLabelName
func (s *Semantic) getJoinLabelName(index int, b bool) string {
	if b {
		return "JOIN-" + strconv.Itoa(index) + "-true"
	}
	return "JOIN-" + strconv.Itoa(index) + "-false"
}

// getJoinLabel 获取joinLabel
func (s *Semantic) getJoinLabel(index int, b bool) *Label {
	labelName := s.getJoinLabelName(index, b)
	return s.LabelMap[labelName]
}

// createJoinLabel 创建一个joinLabel
func (s *Semantic) createJoinLabel(index int, b bool) {
	labelName := s.getJoinLabelName(index, b)
	s.LabelMap[labelName] = &Label{
		Name:      labelName,
		BackPatch: make([]int, 0),
	}
}

// getNotIndex 获取一个新的notIndex
func (s *Semantic) getNotIndex() int {
	s.NotIndex++
	return s.NotIndex
}

// getNotLabelName 获取一个新的notLabelName
func (s *Semantic) getNotLabelName(index int, b bool) string {
	if b {
		return "NOT-" + strconv.Itoa(index) + "-true"
	}
	return "NOT-" + strconv.Itoa(index) + "-false"
}

// getNotLabel 获取notLabel
func (s *Semantic) getNotLabel(index int, b bool) *Label {
	labelName := s.getNotLabelName(index, b)
	return s.LabelMap[labelName]
}

// createNotLabel 创建一个notLabel
func (s *Semantic) createNotLabel(index int, b bool) {
	labelName := s.getNotLabelName(index, b)
	s.LabelMap[labelName] = &Label{
		Name:      labelName,
		BackPatch: make([]int, 0),
	}
}

// getRelIndex 获取一个新的relIndex
func (s *Semantic) getRelIndex() int {
	s.RelIndex++
	return s.RelIndex
}

// getRelLabelName 获取一个新的relLabelName
func (s *Semantic) getRelLabelName(index int, b bool) string {
	if b {
		return "REL-" + strconv.Itoa(index) + "-true"
	}
	return "REL-" + strconv.Itoa(index) + "-false"
}

// getRelLabel 获取relLabel
func (s *Semantic) getRelLabel(index int, b bool) *Label {
	labelName := s.getRelLabelName(index, b)
	return s.LabelMap[labelName]
}

// createRelLabel 创建一个relLabel
func (s *Semantic) createRelLabel(index int, b bool) {
	labelName := s.getRelLabelName(index, b)
	s.LabelMap[labelName] = &Label{
		Name:      labelName,
		BackPatch: make([]int, 0),
	}
}

// STMT      →    id  =  EXPR ;    |   id := BOOL ;
// STMT      →    if  id   then  STMT
// STMT      →    if   id   then  STMT  else STMT
// STMT      →    while   id  do  STMT
// STMT      →    {  STMTS   STMT  }
// STMT      →    read  id  ;
// STMT      →    write  id  ;
// traverseSTMT 遍历STMT
func (s *Semantic) traverseSTMT(node *analyzer.Node) error {
	if node.LeftChild.Class == analyzer.Id {
		switch node.LeftChild.RightBro.Token.Value {
		case "=":
			expr := node.LeftChild.RightBro.RightBro
			s.MallocAttrMap(expr)
			_, err := s.traverseExpr(expr)
			if err != nil {
				fmt.Println(err)
				return err
			}
			//生成中间代码
			if expr.Attr["isAddr"].(bool) {
				s.generateQuadruple("=", expr.Attr["addr"].(string), "_", node.LeftChild.Token.Value)
			} else {
				s.generateQuadruple("=", expr.Attr["value"].(string), "_", node.LeftChild.Token.Value)
			}
		case ":=":
			index := s.getBoolIndex()
			s.createBoolLabel(index, true)
			s.createBoolLabel(index, false)
			boolVar := node.LeftChild.RightBro.RightBro
			s.MallocAttrMap(boolVar)
			_, err := s.traverseBOOL(boolVar, index)
			if err != nil {
				fmt.Println(err)
				return err
			}
			//生成中间代码
			//生成bool为true时的四元式，即为id赋值为1
			s.generateQuadruple(":=", "1", "_", node.LeftChild.Token.Value)
			//拿到当前地址
			currentAddr := len(s.quadrupleList) - 1
			label := s.getBoolLabel(index, true)
			for _, index := range label.BackPatch {
				//将回填列表中的四元式的第四个元素填上label
				s.quadrupleList[index].UpdateResult(strconv.Itoa(currentAddr))
			}
			s.generateQuadruple("j", "_", "_", strconv.Itoa(len(s.quadrupleList)+2))
			//生成bool为false时的四元式，即为id赋值为0
			s.generateQuadruple(":=", "0", "_", node.LeftChild.Token.Value)
			//拿到当前地址
			currentAddr = len(s.quadrupleList) - 1
			label = s.getBoolLabel(index, false)
			for _, index := range label.BackPatch {
				//将回填列表中的四元式的第四个元素填上label
				s.quadrupleList[index].UpdateResult(strconv.Itoa(currentAddr))
			}
		default:

		}
	} else if node.LeftChild.Class == analyzer.If {
		id := node.LeftChild.RightBro
		stmt := id.RightBro.RightBro
		start := len(s.quadrupleList)
		if !s.IsIdDeclared(id.Token.Value) {
			return fmt.Errorf("id %s is not declared", id.Token.Value)
		}
		if !s.IsTypeBool(id.Token.Value) {
			return fmt.Errorf("id %s is not bool type", id.Token.Value)
		}
		s.generateQuadruple("jnz", id.Token.Value, "_", strconv.Itoa(start+2))
		s.generateQuadruple("j", "_", "_", "")
		if err := s.traverseSTMT(stmt); err != nil {
			return err
		}
		end := len(s.quadrupleList)
		s.quadrupleList[start+1].UpdateResult(strconv.Itoa(end))
		if stmt.RightBro != nil {
			s.quadrupleList[start+1].UpdateResult(strconv.Itoa(end + 1))
			start = len(s.quadrupleList)
			s.generateQuadruple("j", "_", "_", "")
			stmt = stmt.RightBro.RightBro
			if err := s.traverseSTMT(stmt); err != nil {
				return err
			}
			end = len(s.quadrupleList)
			s.quadrupleList[start].UpdateResult(strconv.Itoa(end))
		}
	} else if node.LeftChild.Class == analyzer.While {
		id := node.LeftChild.RightBro
		stmt := id.RightBro.RightBro
		if !s.IsIdDeclared(id.Token.Value) {
			return fmt.Errorf("id %s is not declared", id.Token.Value)
		}
		if !s.IsTypeBool(id.Token.Value) {
			return fmt.Errorf("id %s is not bool type", id.Token.Value)
		}
		s.MallocAttrMap(stmt)
		start := len(s.quadrupleList)
		s.generateQuadruple("jnz", id.Token.Value, "_", strconv.Itoa(start+2))
		s.generateQuadruple("j", "_", "_", "")
		if err := s.traverseSTMT(stmt); err != nil {
			return err
		}
		s.generateQuadruple("j", "_", "_", strconv.Itoa(start))
		end := len(s.quadrupleList)
		s.quadrupleList[start+1].UpdateResult(strconv.Itoa(end))
	} else if node.LeftChild.Class == analyzer.Read || node.LeftChild.Class == analyzer.Write {
		op := node.LeftChild.Token.Value
		id := node.LeftChild.RightBro
		s.generateQuadruple(op, id.Token.Value, "_", "mem")
	} else if node.LeftChild.Class == analyzer.LeftBracket {
		stmts := node.LeftChild.RightBro
		s.MallocAttrMap(stmts)
		if err := s.traverseSTMTS(stmts); err != nil {
			return err
		}
	}
	return nil
}

// 布尔表达式的赋值
// BOOL    →    JOIN  ||  BOOL    |    JOIN
// JOIN     →    NOT   &&   JOIN  |   NOT
// NOT      →    REL   |  ! REL
// REL       →    EXPR   ROP  EXPR
// ROP      →     >  |  >=  |  <  |  <=  |  ==  |   !=
// traversalBOOL 语义分析布尔表达式
func (s *Semantic) traverseBOOL(node *analyzer.Node, boolIndex int) (string, error) {
	// 为Join产生两个label，一个是true，一个是false
	index := s.getJoinIndex()
	s.createJoinLabel(index, true)
	s.createJoinLabel(index, false)
	join := node.LeftChild
	_, err := s.traverseJOIN(join, index)
	if err != nil {
		return "", err
	}
	if join.RightBro != nil {
		boolStart := len(s.quadrupleList)
		joinFalseLabel := s.getJoinLabel(index, false)
		//回填
		for _, i := range joinFalseLabel.BackPatch {
			s.quadrupleList[i].UpdateResult(strconv.Itoa(boolStart))
		}
		boolTrueLabel := s.getBoolLabel(boolIndex, true)
		joinTrueLabel := s.getJoinLabel(index, true)
		//将join的回填列表加入到bool的回填列表中
		boolTrueLabel.BackPatch = append(boolTrueLabel.BackPatch, joinTrueLabel.BackPatch...)
		boolVar := join.RightBro.RightBro
		index := s.getBoolIndex()
		s.createBoolLabel(index, true)
		s.createBoolLabel(index, false)
		_, err = s.traverseBOOL(boolVar, index)
		//将boolVar的回填列表加入到bool的回填列表中
		label1 := s.getBoolLabel(index, true)
		label2 := s.getBoolLabel(index, false)
		label3 := s.getBoolLabel(boolIndex, true)
		label4 := s.getBoolLabel(boolIndex, false)
		//将boolVar的回填列表加入到bool的回填列表中
		label3.BackPatch = append(label3.BackPatch, label1.BackPatch...)
		label4.BackPatch = append(label4.BackPatch, label2.BackPatch...)
	} else {
		//只有一个JOIN,则JOIN与BOOL的回填相同
		boolTrueLabel := s.getBoolLabel(boolIndex, true)
		boolFalseLabel := s.getBoolLabel(boolIndex, false)
		joinTrueLabel := s.getJoinLabel(index, true)
		joinFalseLabel := s.getJoinLabel(index, false)
		//将join的回填列表加入到bool的回填列表中
		boolTrueLabel.BackPatch = append(boolTrueLabel.BackPatch, joinTrueLabel.BackPatch...)
		boolFalseLabel.BackPatch = append(boolFalseLabel.BackPatch, joinFalseLabel.BackPatch...)
	}
	return "", nil
}

// traversalJOIN 语义分析JOIN
func (s *Semantic) traverseJOIN(node *analyzer.Node, joinIndex int) (string, error) {
	//	本函数与traverseBOOL类似，只是在生成四元式时，需要将&&改为||，将true改为false，false改为true
	index := s.getNotIndex()
	s.createNotLabel(index, true)
	s.createNotLabel(index, false)
	not := node.LeftChild
	_, err := s.traverseNOT(not, index)
	if err != nil {
		return "", err
	}
	if not.RightBro != nil {
		joinStart := len(s.quadrupleList)
		notTrueLabel := s.getNotLabel(index, true)
		//回填
		for _, i := range notTrueLabel.BackPatch {
			s.quadrupleList[i].UpdateResult(strconv.Itoa(joinStart))
		}
		notFalseLabel := s.getNotLabel(index, false)
		joinFalseLabel := s.getJoinLabel(joinIndex, false)
		//将not的回填列表加入到join的回填列表中
		joinFalseLabel.BackPatch = append(joinFalseLabel.BackPatch, notFalseLabel.BackPatch...)
		joinVar := not.RightBro.RightBro
		index := s.getJoinIndex()
		s.createJoinLabel(index, true)
		s.createJoinLabel(index, false)
		_, err = s.traverseJOIN(joinVar, index)
		//将joinVar的回填列表加入到join的回填列表中
		label1 := s.getJoinLabel(index, true)
		label2 := s.getJoinLabel(index, false)
		label3 := s.getJoinLabel(joinIndex, true)
		label4 := s.getJoinLabel(joinIndex, false)
		//将joinVar的回填列表加入到join的回填列表中
		label3.BackPatch = append(label3.BackPatch, label1.BackPatch...)
		label4.BackPatch = append(label4.BackPatch, label2.BackPatch...)
	} else {
		//只有一个NOT,则NOT与JOIN的回填相同
		joinTrueLabel := s.getJoinLabel(joinIndex, true)
		joinFalseLabel := s.getJoinLabel(joinIndex, false)
		notTrueLabel := s.getNotLabel(index, true)
		notFalseLabel := s.getNotLabel(index, false)
		//将not的回填列表加入到join的回填列表中
		joinTrueLabel.BackPatch = append(joinTrueLabel.BackPatch, notTrueLabel.BackPatch...)
		joinFalseLabel.BackPatch = append(joinFalseLabel.BackPatch, notFalseLabel.BackPatch...)
	}
	return "", nil
}

// traversalNOT 语义分析NOT
func (s *Semantic) traverseNOT(node *analyzer.Node, notIndex int) (string, error) {
	var rel *analyzer.Node
	if node.LeftChild.Class == analyzer.Operator {
		rel = node.LeftChild.RightBro
	} else {
		rel = node.LeftChild
	}
	index := s.getRelIndex()
	s.createRelLabel(index, true)
	s.createRelLabel(index, false)
	s.MallocAttrMap(rel)
	_, err := s.traverseREL(rel, index)
	if err != nil {
		return "", err
	}
	relTrueLabel := s.getRelLabel(index, true)
	relFalseLabel := s.getRelLabel(index, false)
	notTrueLabel := s.getNotLabel(notIndex, true)
	notFalseLabel := s.getNotLabel(notIndex, false)
	if node.LeftChild.Class == analyzer.Operator {
		//将rel的回填列表加入到not的回填列表中
		notTrueLabel.BackPatch = append(notTrueLabel.BackPatch, relFalseLabel.BackPatch...)
		notFalseLabel.BackPatch = append(notFalseLabel.BackPatch, relTrueLabel.BackPatch...)
	} else {
		//将rel的回填列表加入到not的回填列表中
		notTrueLabel.BackPatch = append(notTrueLabel.BackPatch, relTrueLabel.BackPatch...)
		notFalseLabel.BackPatch = append(notFalseLabel.BackPatch, relFalseLabel.BackPatch...)
	}
	return "", nil
}

// traversalREL 语义分析REL
func (s *Semantic) traverseREL(node *analyzer.Node, relIndex int) (string, error) {
	expr1 := node.LeftChild
	s.MallocAttrMap(expr1)
	rop := expr1.RightBro
	s.MallocAttrMap(rop)
	expr2 := rop.RightBro
	s.MallocAttrMap(expr2)
	type1, err := s.traverseExpr(expr1)
	if err != nil {
		return "", err
	}
	type2, err := s.traverseROP(rop)
	if err != nil {
		return "", err
	}
	type3, err := s.traverseExpr(expr2)
	if err != nil {
		return "", err
	}
	if type1 == "int" && type2 == "bool" && type3 == "int" {
		node.Attr["type"] = "bool"
		rop := "j" + rop.Attr["op"].(string)
		//生成中间代码
		// 操作数1
		operand1 := ""
		if expr1.Attr["isAddr"].(bool) {
			operand1 = expr1.Attr["addr"].(string)
		} else {
			operand1 = expr1.Attr["value"].(string)
		}
		// 操作数2
		operand2 := ""
		if expr2.Attr["isAddr"].(bool) {
			operand2 = expr2.Attr["addr"].(string)
		} else {
			operand2 = expr2.Attr["value"].(string)
		}
		relTrueLabel := s.getRelLabel(relIndex, true)
		relFalseLabel := s.getRelLabel(relIndex, false)
		// 生成四元式
		s.generateQuadruple(rop, operand1, operand2, relTrueLabel.Name)
		relTrueLabel.BackPatch = append(relTrueLabel.BackPatch, len(s.quadrupleList)-1)
		s.generateQuadruple("j", "_", "_", relFalseLabel.Name)
		relFalseLabel.BackPatch = append(relFalseLabel.BackPatch, len(s.quadrupleList)-1)
		return "bool", nil
	}
	return "", errors.New("error: expr1 is not an int expr")
}

// traversalROP 语义分析ROP
func (s *Semantic) traverseROP(node *analyzer.Node) (string, error) {
	node.Attr["type"] = "bool"
	node.Attr["op"] = node.LeftChild.Token.Value
	return "bool", nil
}
