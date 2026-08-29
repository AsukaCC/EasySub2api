package service

import "time"

type AccountGroup struct {
	AccountID string
	GroupID   string
	Priority  int
	CreatedAt time.Time

	Account *Account
	Group   *Group
}
