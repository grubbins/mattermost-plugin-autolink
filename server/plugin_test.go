package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func TestPlugin(t *testing.T) {
	link_titles := make([]*LinkTitle, 0)
	link_titles = append(link_titles, &LinkTitle{
		UrlPattern:  "https://mattermost.com",
		TitleTemplate:  "Mattermost",
	})
	link_creates := make([]*LinkCreate, 0)
	link_creates = append(link_creates, &LinkCreate{
		TextPattern:  "Mattermost",
		UrlTemplate:  "https://mattermost.com",
	})
	validConfiguration := Configuration{link_titles, link_creates}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to Mattermost! You may enjoy https://mattermost.com."}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)! You may enjoy [Mattermost](https://mattermost.com).", rpost.Message)
}

func TestSpecialCases(t *testing.T) {
	link_creates := make([]*LinkCreate, 0)
	link_titles := make([]*LinkTitle, 0)
	link_creates = append(link_creates, &LinkCreate{
		TextPattern:  "Mattermost",
		UrlTemplate: "https://mattermost.com",
	}, &LinkCreate{
		TextPattern:  "MM-(?P<jira_id>\\d+)",
		UrlTemplate: "https://mattermost.atlassian.net/browse/MM-$jira_id",
	})
	link_titles = append(link_titles, &LinkTitle{
		UrlPattern:  "https://mattermost.com",
		TitleTemplate: "the Mattermost portal",
	}, &LinkTitle{
		UrlPattern:  "https://www.mattermost.com",
		TitleTemplate: "the Mattermost portal",
		UrlTemplate: "https://mattermost.com",
	}, &LinkTitle{
		UrlPattern:  "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
		TitleTemplate: "MM-$jira_id",
	})
	validConfiguration := Configuration{link_titles, link_creates}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()

	var tests = []struct {
		inputMessage    string
		expectedMessage string
	}{
		{
			"hello ``` Mattermost ``` goodbye",
			"hello ``` Mattermost ``` goodbye",
		}, {
			"hello\n```\nMattermost\n```\ngoodbye",
			"hello\n```\nMattermost\n```\ngoodbye",
		}, {
			"Mattermost ``` Mattermost ``` goodbye",
			"[Mattermost](https://mattermost.com) ``` Mattermost ``` goodbye",
		}, {
			"``` Mattermost ``` Mattermost",
			"``` Mattermost ``` [Mattermost](https://mattermost.com)",
		}, {
			"Mattermost ``` Mattermost ```",
			"[Mattermost](https://mattermost.com) ``` Mattermost ```",
		}, {
			"Mattermost ``` Mattermost ```\n\n",
			"[Mattermost](https://mattermost.com) ``` Mattermost ```\n\n",
		}, {
			"hello ` Mattermost ` goodbye",
			"hello ` Mattermost ` goodbye",
		}, {
			"hello\n`\nMattermost\n`\ngoodbye",
			"hello\n`\nMattermost\n`\ngoodbye",
		}, {
			"Mattermost ` Mattermost ` goodbye",
			"[Mattermost](https://mattermost.com) ` Mattermost ` goodbye",
		}, {
			"` Mattermost ` Mattermost",
			"` Mattermost ` [Mattermost](https://mattermost.com)",
		}, {
			"Mattermost ` Mattermost `",
			"[Mattermost](https://mattermost.com) ` Mattermost `",
		}, {
			"Mattermost ` Mattermost `\n\n",
			"[Mattermost](https://mattermost.com) ` Mattermost `\n\n",
		}, {
			"hello ``` Mattermost ``` goodbye ` Mattermost ` end",
			"hello ``` Mattermost ``` goodbye ` Mattermost ` end",
		}, {
			"hello\n```\nMattermost\n```\ngoodbye ` Mattermost ` end",
			"hello\n```\nMattermost\n```\ngoodbye ` Mattermost ` end",
		}, {
			"Mattermost ``` Mattermost ``` goodbye ` Mattermost ` end",
			"[Mattermost](https://mattermost.com) ``` Mattermost ``` goodbye ` Mattermost ` end",
		}, {
			"``` Mattermost ``` Mattermost",
			"``` Mattermost ``` [Mattermost](https://mattermost.com)",
		}, {
			"```\n` Mattermost `\n```\nMattermost",
			"```\n` Mattermost `\n```\n[Mattermost](https://mattermost.com)",
		}, {
			"  Mattermost",
			"  [Mattermost](https://mattermost.com)",
		}, {
			"    Mattermost",
			"    Mattermost",
		}, {
			"    ```\nMattermost\n    ```",
			"    ```\n[Mattermost](https://mattermost.com)\n    ```",
		}, {
			"` ``` `\nMattermost\n` ``` `",
			"` ``` `\n[Mattermost](https://mattermost.com)\n` ``` `",
		}, {
			"Mattermost \n Mattermost",
			"[Mattermost](https://mattermost.com) \n [Mattermost](https://mattermost.com)",
		}, {
			"[Mattermost](https://mattermost.com)",
			"[Mattermost](https://mattermost.com)",
		}, {
			"[  Mattermost  ](https://mattermost.com)",
			"[  Mattermost  ](https://mattermost.com)",
		}, {
			"[  Mattermost  ][1]\n\n[1]: https://mattermost.com",
			"[  Mattermost  ][1]\n\n[1]: https://mattermost.com",
		}, {
			"![  Mattermost  ](https://mattermost.com/example.png)",
			"![  Mattermost  ](https://mattermost.com/example.png)",
		}, {
			"![  Mattermost  ][1]\n\n[1]: https://mattermost.com/example.png",
			"![  Mattermost  ][1]\n\n[1]: https://mattermost.com/example.png",
		}, {
			"Why not visit https://mattermost.com?",
			"Why not visit [the Mattermost portal](https://mattermost.com)?",
		}, {
			"Why not visit https://www.mattermost.com?",
			"Why not visit [the Mattermost portal](https://mattermost.com)?",
		}, {
			"Please check https://mattermost.atlassian.net/browse/MM-123 for details",
			"Please check [MM-123](https://mattermost.atlassian.net/browse/MM-123) for details",
		}, {
			"Please check MM-123 for details",
			"Please check [MM-123](https://mattermost.atlassian.net/browse/MM-123) for details",
		}, {
			"Please check the ticket (MM-123) for details",
			"Please check the ticket ([MM-123](https://mattermost.atlassian.net/browse/MM-123)) for details",
		}, {
			"Don't confuse it with https://someone.elses/bugtracker/id=MM-123, which is not ours",
			"Don't confuse it with https://someone.elses/bugtracker/id=MM-123, which is not ours",
		},
	}

	for _, tt := range tests {
		post := &model.Post{
			Message: tt.inputMessage,
		}

		rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

		assert.Equal(t, tt.expectedMessage, rpost.Message)

		upost, _ := p.MessageWillBeUpdated(&plugin.Context{}, post, post)

		assert.Equal(t, tt.expectedMessage, upost.Message)
	}
}
