// Package main  @Author xiaobaiio 2023/3/11 14:36:00
package main

import (
	"fmt"
	"log"
	"strconv"
)

type Stack[T byte | int] struct {
	index int
	data  []T
}

func (stack *Stack[T]) push(op T) {
	stack.index++
	if len(stack.data) < stack.index+1 {
		stack.data = append(stack.data, op)
		return
	}
	stack.data[stack.index] = op
}

func (stack *Stack[T]) pop() T {
	if stack.index >= 0 {
		v := stack.data[stack.index]
		stack.index--
		return v
	}
	return 0
}
func (stack *Stack[T]) poll() T {
	if stack.index >= 0 {
		v := stack.data[stack.index]
		return v
	}
	return 0
}

var stack *Stack[byte]
var priorityMap map[byte]int

func init() {
	stack = &Stack[byte]{
		index: -1,
		data:  []byte{},
	}
	priorityMap = make(map[byte]int)
	priorityMap['+'] = 1
	priorityMap['-'] = 1
	priorityMap['*'] = 2
	priorityMap['/'] = 2
	priorityMap['('] = 0
	priorityMap[')'] = 3
}
func main() {
	scanExpression()
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
	stack.index = -1
	var buf []byte
	i := 0
	buf = []byte(*expression)
	var res []byte
	for i != len(buf) {
		c := buf[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			res = append(res, c)
		case '+', '-', '*', '/', ')':
			res = opService(c, res)
		case '(':
			stack.push('(')
		}

	}
	for stack.index != -1 {
		res = append(res, stack.pop())
	}

	fmt.Println(string(res))
	number := cal(res)
	fmt.Println(number)
}
func cal(res []byte) int {
	numStack := &Stack[int]{
		index: -1,
		data:  []int{},
	}
	for i := 0; i < len(res); i++ {
		c := res[i]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			n, _ := strconv.Atoi(string(c))
			numStack.push(n)
		case '+', '-', '*', '/', ')':
			n2 := numStack.pop()
			n1 := numStack.pop()
			r := calNumber(n1, n2, c)
			numStack.push(r)
		}
	}
	return numStack.pop()
}
func calNumber(n1, n2 int, op byte) int {
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
func opService(op byte, res []byte) []byte {
	if op == ')' {
		c := stack.pop()
		for c != '(' {
			res = append(res, c)
			c = stack.pop()
		}
	} else {
		ok := comparePriority(op)
		if ok {
			stack.push(op)
		} else {
			c := stack.pop()
			res = append(res, c)
			stack.push(op)
		}
	}
	return res
}
func comparePriority(op byte) bool {
	if stack.index == -1 {
		return true
	} else {
		pop := stack.poll()
		if priorityMap[op] <= priorityMap[pop] {
			return false
		} else {
			return true
		}
	}
}
