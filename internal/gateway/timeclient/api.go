package timeclient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GetMe returns the authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	return doJSON[User](ctx, c, MeEndpoint)
}

// GetMyTeams returns teams the authenticated user belongs to.
func (c *Client) GetMyTeams(ctx context.Context) ([]Team, error) {
	out, err := doJSON[[]Team](ctx, c, MyTeamsEndpoint)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// GetTeamByName looks up a team by its URL slug.
func (c *Client) GetTeamByName(ctx context.Context, name string) (*Team, error) {
	return doJSON[Team](ctx, c, fmt.Sprintf(TeamByNameEndpoint, url.PathEscape(name)))
}

// GetUserChannelsForTeam returns channels the user is a member of for the given team.
func (c *Client) GetUserChannelsForTeam(ctx context.Context, userID, teamID string) ([]Channel, error) {
	out, err := doJSON[[]Channel](ctx, c, fmt.Sprintf(UserChannelsEndpoint,
		url.PathEscape(userID), url.PathEscape(teamID)))
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// GetUserByUsername resolves a user by their username (without leading '@').
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return doJSON[User](ctx, c, fmt.Sprintf(UserByUsernameEndpoint, url.PathEscape(username)))
}

// PostsQuery controls /channels/{id}/posts paging.
type PostsQuery struct {
	Since   int64  // unix ms; if >0 returns posts created/updated since this time
	After   string // post ID; returns posts after this one
	Before  string // post ID; returns posts before this one
	Page    int
	PerPage int
}

func (q PostsQuery) values() url.Values {
	v := url.Values{}
	if q.Since > 0 {
		v.Set("since", strconv.FormatInt(q.Since, 10))
	}
	if q.After != "" {
		v.Set("after", q.After)
	}
	if q.Before != "" {
		v.Set("before", q.Before)
	}
	if q.Page > 0 {
		v.Set("page", strconv.Itoa(q.Page))
	}
	if q.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(q.PerPage))
	}
	return v
}

// GetChannelPosts returns a page of posts for a channel.
func (c *Client) GetChannelPosts(ctx context.Context, channelID string, q PostsQuery) (*PostList, error) {
	endpoint := fmt.Sprintf(ChannelPostsEndpoint, url.PathEscape(channelID))
	if vals := q.values(); len(vals) > 0 {
		endpoint += "?" + vals.Encode()
	}
	return doJSON[PostList](ctx, c, endpoint)
}

// ResolveBotChannel finds the direct-message channel between the authenticated
// user and the bot identified by botUsername within the team teamName.
// teamName may be empty: the first team the user belongs to is used.
func (c *Client) ResolveBotChannel(
	ctx context.Context,
	teamName, botUsername string,
) (*User, *User, *Channel, error) {
	me, err := c.GetMe(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get me: %w", err)
	}

	var team *Team
	if teamName != "" {
		team, err = c.GetTeamByName(ctx, teamName)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get team %q: %w", teamName, err)
		}
	} else {
		teams, terr := c.GetMyTeams(ctx)
		if terr != nil {
			return nil, nil, nil, fmt.Errorf("get teams: %w", terr)
		}
		if len(teams) == 0 {
			return nil, nil, nil, errors.New("user belongs to no teams")
		}
		t := teams[0]
		team = &t
	}

	bot, err := c.GetUserByUsername(ctx, botUsername)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get bot %q: %w", botUsername, err)
	}

	channels, err := c.GetUserChannelsForTeam(ctx, me.ID, team.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list channels: %w", err)
	}

	dmName := directChannelName(me.ID, bot.ID)
	for i := range channels {
		ch := &channels[i]
		if ch.Type == "D" && ch.Name == dmName {
			return me, bot, ch, nil
		}
	}

	return nil, nil, nil, fmt.Errorf("direct channel with @%s not found in team %q", botUsername, team.Name)
}

// directChannelName builds the Mattermost DM channel name: sorted "<id1>__<id2>".
func directChannelName(a, b string) string {
	if strings.Compare(a, b) <= 0 {
		return a + "__" + b
	}
	return b + "__" + a
}
