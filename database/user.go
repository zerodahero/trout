package database

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"gorm.io/gorm"
)

// User struct represents user model.
type User struct {
	ID          uint `gorm:"primarykey"`
	SlackID     string
	TeamID      string
	DisplayName string
	RealName    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func CreateUserFromSlackUser(slackUser *slack.User, db *gorm.DB) *User {
	var user User

	user.SlackID = slackUser.ID
	user.TeamID = slackUser.TeamID

	// Counting on one of these always being present
	if slackUser.Profile.RealNameNormalized != "" {
		user.RealName = slackUser.Profile.RealNameNormalized
	} else {
		user.RealName = slackUser.Profile.RealName
	}

	// Fallback to real name if display name is empty
	if slackUser.Profile.DisplayNameNormalized != "" {
		user.DisplayName = slackUser.Profile.DisplayNameNormalized
	} else if slackUser.Profile.DisplayName != "" {
		user.DisplayName = slackUser.Profile.DisplayName
	} else {
		user.DisplayName = user.RealName
	}

	db.Create(&user)

	return &user
}

func GetOrFetchUser(userID string, fetch func(string) (*slack.User, error)) (*User, error) {
	user, err := GetUser(userID)
	if err != nil {
		return nil, err
	}

	// If the user exists in DB already, return
	if user != nil {
		return user, nil
	}

	// Fetch user info for DB
	slackUser, err := fetch(userID)
	if err != nil || slackUser == nil {
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	user = CreateUserFromSlackUser(slackUser, db)

	return user, nil
}

func GetUser(userID string) (*User, error) {
	var user User
	result := db.Where("slack_id = ?", userID).First(&user)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error querying for db user: %v", result.Error)
	}

	// No user found
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &user, nil
}
