// Package lexer  @Author xiaobaiio 2023/3/18 14:44:00
package lexer

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	Identifier = iota + 1 //标识符
	IntConst              //整常数
	BoolConst             //布尔常数
	Keyword               //关键字
	Operator              //运算符
	Separator             //分隔符
)
const (
	space      = byte(32)
	enter      = byte(13)
	newLine    = byte(10)
	slash      = byte(47)
	star       = byte(42)
	singleLine = 1
	multiLine  = 2
)
const (
	InitState = iota
	IDState
	NumberState
	SeparatorState
	OperatorState
	ErrState
)

var (
	IdentifierTooLongErr   = errors.New("Identifier's length greater than 8")
	NumberTooLongErr       = errors.New("number's length greater than 8")
	NumberStartWithZeroErr = errors.New("number start with zero")
	InvalidOperatorErr     = errors.New("invalid Operator")
	InvalidCharacter       = errors.New("invalid character")
)

var (
	keywordList   []string
	operatorList  []byte
	separatorList []byte
)
var (
	validOperators []string
)
var classMap map[int]string
var SymbolTable map[string]*Symbol
var boolOperatorList = []string{
	">=",
	"<=",
	"==",
	"!=",
	"||",
	"&&",
	"!",
	">",
	"<",
}

type Symbol struct {
	Name     []byte //变量的名字
	Kind     int    //标识符类型，值为varKind，表明是一个变量标识符。
	Class    int    // 种别
	Value    string //值
	Type     string //数据类型 int or bool
	IsValued bool   //是否有值
}

type Lexer struct {
	source []byte
	target []*Token
	err    error
}
type Token struct {
	Class int
	Value string
}

func InitLexer() error {
	initPrefix, _ := os.Getwd()
	keywords, err := os.ReadFile(initPrefix + "/lexer/init/Keyword.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "Keyword.txt")
		return err
	}
	keywordList = strings.Split(string(keywords), ",")
	operatorList, err = os.ReadFile(initPrefix + "/lexer/init/Operator.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "Operator.txt")
		return err
	}
	separatorList, err = os.ReadFile(initPrefix + "/lexer/init/Separator.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "Separator.txt")
		return err
	}
	data, err := os.ReadFile(initPrefix + "/lexer/init/validOperator.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "validOperator.txt")
		return err
	}
	validOperators = strings.Split(string(data), ",")
	classMap = make(map[int]string)
	classMap[Identifier] = "Identifier"
	classMap[IntConst] = "IntConst"
	classMap[BoolConst] = "BoolConst"
	classMap[Keyword] = "Keyword"
	classMap[Operator] = "Operator"
	classMap[Separator] = "Separator"
	SymbolTable = make(map[string]*Symbol)
	return nil
}
func NewLexer() *Lexer {
	return &Lexer{}
}

func (l *Lexer) ReadFromFile(filename string) error {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("cannot read source code from file: ", filename)
		return err
	}
	l.source = bytes
	return nil
}
func (l *Lexer) WriteToFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		log.Fatal("cannot open target  file: ", filename)
		return err
	}
	writer := bufio.NewWriter(file)
	for i, token := range l.target {
		var item string
		if i != len(l.target)-1 {
			item = fmt.Sprintf("(%s,'%s'),", classMap[token.Class], token.Value)
		} else {
			item = fmt.Sprintf("(%s,'%s')", classMap[token.Class], token.Value)
		}
		if i != 0 && i%5 == 0 {
			writer.Write([]byte("\n"))
		}
		writer.Write([]byte(item))
	}
	if l.err != nil {
		writer.Write([]byte("\n"))
		e := fmt.Sprintf("error: %s", l.err.Error())
		writer.Write([]byte(e))
	}
	writer.Flush()
	return nil
}

// Clean 数据清洗
func (l *Lexer) clean() error {
	var pre byte
	var cleanSpace bool
	var cleanAnnotation bool
	var cleanAnnotationType int
	var cleanTarget []byte
	for _, b := range l.source {
		switch b {
		case enter:
			pre = enter
		case newLine:
			if cleanAnnotation {
				if cleanAnnotationType == multiLine {
					continue
				} else if cleanAnnotationType == singleLine {
					cleanAnnotation = false
					cleanAnnotationType = 0
				} else {
					return errors.New("cleanAnnotationType is invalid: " + strconv.Itoa(cleanAnnotationType))
				}
			}
			//if pre == enter {
			cleanSpace = true
			//} else {
			//	return errors.New("after enter is not a newline")
			//}
		case space:
			if cleanAnnotation {
				continue
			}
			if !cleanSpace {
				cleanTarget = append(cleanTarget, space)
			}
		case slash:
			if pre == slash {
				cleanAnnotation = true
				cleanAnnotationType = singleLine
				pre = 0
			} else if pre == star {
				cleanAnnotation = false
				cleanAnnotationType = 0
			} else {
				pre = slash
			}
		case star:
			if pre == slash {
				cleanAnnotation = true
				cleanAnnotationType = multiLine
				pre = 0
			} else {
				if cleanAnnotation {
					pre = star
				} else {
					cleanTarget = append(cleanTarget, star)
				}
			}
		default:
			if cleanAnnotation {
				continue
			}
			cleanTarget = append(cleanTarget, b)
			if cleanSpace {
				cleanSpace = false
				pre = 0
			}
		}
	}
	l.source = cleanTarget
	//fmt.Printf("%s\n", l.source)
	return nil
}

// Run 进行词法分析
func (l *Lexer) Run() {
	err := l.clean()
	if err != nil {
		l.err = err
		log.Println(err)
		return
	}
	err = l.parse()
	if err != nil {
		l.err = err
		log.Println(err)
	}
}
func (l *Lexer) Print() {
	for _, token := range l.target {
		fmt.Printf("class: %10s , value: %s\n", classMap[token.Class], token.Value)
	}
	if l.err != nil {
		fmt.Println(l.err)
	}
}
func (l *Lexer) Target() []*Token {
	return l.target
}

