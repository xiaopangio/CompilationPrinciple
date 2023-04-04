// Package v2  @Author xiaobaiio 2023/3/11 15:42:00
package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

var (
	errorOperateNumberLack    = errors.New("cannot pop number from number stack")
	errorRedundantNumber      = errors.New("redundant digital exist in the stack")
	errorMatchLeftParenthesis = errors.New("cannot match (")
)

var opStack *Stack[byte]
var numStack *Stack[int]
var priorityMap map[byte]int

type Stack[T byte | int] struct {
	index int
	data  []T
}

func (stack *Stack[T]) push(v T) {
	stack.index++
	if len(stack.data) < stack.index+1 {
		stack.data = append(stack.data, v)
		return
	}
	stack.data[stack.index] = v
}

func (stack *Stack[T]) pop() (T, bool) {
	if stack.index >= 0 {
		v := stack.data[stack.index]
		stack.index--
		return v, true
	}
	return 0, false
}
func (stack *Stack[T]) poll() T {
	if stack.index >= 0 {
		v := stack.data[stack.index]
		return v
	}
	return 0
}

func init() {
	opStack = &Stack[byte]{
		index: -1,
		data:  []byte{},
	}
	numStack = &Stack[int]{
		index: -1,
		data:  []int{},
	}
	priorityMap = make(map[byte]int)
	priorityMap['+'] = 1
	priorityMap['-'] = 1
	priorityMap['*'] = 2
	priorityMap['/'] = 2
	priorityMap['('] = 0
	priorityMap[')'] = 3
}
func scanExpression() {
	var expression string
	for {
		fmt.Println("Please enter the expression：")
		_, err := fmt.Scanln(&expression)
		if err != nil {
			log.Println(err)
			return
		}
		service(&expression)
	}
}
func service(expression *string) {
	opStack.index = -1
	numStack.index = -1
	var buf []byte
	i := 0
	buf = []byte(*expression)
	var res int
	for i != len(buf) {
		c := buf[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			num, _ := strconv.Atoi(string(c))
			numStack.push(num)
		case '+', '-', '*', '/', ')':
			err := opService(c)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		case '(':
			opStack.push('(')
		}

	}
	for opStack.index != -1 {
		c, ok := opStack.pop()
		n2, ok := numStack.pop()
		if !ok {
			fmt.Println(errorOperateNumberLack.Error())
			return
		}
		n1, ok := numStack.pop()
		if !ok {
			fmt.Println(errorOperateNumberLack.Error())
			return
		}
		res = cal(n1, n2, c)
		numStack.push(res)
	}
	if numStack.index != 0 {
		fmt.Println(errorRedundantNumber.Error())
		return
	}
	res = numStack.data[0]
	fmt.Println(res)
}
func opService(op byte) error {
	res := 0
	if op == ')' {
		c, ok := opStack.pop()
		if !ok {
			return errorMatchLeftParenthesis
		}
		for c != '(' {
			n2, ok := numStack.pop()
			if !ok {
				return errorOperateNumberLack
			}
			n1, ok := numStack.pop()
			if !ok {
				return errorOperateNumberLack
			}
			res = cal(n1, n2, c)
			numStack.push(res)
			c, ok = opStack.pop()
			if !ok {
				return errorMatchLeftParenthesis
			}
		}
	} else {
		ok := comparePriority(op)
		if ok {
			opStack.push(op)
		} else {
			c, ok := opStack.pop()
			if !ok {
				return errorMatchLeftParenthesis
			}
			n2, ok := numStack.pop()
			if !ok {
				return errorOperateNumberLack
			}
			n1, ok := numStack.pop()
			if !ok {
				return errorOperateNumberLack
			}
			res = cal(n1, n2, c)
			numStack.push(res)
			opStack.push(op)
		}
	}
	return nil
}
func cal(n1, n2 int, op byte) int {
	res := 0
	switch op {
	case '+':
		res = n1 + n2
	case '-':
		res = n1 - n2
	case '*':
		res = n1 * n2
	case '/':
		res = n1 / n2
	}
	return res
}
func comparePriority(op byte) bool {
	if opStack.index == -1 {
		return true
	} else {
		pop := opStack.poll()
		if priorityMap[op] <= priorityMap[pop] {
			return false
		} else {
			return true
		}
	}
}
func main() {
	scanExpression()
}
