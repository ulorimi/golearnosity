package learnosity

import (
	"time"
)

//SecurityPacket contains learnosity security info
type SecurityPacket struct {
	ConsumerKey string     `json:"consumer_key"`
	Domain      string     `json:"domain"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	UserID      string     `json:"user_id"`
	Signature   string     `json:"signature"`
}
