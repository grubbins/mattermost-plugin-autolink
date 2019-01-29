package main

import (
	"fmt"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/utils/markdown"
)

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin

	link_titles  atomic.Value
	link_creates atomic.Value
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var c Configuration
	err := p.API.LoadPluginConfiguration(&c)
	if err != nil {
		return err
	}

	link_titles := make([]*RegexReplacer, 0)
	for _, l := range c.LinkTitles {
		al, lerr := NewLinkTitleRegexReplacer(l)
		if lerr != nil {
			mlog.Error("Error creating regex replacer: ")
		}

		link_titles = append(link_titles, al)
	}

	link_creates := make([]*RegexReplacer, 0)
	for _, l := range c.LinkCreates {
		al, lerr := NewLinkCreateRegexReplacer(l)
		if lerr != nil {
			mlog.Error("Error creating regex replacer: ")
		}

		link_creates = append(link_creates, al)
	}

	p.link_creates.Store(link_creates)
	p.link_titles.Store(link_titles)
	return nil
}

func (p *Plugin) processPost(c *plugin.Context, post *model.Post) (*model.Post, string) {
	link_creates := p.link_creates.Load().([]*RegexReplacer)
	link_titles := p.link_titles.Load().([]*RegexReplacer)

	postText := post.Message
	offset := 0

	doReplacements := func(linker *RegexReplacer, nodeText string, textRange markdown.Range) bool {
		startPos, endPos := textRange.Position+offset, textRange.End+offset
		origText := postText[startPos:endPos]
		if nodeText != origText {
			// TODO: ignore if the difference is because of http:// prefix
			mlog.Error(fmt.Sprintf("Markdown text did not match range text, '%s' != '%s'", nodeText, origText))
			return false
		}

		newText := origText
		newText = linker.Replace(newText)

		if origText != newText {
			postText = postText[:startPos] + newText + postText[endPos:]
			offset += len(newText) - len(origText)
			return true
		}

		return false
	}

	for _, l := range link_creates {
		offset = 0
		markdown.Inspect(postText, func(node interface{}) bool {
			switch thisnode := node.(type) {
			// never descend into the text content of a link/image
			case *markdown.InlineLink:
				return false
			case *markdown.InlineImage:
				return false
			case *markdown.ReferenceLink:
				return false
			case *markdown.ReferenceImage:
				return false
			case *markdown.Text:
				doReplacements(l, thisnode.Text, thisnode.Range)
			case *markdown.Autolink:
				return false
			}

			return true
		})
	}

	for _, l := range link_titles {
		offset = 0
		markdown.Inspect(postText, func(node interface{}) bool {
			switch thisnode := node.(type) {
			// never descend into the text content of a link/image
			case *markdown.InlineLink:
				return false
			case *markdown.InlineImage:
				return false
			case *markdown.ReferenceLink:
				return false
			case *markdown.ReferenceImage:
				return false
			case *markdown.Text:
				return false
			case *markdown.Autolink:
				doReplacements(l, thisnode.Destination(), thisnode.RawDestination)
			}

			return true
		})
	}
	post.Message = postText

	return post, ""
}

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.processPost(c, post)
}

// MessageWillBeUpdated is invoked when a message is updated by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBeUpdated(c *plugin.Context, post *model.Post, _ *model.Post) (*model.Post, string) {
	return p.processPost(c, post)
}
