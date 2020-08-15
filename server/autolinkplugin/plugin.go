package autolinkplugin

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/utils/markdown"

	"github.com/mattermost/mattermost-plugin-autolink/server/api"
)

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin

	handler *api.Handler

	// configuration and a muttex to control concurrent access
	conf     *Config
	confLock sync.RWMutex
}

func New() *Plugin {
	return &Plugin{
		conf: new(Config),
	}
}

func (p *Plugin) OnActivate() error {
	p.handler = api.NewHandler(p, p)

	return nil
}

func (p *Plugin) IsAuthorizedAdmin(userID string) (bool, error) {
	user, err := p.API.GetUser(userID)
	if err != nil {
		return false, fmt.Errorf(
			"failed to obtain information about user `%s`: %w", userID, err)
	}
	if strings.Contains(user.Roles, "system_admin") {
		p.API.LogInfo(
			fmt.Sprintf("UserID `%s` is authorized basing on the sysadmin role membership", userID))
		return true, nil
	}

	conf := p.getConfig()
	if _, ok := conf.AdminUserIds[userID]; ok {
		p.API.LogInfo(
			fmt.Sprintf("UserID `%s` is authorized basing on the list of plugin admins list", userID))
		return true, nil
	}

	return false, nil
}

func (p *Plugin) getChannelInfo(channelID string) (string, string, error) {
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		return "", "", appErr
	}

	if channel.TeamId == "" {
		return channel.Name, "", nil
	}

	team, appErr := p.API.GetTeam(channel.TeamId)
	if appErr != nil {
		return "", "", appErr
	}

	return channel.Name, team.Name, nil
}

func (p *Plugin) inScope(scope []string, channelName string, teamName string) bool {
	if len(scope) == 0 {
		return true
	}

	if teamName == "" {
		return false
	}

	for _, teamChannel := range scope {
		split := strings.Split(teamChannel, "/")

		splitLength := len(split)

		if splitLength == 1 && split[0] == "" {
			return false
		}

		if splitLength == 1 && strings.EqualFold(split[0], teamName) {
			return true
		}

		scopeMatch := strings.EqualFold(split[0], teamName) && strings.EqualFold(split[1], channelName)
		if splitLength == 2 && scopeMatch {
			return true
		}
	}

	return false
}

func (p *Plugin) isBotUser(userID string) (bool, *model.AppError) {
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("failed to check if message for rewriting was send by a bot", "error", appErr)
		return false, appErr
	}

	return user.IsBot, nil
}

func (p *Plugin) ProcessPost(post *model.Post) (*model.Post, string) {
	conf := p.getConfig()
	postText := post.Message

	hasOneOrMoreScopes := false
	for _, link := range conf.Links {
		if len(link.Scope) > 0 {
			hasOneOrMoreScopes = true
			break
		}
	}

	channelName := ""
	teamName := ""
	if hasOneOrMoreScopes {
		var err error
		channelName, teamName, err = p.getChannelInfo(post.ChannelId)
		if err != nil {
			p.API.LogError("Failed to resolve scope", "error", err.Error())
		}
	}

	for _, link := range conf.Links {
		if !p.inScope(link.Scope, channelName, teamName) {
			continue
		}

		offset := 0
		markdown.Inspect(postText, func(node interface{}) bool {
			switch node.(type) {
			// never descend into the text content of a link/image
			case *markdown.InlineLink,
				*markdown.InlineImage,
				*markdown.ReferenceLink,
				*markdown.ReferenceImage:
				return false
			}

			origText := ""
			startPos := 0
			endPos := 0

			if autolinkNode, ok := node.(*markdown.Autolink); ok {
				if link.DisableInHyperlinks {
					return true
				}
				startPos, endPos = autolinkNode.RawDestination.Position+offset, autolinkNode.RawDestination.End+offset
				origText = postText[startPos:endPos]
				if autolinkNode.Destination() != origText {
					p.API.LogError(fmt.Sprintf("Markdown autolink did not match range text, '%s' != '%s'",
						autolinkNode.Destination(), origText))
					return true
				}
			} else if textNode, ok := node.(*markdown.Text); ok {
				startPos, endPos = textNode.Range.Position+offset, textNode.Range.End+offset
				origText = postText[startPos:endPos]
				if textNode.Text != origText {
					p.API.LogError(fmt.Sprintf("Markdown text did not match range text, '%s' != '%s'", textNode.Text,
						origText))
					return true
				}
			}

			if origText != "" {
				newText := origText

				newText = link.Replace(newText)
				if origText != newText {
					postText = postText[:startPos] + newText + postText[endPos:]
					offset += len(newText) - len(origText)
				}
			}

			return true
		})
	}

	if post.Message != postText {
		isBot, appErr := p.isBotUser(post.UserId)
		if appErr != nil {
			// NOTE: Not sure how we want to handle errors here, we can either:
			// * assume that occasional rewrites of Bot messages are ok
			// * assume that occasional not rewriting of all messages is ok
			// Let's assume for now that former is a lesser evil and carry on.
		} else if isBot {
			// We intentionally use a single if/else block so that the code is
			// more readable and does not rely on hidden side effect of
			// isBot==false when appErr!=nil.
			p.API.LogDebug("not rewriting message from bot", "userID", post.UserId)
			return nil, ""
		}

		post.Message = postText
	}

	post.Hashtags, _ = model.ParseHashtags(post.Message)

	return post, ""
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("Mattermost-Plugin-ID", c.SourcePluginId)
	p.handler.ServeHTTP(w, r)
}

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.ProcessPost(post)
}

// MessageWillBeUpdated is invoked when a message is updated by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBeUpdated(c *plugin.Context, post *model.Post, _ *model.Post) (*model.Post, string) {
	conf := p.getConfig()
	if !conf.EnableOnUpdate {
		return post, ""
	}

	return p.ProcessPost(post)
}
