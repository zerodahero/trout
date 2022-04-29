package database

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/zerodahero/trout/parser"

	"github.com/slack-go/slack/slackevents"
	"gopkg.in/guregu/null.v4"
)

// Kudo struct represents shout_out model.
type Kudo struct {
	ID          uint `gorm:"primarykey"`
	FromUserID  string
	ToUserID    string
	Message     string
	IsPublic    bool
	IsAnonymous bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SharedAt    null.Time
}

func NewKudo(from, to, message string) *Kudo {
	return &Kudo{
		FromUserID:  from,
		ToUserID:    to,
		Message:     message,
		IsPublic:    true,
		IsAnonymous: false,
	}
}

func NewKudoFromMentionEvent(ev *slackevents.AppMentionEvent, userID string) (*Kudo, error) {
	msg := strings.ReplaceAll(ev.Text, fmt.Sprintf("<@%s>", userID), "")

	return NewKudoFromText(msg, ev.User)
}

func NewKudoFromText(text, fromUser string) (*Kudo, error) {
	text = parser.RemoveWhitespace(text)

	to, err := parser.ParseRecipientFromText(text)
	if err != nil {
		return nil, err
	}

	return &Kudo{
		FromUserID:  fromUser,
		ToUserID:    to,
		Message:     text,
		IsPublic:    true,
		IsAnonymous: false,
	}, nil
}

func GetKudoByID(kudoID int) (*Kudo, error) {
	var kudo Kudo

	result := db.First(&kudo, kudoID)
	if result.Error != nil {
		return nil, fmt.Errorf("could not find shout out: %v", result.Error)
	}

	return &kudo, nil
}

func (k *Kudo) Save() error {
	result := db.Save(k)
	return result.Error
}

func GetUnsharedKudos(public bool) ([]*Kudo, error) {
	var kudos []*Kudo
	result := db.Where("shared_at IS NULL").
		Where("is_public = ?", public).
		Order("to_user_id ASC").
		Find(&kudos)

	if result.Error != nil {
		return nil, result.Error
	}

	return kudos, nil
}

func (k *Kudo) MarkShared(t time.Time) error {
	k.SharedAt = null.TimeFrom(t)
	return k.Save()
}

var anonymousNames = []string{
	"Sue Doe Nimm",
	"A. Nonny Muz",
	"Naym Less",
	"Mr. E",
	"Sohm Bahdy",
	"See Krett",
	"Hayden P. Son",
	"Ehn Kagn Ito",
	"Cass E. Fied",
	"E. Nigma",
	"D. Sgeyzed",
	"Miss Teekal",
	"Coe Bert",
	"Carrie Terr",
	"Annie Juan",
	"Uda Kuvah",
	"A. Liam",
	"Cohen Seeld",
	"A. Dewd",
	"Guy",
	"Hugh Mann",
	"Indie Vitual",
	"Creed Chore",
	"Moe Sapian",
	"NPC",
	"Roe L.",
	"Kal Eague",
	"Coe Warker",
}

func (k *Kudo) GetDisplayFrom(mention bool) string {
	if k.IsAnonymous {
		return anonymousNames[rand.Intn(len(anonymousNames))]
	}

	if mention {
		return parser.WrapUserIdForMention(k.FromUserID)
	}

	// We expect the user to exist (no fetch from slack required)
	fromUser, err := GetUser(k.FromUserID)
	if err != nil {
		return "?"
	}

	return fromUser.DisplayName
}
