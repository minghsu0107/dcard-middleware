package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func helloworld(c *gin.Context) {
	c.String(http.StatusOK, "Hello, world!")
}
