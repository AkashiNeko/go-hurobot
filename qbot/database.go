package qbot

import (
	"errors"
	"fmt"
	"go-hurobot/config"
	"log"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var PsqlConnected bool

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

func init() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PsqlHost, strconv.Itoa(int(config.PsqlPort)), config.PsqlUser, config.PsqlPassword, config.PsqlDbName)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("gorm.Open: %v", err)
	}
	err = db.AutoMigrate(&Users{}, &Messages{})
	if err != nil {
		log.Fatalf("db.AutoMigrate: %v", err)
	}
}

func saveDatabase(msg *Message) error {
	var user Users
	result := db.First(&user, "user_id = ?", msg.Sender.UserID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = Users{
				UserID:   msg.Sender.UserID,
				Name:     msg.Sender.Nickname,
				Nickname: msg.Sender.Card,
			}
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		} else {
			return result.Error
		}
	}

	if user.Name != msg.Sender.Nickname {
		if err := db.Model(&user).Update("name", msg.Sender.Nickname).Error; err != nil {
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
	if err := db.Create(&newMessage).Error; err != nil {
		return err
	}
	return nil
}
