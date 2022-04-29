package parser

import (
	"fmt"
	"testing"
)

func TestGetMentionCount(t *testing.T) {
	var tests = []struct {
		text string
		want int
	}{
		{"<@ABCDE12345> Something.", 1},
		{"<@ABCDE12345> Something <@EDCFA19238>", 2},
		{"<@ABCDE12345> Something. <@SDFJ09812> <@012984ASFD>", 3},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.text)
		t.Run(testname, func(t *testing.T) {
			ans := GetMentionCount(tt.text)
			if ans != tt.want {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
	}
}

func TestParseRecipientFromText(t *testing.T) {
	var tests = []struct {
		text string
		want string
	}{
		{"<@ABCDE12345> Something.", "ABCDE12345"},
		{"Something about <@ABCDE12345>.", "ABCDE12345"},
		{"<@ABCDE12345|jim_bob> Something.", "ABCDE12345"},
		{"Something about <@ABCDE12345|jim_bob>.", "ABCDE12345"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.text)
		t.Run(testname, func(t *testing.T) {
			ans, _ := ParseRecipientFromText(tt.text)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestReplacesUserInText(t *testing.T) {
	var tests = []struct {
		text   string
		userID string
		name   string
		want   string
	}{
		{"<@ABCDE12345> Something.", "ABCDE12345", "Jim Bob", "Jim Bob Something."},
		{"Something about <@ABCDE12345>.", "ABCDE12345", "Suzy", "Something about Suzy."},
		{"<@ABCDE12345|jim_bob> Something.", "ABCDE12345", "jimbo", "jimbo Something."},
		{"Something about <@ABCDE12345|jim_bob>.", "ABCDE12345", "jimmy beans the dude", "Something about jimmy beans the dude."},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.text)
		t.Run(testname, func(t *testing.T) {
			ans := ReplaceUserInText(tt.text, tt.userID, tt.name)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestWrapsUserIDForMention(t *testing.T) {
	var tests = []struct {
		userID string
		want   string
	}{
		{"ABCDE12345", "<@ABCDE12345>"},
		{"123UAKDU", "<@123UAKDU>"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.userID)
		t.Run(testname, func(t *testing.T) {
			ans := WrapUserIdForMention(tt.userID)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}
