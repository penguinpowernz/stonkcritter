package politstonk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/timshannon/badgerhold/v4"
)

func (bot *Bot) HandleListReps(c *gin.Context) {
	reps := []Rep{}
	err := bot.store.Find(
		&reps,
		new(badgerhold.Query))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	s := []string{}
	for _, r := range reps {
		s = append(s, r.Name)
	}
	c.JSON(200, s)
}

func (bot *Bot) HandleDisclosures(c *gin.Context) {
	data, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	fmt.Println(len(data), err)

	var dd []Disclosure

	err = json.Unmarshal(data, &dd)
	fmt.Println(len(dd), err)

	// allow setting a custom cursor to send from (allow sending yesterdays)
	date := bot.GetCursor()
	if cursor := c.Query("cursor"); cursor != "" {
		t, err := time.Parse("2006-01-02", cursor)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		date = NewDate(t)
	}

	dd = Disclosures(dd).After(date)

	bot.StoreReps(dd)
	for _, d := range dd {
		bot.Broadcast(d.String())
		bot.DispatchDisclosure(d)
	}

	c.Status(204)
}
