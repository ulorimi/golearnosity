package learnosity

import (
	"fmt"
	"testing"
	"time"
)

var consumerKey = "yis0TYCu7U9V4o7M"
var consumerSecret = "74c5fd430cf1242a527f6223aebd42d30464be22"
var expectedSignature = "e9cd04b624d1dbe89fd4cad0a447f485e0fcec1392cbd3e2841826a954cc4e8e"

func TestGenerateSignaureBasic(t *testing.T) {
	timestamp, err := time.Parse("20060102-1504", "20140612-0438")
	loc, _ := time.LoadLocation("GMT")
	timestamp = timestamp.In(loc)
	tim := time.Now().In(loc)
	fmt.Println(tim)
	fmt.Println(timestamp)
	if err != nil {
		panic(err)
	}
	security := M{
		"consumer_key": consumerKey,
		"user_id":      "12345678",
		"timestamp":    "20140612-0438",
	}

	req := &map[string]interface{}{}

	packet, err := Init("questions", security, consumerSecret, req, "")
	if err != nil {
		t.FailNow()
	}
	fmt.Println(packet)
	signature := packet["signature"]
	fmt.Println(signature)
	if signature != expectedSignature {
		t.Fail()
	}
}
