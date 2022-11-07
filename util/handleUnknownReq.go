package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)
func HandleNoRoute(c *gin.Context){
	c.JSON(http.StatusNotFound,gin.H{
		"Status":"RouteNotFound",
		"Info":"NotFound",
	})
}
func HandleNoMethod(c *gin.Context){
	c.JSON(http.StatusMethodNotAllowed,gin.H{
		"Status":"MethodNotAllowed",
		"Info":"NotAllowed",
	})
}