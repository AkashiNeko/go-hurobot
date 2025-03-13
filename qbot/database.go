package qbot

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var PsqlDB *gorm.DB = nil
var PsqlConnected bool = false

type Users struct {
	UserID   uint64 `gorm:"primaryKey;column:user_id"`
	Name     string `gorm:"not null;column:name"`
	Nickname string `gorm:"column:nick_name"`
	Gender   bool   `gorm:"column:gender"`
}

type Messages struct {
	MsgID   uint64    `gorm:"primaryKey;column:msg_id"`
	UserID  uint64    `gorm:"not null;column:user_id"`
	GroupID uint64    `gorm:"not null;column:group_id"`
	Content string    `gorm:"not null;column:content"`
	Time    time.Time `gorm:"not null;column:time"`
	Deleted bool      `gorm:"column:deleted"`
}

func initPsqlDB(dsn string) error {
	var err error
	if PsqlDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		return err
	}
	PsqlConnected = true
	return PsqlDB.AutoMigrate(&Users{}, &Messages{})
}

func saveDatabase(msg *Message) error {
	return PsqlDB.Transaction(func(tx *gorm.DB) error {
		user := Users{
			UserID:   msg.UserID,
			Name:     msg.Nickname,
			Nickname: msg.Card,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"name": gorm.Expr("EXCLUDED.name"),
				},
			),
		}).Where("users.name <> EXCLUDED.name").Create(&user).Error; err != nil {
			return err
		}

		newMessage := Messages{
			MsgID:   msg.MsgID,
			UserID:  msg.UserID,
			GroupID: msg.GroupID,
			Time:    time.Unix(int64(msg.Time), 0),
			Content: msg.Raw,
			Deleted: false,
		}
		if err := tx.Create(&newMessage).Error; err != nil {
			return err
		}
		return nil
	})
}
