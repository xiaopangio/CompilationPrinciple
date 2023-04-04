// Package main  @Author xiaobaiio 2023/3/18 14:44:00
package main

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
	identifier = iota + 1 //标识符
	intConst              //整常数
	boolConst             //布尔常数
	keyword               //关键字
	operator              //运算符
	separator             //分隔符
)

var classMap map[int]string

const (
	space      = byte(32)
	enter      = byte(13)
	newLine    = byte(10)
	slash      = byte(47)
	star       = byte(42)
	singleLine = 1
	multiLine  = 2
)

var (
	keywordList   []string
	operatorList  []byte
	separatorList []byte
)
var (
	validOperators []string
)

type Lexer struct {
	source []byte
	target []*Token
	err    error
}
type Token struct {
	Class int
	Value []byte
}

func InitLexer() error {
	keywords, err := os.ReadFile("init/keyword.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "keyword.txt")
		return err
	}
	keywordList = strings.Split(string(keywords), ",")
	operatorList, err = os.ReadFile("init/operator.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "operator.txt")
		return err
	}
	separatorList, err = os.ReadFile("init/separator.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "separator.txt")
		return err
	}
	data, err := os.ReadFile("init/validOperator.txt")
	if err != nil {
		log.Fatal("cannot read source code from file: ", "validOperator.txt")
		return err
	}
	validOperators = strings.Split(string(data), ",")
	classMap = make(map[int]string)
	classMap[identifier] = "identifier"
	classMap[intConst] = "intConst"
	classMap[boolConst] = "boolConst"
	classMap[keyword] = "keyword"
	classMap[operator] = "operator"
	classMap[separator] = "separator"
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
			if pre == enter {
				cleanSpace = true
			} else {
				return errors.New("after enter is not a newline")
			}
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

const (
	InitState = iota
	IDState
	NumberState
	SeparatorState
	OperatorState
	ErrState
)

var (
	IdentifierTooLongErr   = errors.New("identifier's length greater than 8")
	NumberTooLongErr       = errors.New("number's length greater than 8")
	NumberStartWithZeroErr = errors.New("number start with zero")
	InvalidOperatorErr     = errors.New("invalid operator")
	InvalidCharacter       = errors.New("invalid character")
)

// pares 对clean后的结果进行词法识别
func (l *Lexer) parse() error {
	var tokenList []*Token
	var state int
	var token *Token
	//fmt.Printf("%v\n", l.source)

	for i := 0; i < len(l.source); {
		b := l.source[i]
		switch state {
		case InitState:
			token = &Token{
				Value: []byte{},
			}
			if isLetter(b) {
				state = IDState
				token.Value = append(token.Value, b)
				token.Class = identifier
			} else if isDigit(b) {
				state = NumberState
				token.Value = append(token.Value, b)
				token.Class = intConst
			} else if isSeparator(b) {
				state = SeparatorState
				token.Value = append(token.Value, b)
				token.Class = separator
			} else if isOperator(b) {
				state = OperatorState
				token.Value = append(token.Value, b)
				token.Class = operator
			} else if isSpace(b) || isNewline(b) {
				state = InitState
			} else {
				state = ErrState
			}
			i++
		case IDState:
			if isLetter(b) || isDigit(b) || isUnderline(b) {
				token.Value = append(token.Value, b)
				i++
			} else {
				if isKeyword(token.Value) {
					token.Class = keyword
				}
				if isBoolConst(token.Value) {
					token.Class = boolConst
				}
				if len(token.Value) > 8 {
					l.target = tokenList
					return fmt.Errorf("%w : %s", IdentifierTooLongErr, string(token.Value))
				}
				tokenList = append(tokenList, token)
				state = InitState
			}
		case NumberState:
			if isDigit(b) {
				token.Value = append(token.Value, b)
				i++
			} else {
				if len(token.Value) > 1 && token.Value[0] == 48 {
					l.target = tokenList
					return fmt.Errorf("%w : %s", NumberStartWithZeroErr, string(token.Value))
				}
				if len(token.Value) > 8 {
					l.target = tokenList
					return fmt.Errorf("%w : %s", NumberTooLongErr, string(token.Value))
				}
				tokenList = append(tokenList, token)
				state = InitState
			}
		case SeparatorState:
			tokenList = append(tokenList, token)
			state = InitState
		case OperatorState:
			if isOperator(b) {
				token.Value = append(token.Value, b)
				i++
			} else {
				if !isValidOperator(token.Value) {
					l.target = tokenList
					return fmt.Errorf("%w : %s", InvalidOperatorErr, string(token.Value))
				}
				tokenList = append(tokenList, token)
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
func toLower(b byte) byte {
	return b | 1<<6
}
func isDigit(b byte) bool {
	return b >= 48 && b <= 57
}
func isNotZeroDigit(b byte) bool {
	return b > 48 && b <= 57
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
		fmt.Printf("class: %s, value: %s\n", classMap[token.Class], token.Value)
	}
}
