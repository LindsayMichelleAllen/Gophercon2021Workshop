package main

import (
	"fmt"
	"time"
)

type Entry struct {
	Time    time.Time `json:"time"`
	User    string    `json:"user"`
	Content string    `json:"content"`
}

func (e Entry) Validate() error {
	if e.User == "" {
		return fmt.Errorf("empty user")
	}
	if e.Content == "" {
		return fmt.Errorf("empty content")
	}
	return nil
}
