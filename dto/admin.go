package dto

import "time"

type AdminInfoOutput struct {
	Id           int       `json:"id"`
	UserName     string    `json:"username"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}