// pares 对clean后的结果进行词法识别
func (l *Lexer) parse() error {
	var tokenList []*Token
	var state int
	var symbol *Symbol
	for i := 0; i < len(l.source) || (i == len(l.source) && state != InitState); {
		b := byte(' ')
		if i != len(l.source) {
			b = l.source[i]
		}
		switch state {
		case InitState:
			symbol = &Symbol{
				Name: []byte{},
			}
			if isLetter(b) {
				state = IDState
				symbol.Name = append(symbol.Name, b)
				symbol.Class = Identifier
			} else if isDigit(b) {
				state = NumberState
				symbol.Name = append(symbol.Name, b)
				symbol.Class = IntConst
			} else if isSeparator(b) {
				state = SeparatorState
				symbol.Name = append(symbol.Name, b)
				symbol.Class = Separator

			} else if isOperator(b) {
				state = OperatorState
				symbol.Name = append(symbol.Name, b)
				symbol.Class = Operator
			} else if isSpace(b) || isNewline(b) {
				state = InitState
			} else {
				state = ErrState
			}
			i++
		case IDState:
			if isLetter(b) || isDigit(b) || isUnderline(b) {
				symbol.Name = append(symbol.Name, b)
				i++
			} else {
				if isKeyword(symbol.Name) {
					symbol.Class = Keyword
				}
				if isBoolConst(symbol.Name) {
					symbol.Class = BoolConst
				}
				if len(symbol.Name) > 8 {
					l.target = tokenList
					return fmt.Errorf("%w : %s", IdentifierTooLongErr, string(symbol.Name))
				}
				if v, ok := SymbolTable[string(symbol.Name)]; !ok {
					SymbolTable[string(symbol.Name)] = symbol
					tokenList = append(tokenList, &Token{
						Class: symbol.Class,
						Value: string(symbol.Name),
					})
				} else {
					tokenList = append(tokenList, &Token{
						Class: v.Class,
						Value: string(v.Name),
					})
				}
				state = InitState
			}
		case NumberState:
			if isDigit(b) {
				symbol.Name = append(symbol.Name, b)
				i++
			} else {
				if len(symbol.Name) > 1 && symbol.Name[0] == 48 {
					l.target = tokenList
					return fmt.Errorf("%w : %s", NumberStartWithZeroErr, string(symbol.Name))
				}
				if len(symbol.Name) > 8 {
					l.target = tokenList
					return fmt.Errorf("%w : %s", NumberTooLongErr, string(symbol.Name))
				}
				if v, ok := SymbolTable[string(symbol.Name)]; !ok {
					SymbolTable[string(symbol.Name)] = symbol
					tokenList = append(tokenList, &Token{
						Class: symbol.Class,
						Value: string(symbol.Name),
					})
				} else {
					tokenList = append(tokenList, &Token{
						Class: v.Class,
						Value: string(v.Name),
					})
				}
				state = InitState
			}
		case SeparatorState:
			if v, ok := SymbolTable[string(symbol.Name)]; !ok {
				SymbolTable[string(symbol.Name)] = symbol
				tokenList = append(tokenList, &Token{
					Class: symbol.Class,
					Value: string(symbol.Name),
				})
			} else {
				tokenList = append(tokenList, &Token{
					Class: v.Class,
					Value: string(v.Name),
				})
			}
			state = InitState
		case OperatorState:
			if isOperator(b) {
				symbol.Name = append(symbol.Name, b)
				i++
			} else {
				if !isValidOperator(symbol.Name) {
					l.target = tokenList
					return fmt.Errorf("%w : %s", InvalidOperatorErr, string(symbol.Name))
				}
				if v, ok := SymbolTable[string(symbol.Name)]; !ok {
					SymbolTable[string(symbol.Name)] = symbol
					tokenList = append(tokenList, &Token{
						Class: symbol.Class,
						Value: string(symbol.Name),
					})
				} else {
					tokenList = append(tokenList, &Token{
						Class: v.Class,
						Value: string(v.Name),
					})
				}
				state = InitState
			}
		case ErrState:
			l.target = tokenList
			return fmt.Errorf("%w : %v", InvalidCharacter, b)
		}
	}
	l.target = tokenList
	return nil
}
func isValidOperator(b []byte) bool {
	s := string(b)
	for _, op := range validOperators {
		if strings.EqualFold(s, op) {
			return true
		}
	}
	return false
}
func isBoolConst(b []byte) bool {
	s := string(b)
	return strings.EqualFold(s, "true") || strings.EqualFold(s, "false")
}
func isSpace(b byte) bool {
	return b == 32
}
func isNewline(b byte) bool {
	return b == 10
}
func isOperator(b byte) bool {
	for _, s := range operatorList {
		if b == s {
			return true
		}
	}
	return false
}
func isSeparator(b byte) bool {
	for _, v := range separatorList {
		if b == v {
			return true
		}
	}

	return false
}
func isUnderline(b byte) bool {
	return b == 95
}
func isKeyword(v []byte) bool {
	str := string(v)
	for _, s := range keywordList {
		if strings.EqualFold(str, s) {
			return true
		}
	}
	return false
}
func isLetter(b byte) bool {
	return (b >= 65 && b <= 90) || (b >= 97 && b <= 122)
}
func isDigit(b byte) bool {
	return b >= 48 && b <= 57
}
func IsBoolOperator(op string) bool {
	for _, s := range boolOperatorList {
		if strings.EqualFold(op, s) {
			return true
		}
	}
	return false
}
