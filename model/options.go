package model

import "gorm.io/gorm"

type TreeOptions struct {
	DB      *gorm.DB
	Table   string
	AppId   string
	Id      string
	Preload []string
	Lazy    bool
}
