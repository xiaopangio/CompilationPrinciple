// Package test  @Author xiaobaiio 2023/3/26 22:01:00
package main

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

func removeNthFromEnd(head *ListNode, n int) *ListNode {
	n1, n2 := head, head
	flag := false
	count := 1
	for n2.Next != nil {
		n1 = n1.Next
		count++
		if n2.Next.Next == nil {
			flag = true
			break
		}
		n2 = n2.Next.Next
	}
	x := 0
	if flag {
		x = count - n - 1
	} else {
		x = count - n
	}
	pre := n1
	if x <= 0 {
		x = x + count - 1
		pre = head
		n1 = head
		for i := 0; i < x; i++ {
			pre = n1
			n1 = n1.Next
		}
		if pre == head && n1 == head {
			return pre.Next
		}
	} else {
		for i := 0; i < x; i++ {
			pre = n1
			n1 = n1.Next
		}
	}
	pre.Next = n1.Next
	return head
}
func main() {
	a := &ListNode{Val: 1}
	b := &ListNode{Val: 2}
	c := &ListNode{Val: 3}
	d := &ListNode{Val: 4}
	e := &ListNode{Val: 5}
	a.Next = b
	b.Next = c
	c.Next = d
	d.Next = e
	head := removeNthFromEnd(a, 5)
	n := head
	for n != nil {
		fmt.Printf("%d ", n.Val)
		n = n.Next
	}
}
