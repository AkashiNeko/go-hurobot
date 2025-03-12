package qbot

import (
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	var user Users
	result := PsqlDB.First(&user, "user_id = ?", msg.Sender.UserID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = Users{
				UserID:   msg.Sender.UserID,
				Name:     msg.Sender.Nickname,
				Nickname: msg.Sender.Card,
			}
			if err := PsqlDB.Create(&user).Error; err != nil {
				return err
			}
		} else {
			return result.Error
		}
	}

	if user.Name != msg.Sender.Nickname {
		if err := PsqlDB.Model(&user).Update("name", msg.Sender.Nickname).Error; err != nil {
			return err
		}
	}

	newMessage := Messages{
		MsgID:   msg.MessageID,
		UserID:  msg.Sender.UserID,
		GroupID: msg.GroupID,
		Time:    time.Unix(int64(msg.Time), 0),
		Content: msg.RawMessage,
		Deleted: false,
	}
	if err := PsqlDB.Create(&newMessage).Error; err != nil {
		return err
	}
	return nil
}
