// Package main  @Author xiaobaiio 2023/3/11 16:54:00
package main

import "github.com/gin-gonic/gin"

func main() {
	g := gin.New()
	g.POST("/api/calculate", Calculate)
	g.Run(":8081")
}
