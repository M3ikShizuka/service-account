package domain

import "time"

type User struct {
	Id               uint32
	Username         string
	Email            string
	PasswordHash     []byte
	DateRegistration time.Time
	DateLastOnline   time.Time
}
