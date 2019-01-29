package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkCreateRegexReplacer(t *testing.T) {
	var tests = []struct {
		LinkCreate      *LinkCreate
		inputMessage    string
		expectedMessage string
	}{
		{
			&LinkCreate{
				TextPattern: "Mattermost",
				UrlTemplate: "https://mattermost.com",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		},
		{
			&LinkCreate{
				TextPattern: "(?P<key>Mattermost)",
				UrlTemplate: "https://mattermost.com",
			},
			"Welcome to Mattermost and have fun with Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com) and have fun with [Mattermost](https://mattermost.com)!",
		},
		{
			&LinkCreate{
				TextPattern: "MM-(?P<jira_id>\\d+)",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&LinkCreate{
				TextPattern: "MM-(?P<jira_id>\\d+)",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"Link in brackets should link (see MM-12345)",
			"Link in brackets should link (see [MM-12345](https://mattermost.atlassian.net/browse/MM-12345))",
		},
		{
			&LinkCreate{
				TextPattern: "MM-(?P<jira_id>\\d+)",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"Link a ticket MM-12345, before a comma",
			"Link a ticket [MM-12345](https://mattermost.atlassian.net/browse/MM-12345), before a comma",
		},
		{
			&LinkCreate{
				TextPattern: "MM-(?P<jira_id>\\d+)",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&LinkCreate{
				TextPattern: "\\bMM-(?P<jira_id>\\d+)\\b",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"WelcomeMM-12345should not link!",
			"WelcomeMM-12345should not link!",
		},
		{
			&LinkCreate{
				TextPattern: "\\bMM-(?P<jira_id>\\d+)\\b",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"MM-12345, (MM-12345), notMM-12345though, and:MM-12345",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345), " +
				"([MM-12345](https://mattermost.atlassian.net/browse/MM-12345)), " +
				"notMM-12345though, and:[MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
	}

	for _, tt := range tests {
		al, _ := NewLinkCreateRegexReplacer(tt.LinkCreate)
		actual := al.Replace(tt.inputMessage)

		assert.Equal(t, tt.expectedMessage, actual)
	}
}

func TestLinkTitleRegexReplacer(t *testing.T) {
	var tests = []struct {
		LinkTitle      *LinkTitle
		inputMessage    string
		expectedMessage string
	}{
		{
			&LinkTitle{
				UrlPattern: "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
				TitleTemplate: "MM-$jira_id",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&LinkTitle{
				UrlPattern: "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
				TitleTemplate: "MM-$jira_id",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link!",
		},
		{
			&LinkTitle{
				UrlPattern: "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
				TitleTemplate: "MM-$jira_id",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link https://mattermost.atlassian.net/browse/MM-12346 !",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link [MM-12346](https://mattermost.atlassian.net/browse/MM-12346) !",
		},
		{
			&LinkTitle{
				UrlPattern: "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
				TitleTemplate: "MM-$jira_id",
			},
			"https://mattermost.atlassian.net/browse/MM-12345 https://mattermost.atlassian.net/browse/MM-12345",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
		{
			&LinkTitle{
				UrlPattern: "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
				TitleTemplate: "MM-$jira_id",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
		{
			&LinkTitle{
				UrlPattern: "https://mattermost(.atlassian.net)?/browse/MM-(?P<jira_id>\\d+)",
				TitleTemplate: "MM-$jira_id",
				UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
			},
			"Both https://mattermost.atlassian.net/browse/MM-12345 and https://mattermost/browse/MM-12345 work",
			"Both [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) and [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) work",
		},
	}

	for _, tt := range tests {
		al, _ := NewLinkTitleRegexReplacer(tt.LinkTitle)
		actual := al.Replace(tt.inputMessage)

		assert.Equal(t, tt.expectedMessage, actual)
	}
}

func TestLinkCreateErrors(t *testing.T) {
	var tests = []struct {
		Link *LinkCreate
	}{
		{}, {
			&LinkCreate{},
		},
		{
			&LinkCreate{
				TextPattern: "",
				UrlTemplate: "blah",
			},
		},
		{
			&LinkCreate{
				TextPattern: "blah",
				UrlTemplate: "",
			},
		},
	}

	for _, tt := range tests {
		_, err := NewLinkCreateRegexReplacer(tt.Link)
		assert.NotNil(t, err)
	}
}

func TestLinkTitleErrors(t *testing.T) {
	var tests = []struct {
		Link *LinkTitle
	}{
		{}, {
			&LinkTitle{},
		},
		{
			&LinkTitle{
				UrlPattern: "",
				TitleTemplate: "blah",
			},
		},
		{
			&LinkTitle{
				UrlPattern: "blah",
				TitleTemplate: "",
			},
		},
	}

	for _, tt := range tests {
		_, err := NewLinkTitleRegexReplacer(tt.Link)
		assert.NotNil(t, err)
	}
}
