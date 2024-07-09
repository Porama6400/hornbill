package apiserver

import "github.com/gin-gonic/gin"

func ReadinessProbe(g *gin.Context) {
	g.JSON(200, gin.H{})
}
