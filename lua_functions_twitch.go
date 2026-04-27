package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/services"
	"MyStreamBot/services/twitch"

	tf "MyStreamBot/services/twitch/fetch"
)

var TwitchFunctionList = []services.LuaFunction{
	{Name: "get_state", Fn: func() TwitchStateLuaData {
		state := globals.GetState().GetTwitchUser()
		return TwitchStateLuaData{
			UserID:                 state.UserID,
			UserLogin:              state.UserLogin,
			DisplayName:            state.DisplayName,
			Type:                   state.Type,
			BroadcasterType:        state.BroadcasterType,
			Description:            state.Description,
			ProfileImageURL:        state.ProfileImageURL,
			ProfileOfflineImageURL: state.ProfileOfflineImageURL,
		}
	}},
	{Name: "get_cache_stream", Fn: func() TwitchStreamLuaData {
		stream := globals.GetState().GetTwitchUser().StreamDetails
		if stream == nil {
			return TwitchStreamLuaData{}
		}
		return TwitchStreamLuaData{
			ID:           stream.ID,
			GameId:       stream.GameId,
			GameName:     stream.GameName,
			Type:         stream.Type,
			Title:        stream.Title,
			Tags:         stream.Tags,
			ViewerCount:  stream.ViewerCount,
			StartedAt:    stream.StartedAt,
			Language:     stream.Language,
			ThumbnailURL: stream.ThumbnailURL,
			IsMature:     stream.IsMature,
		}
	}},
	{Name: "get_cache_user_chat_color", Fn: twitch.GetCacheUserChatColor},
	{Name: "get_user_data", Fn: func(username string) *tf.User {
		d, _ := tf.GetUser(nil, []string{username})
		return d
	}},
	{Name: "get_user_data_by_id", Fn: func(userid string) *tf.User {
		d, _ := tf.GetUser([]string{userid}, nil)
		return d
	}},
	{Name: "get_followers_data", Fn: func() []tf.TwitchViewerData {
		data, _ := tf.GetChannelFollowers("", nil)
		return data.Data
	}},
	{Name: "get_follower_data", Fn: func(userId string) ([]tf.TwitchViewerData, error) {
		data, err := tf.GetChannelFollowers(userId, nil)
		return data.Data, err
	}},
	{Name: "get_channel_stream_data", Fn: twitch.GetChannelStreamData},
	{Name: "ban_user", Fn: twitch.BanUser},
	{Name: "delete_message", Fn: twitch.DeleteMessage},
	{Name: "get_channel_followers", Fn: func(userId string) []tf.TwitchViewerData {
		data, _ := tf.GetChannelFollowers(userId, nil)
		return data.Data
	}},
}
