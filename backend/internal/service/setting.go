package service

import "time"

type Setting struct {
	ID        string
	Key       string
	Value     string
	UpdatedAt time.Time
}
