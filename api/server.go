package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/stonkcritter/models"
)

type Capabilities interface {
	// brain
	ListCritters() ([]string, error)

	// bot
	Subs() []models.Sub

	// watcher
	CheckNow()
	CurrentCursor() time.Time
	Checks() int
	Dispatched() int
	Inflight() int
}

type Server struct {
	capabilities Capabilities
}

func NewServer(capabilities Capabilities) *Server {
	return &Server{capabilities}
}

func (svr *Server) SetupRoutes(r gin.IRouter) {
	r.GET("/", svr.getInfo)
	r.GET("/critters", svr.listCritters)
	r.GET("/subs", svr.listSubs)
	r.PUT("/watcher/check", svr.checkNow)
}

func (svr *Server) getInfo(c *gin.Context) {
	c.JSON(200, map[string]interface{}{
		"checks":     svr.capabilities.Checks(),
		"dispatched": svr.capabilities.Dispatched(),
		"inflight":   svr.capabilities.Inflight(),
		"cursor":     svr.capabilities.CurrentCursor().Format("2006-01-02"),
	})
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
