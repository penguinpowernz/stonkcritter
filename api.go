package politstonk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/timshannon/badgerhold/v4"
)

func (bot *Bot) HandleListReps(c *gin.Context) {
	search := c.Query("q")
	s, err := bot.searchCritters(search)
	if err != nil {
		c.AbortWithError(500, err)
		return
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

	go bot.ConsumeDisclosures(dd)

	c.Status(204)
}

func (bot *Bot) HandlePullFromS3(c *gin.Context) {
	log.Println("downloading disclosures from S3 as per API")
	dd, err := GetDisclosuresFromS3()
	if err != nil {
		panic(err)
	}

	go bot.ConsumeDisclosures(dd)
	c.Status(204)
}

func (bot *Bot) HandleSetCursor(c *gin.Context) {
	cursor := c.Param("cursor")
	t, err := time.Parse("2006-01-02", cursor)
	if err != nil {
		log.Fatalf("failed to parse the cursor: %s: %s", cursor, err)
	}
	log.Printf("parsed cursor time as %s", t)
	d := NewDate(t)
	if err := bot.store.Upsert("cursor", &d); err != nil {
		log.Fatalf("failed to save the cursor: %s: %s", cursor, err)
	}
	log.Printf("updated cursor to %s (%s)", d.S, d.Time())

	c.JSON(200, d.S)
}
