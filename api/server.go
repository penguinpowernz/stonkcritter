package api

import (
	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/stonkcritter/models"
)

type Capabilities interface {
	CheckNow()
	ListCritters() ([]string, error)
	Subs() []models.Sub
}

type Server struct {
	capabilities Capabilities
}

func NewServer(capabilities Capabilities) *Server {
	return &Server{capabilities}
}

func (svr *Server) SetupRoutes(r gin.IRouter) {
	r.GET("/critters", svr.listCritters)
	r.GET("/subs", svr.listSubs)
	r.PUT("/watcher/check", svr.checkNow)
}

func (svr *Server) listSubs(c *gin.Context) {
	ss := svr.capabilities.Subs()
	c.JSON(200, ss)
}
func (svr *Server) listCritters(c *gin.Context) {
	ss, err := svr.capabilities.ListCritters()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, ss)
}

func (svr *Server) checkNow(c *gin.Context) {
	go svr.capabilities.CheckNow()
	c.Status(202)
}
