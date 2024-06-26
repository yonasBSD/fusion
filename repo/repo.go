package repo

import (
	"errors"

	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	conn, err := gorm.Open(
		sqlite.Open(conf.Conf.DB),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		panic(err)
	}
	DB = conn

	migrage()
	registerCallback()
}

func migrage() {
	// FIX: gorm not auto drop index and change 'not null'
	if err := DB.AutoMigrate(&model.Feed{}, &model.Group{}, &model.Item{}); err != nil {
		panic(err)
	}

	defaultGroup := "Default"
	if err := DB.Model(&model.Group{}).Where("id = ?", 1).
		FirstOrCreate(&model.Group{ID: 1, Name: &defaultGroup}).Error; err != nil {
		panic(err)
	}
}

func registerCallback() {
	if err := DB.Callback().Query().After("*").Register("convert_error", func(db *gorm.DB) {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			db.Error = ErrNotFound
		}
	}); err != nil {
		panic(err)
	}

	if err := DB.Callback().Create().After("*").Register("convert_error", func(db *gorm.DB) {
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		panic(err)
	}

	if err := DB.Callback().Update().After("*").Register("convert_error", func(db *gorm.DB) {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = ErrNotFound
		}
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		panic(err)
	}

	if err := DB.Callback().Delete().After("*").Register("convert_error", func(db *gorm.DB) {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = ErrNotFound
		}
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		panic(err)
	}
}
