package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/services"
	"MyStreamBot/services/twitch"

	tf "MyStreamBot/services/twitch/fetch"
)

var TwitchFunctionList = []services.LuaFunction{
	// bot state
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

	// fetch.go
	{Name: "get_cache_user_chat_color", Fn: twitch.GetCacheUserChatColor},

	// fetch/user.go
	{Name: "get_user_data", Fn: func(username string) *tf.User {
		d, _ := tf.GetUser(nil, []string{username})
		return d
	}},
	{Name: "get_user_data_by_id", Fn: func(userid string) *tf.User {
		d, _ := tf.GetUser([]string{userid}, nil)
		return d
	}},
	// check
	{Name: "get_authorization_by_user", Fn: func(userId string) *tf.GetUserAuthorizationResponse {
		d, _ := tf.GetAuthorizationByUser([]string{userId})
		return d
	}},
	{Name: "get_user_block_list", Fn: tf.GetUserBlockList},
	{Name: "get_user_extensions", Fn: tf.GetUserExtensions},
	{Name: "get_user_active_extensions", Fn: tf.GetUserActiveExtensions},
	// TODO:
	// update_user -> UpdateUser
	// update_user_extensions -> UpdateUserExtensions
	// block_user -> BlockUser
	// unblock_user -> UnblockUser

	// fetch/channels.go
	{Name: "get_followers_data", Fn: func() []tf.TwitchViewerData {
		data, _ := tf.GetChannelFollowers("", nil)
		return data.Data
	}},
	{Name: "get_follower_data", Fn: func(userId string) ([]tf.TwitchViewerData, error) {
		data, err := tf.GetChannelFollowers(userId, nil)
		return data.Data, err
	}},
	// modify_channel_information -> ModifyChannelInformation
	// get_channel_editors -> GetChannelEditors
	// get_followed_channels -> GetFollowedChannels
	{Name: "get_channel_stream_data", Fn: func(userId string) tf.ChannelInfo {
		d, _ := tf.GetChannelInformation([]string{userId})
		return d[0]
	}},

	// fetch/chat.go
	{Name: "delete_message", Fn: tf.DeleteMessage},
	// get_chatters -> GetChatters
	// get_channel_emotes -> GetChannelEmotes
	// get_global_emotes -> GetGlobalEmotes
	// get_channel_chat_badges -> GetChannelChatBadges
	// get_global_chat_badges -> GetGlobalChatBadges
	// get_chat_settings -> GetChatSettings
	// update_chat_settings -> UpdateChatSettings
	// send_chat_announcement -> SendChatAnnouncement
	// send_shoutout -> SendShoutout
	// update_user_chat_color -> UpdateUserChatColor
	// get_emote_sets -> GetEmoteSets
	// send_chat_message -> SendChatMessage
	// get_shared_chat_session -> GetSharedChatSession
	// get_user_emotes -> GetUserEmotes
	// get_user_chat_color -> GetUserChatColor

	// fetch/moderation.go
	{Name: "ban_user", Fn: tf.BanUser},
	// get_banned_users -> GetBannedUsers
	// get_moderators -> GetModerators
	// get_vips -> GetVIPs
	// get_shield_mode_status -> GetShieldModeStatus
	// update_shield_mode_status -> UpdateShieldModeStatus

	// check_auto_mod_status -> CheckAutoModStatus
	// manage_held_auto_mod_messages -> ManageHeldAutoModMessages
	// get_auto_mod_settings -> GetAutoModSettings
	// update_auto_mod_settings -> UpdateAutoModSettings
	// unban_user -> UnbanUser
	// get_blocked_terms -> GetBlockedTerms
	// add_blocked_term -> AddBlockedTerm
	// remove_blocked_term -> RemoveBlockedTerm
	// get_moderated_channels -> GetModeratedChannels
	// add_channel_moderator -> AddChannelModerator
	// remove_channel_moderator -> RemoveChannelModerator
	// add_channel_vip -> AddChannelVIP
	// remove_channel_vip -> RemoveChannelVIP
	// warn_chat_user -> WarnChatUser
	// add_suspicious_status_to_chat_user -> AddSuspiciousStatusToChatUser
	// remove_suspicious_status_from_chat_user -> RemoveSuspiciousStatusFromChatUser
	// get_unban_requests -> GetUnbanRequests
	// resolve_unban_request -> ResolveUnbanRequest

	// fetch/search.go
	// search_categories -> SearchCategories
	// search_channels -> SearchChannels

	// fetch/streams.go
	// get_streams -> GetStreams
	// create_stream_marker -> CreateStreamMarker

	// fetch/polls.go
	// get_polls -> GetPolls
	// create_poll -> CreatePoll

	// fetch/predictions.go
	// get_predictions -> GetPredictions
	// create_prediction -> CreatePrediction

	// fetch/hype_train.go
	// get_hype_train_status -> GetHypeTrainStatus

	// fetch/goals.go
	// get_creator_goals -> GetCreatorGoals

	// fetch/whispers.go
	// send_whisper -> SendWhisper

	// fetch/streams.go (additional)
	// get_stream_key -> GetStreamKey
	// get_followed_streams -> GetFollowedStreams
	// get_stream_markers -> GetStreamMarkers

	// fetch/videos.go
	// get_videos -> GetVideos
	// delete_videos -> DeleteVideos

	// fetch/teams.go
	// get_channel_teams -> GetChannelTeams
	// get_teams -> GetTeams

	// fetch/tags.go
	// get_all_stream_tags -> GetAllStreamTags
	// get_stream_tags -> GetStreamTags

	// fetch/subscriptions.go
	// get_broadcaster_subscriptions -> GetBroadcasterSubscriptions
	// check_user_subscription -> CheckUserSubscription

	// fetch/schedule.go
	// get_channel_stream_schedule -> GetChannelStreamSchedule
	// get_channel_icalendar -> GetChanneliCalendar
	// update_channel_stream_schedule -> UpdateChannelStreamSchedule
	// create_channel_stream_schedule_segment -> CreateChannelStreamScheduleSegment
	// update_channel_stream_schedule_segment -> UpdateChannelStreamScheduleSegment
	// delete_channel_stream_schedule_segment -> DeleteChannelStreamScheduleSegment

	// fetch/raids.go
	// start_raid -> StartRaid
	// cancel_raid -> CancelRaid

	// fetch/predictions.go (additional)
	// end_prediction -> EndPrediction

	// fetch/polls.go (additional)
	// end_poll -> EndPoll

	// fetch/guest_star.go
	// get_channel_guest_star_settings -> GetChannelGuestStarSettings
	// update_channel_guest_star_settings -> UpdateChannelGuestStarSettings
	// get_guest_star_session -> GetGuestStarSession
	// create_guest_star_session -> CreateGuestStarSession
	// end_guest_star_session -> EndGuestStarSession

	// fetch/eventsub.go
	// get_event_subscriptions -> GetEventSubscriptions
	// delete_event_subscriptions -> DeleteEventSubscriptions

	// fetch/games.go
	// get_top_games -> GetTopGames
	// get_games -> GetGames

	// fetch/extensions.go
	// get_extension_configuration_segment -> GetExtensionConfigurationSegment
	// save_extension_configuration_segment -> SaveExtensionConfigurationSegment
	// send_extension_pub_sub_message -> SendExtensionPubSubMessage
	// get_extension_live_channels -> GetExtensionLiveChannels
	// get_extension_secrets -> GetExtensionSecrets
	// create_extension_secret -> CreateExtensionSecret
	// send_extension_chat_message -> SendExtensionChatMessage
	// get_extensions -> GetExtensions
	// get_released_extensions -> GetReleasedExtensions

	// fetch/entitlements.go
	// get_entitlements -> GetEntitlements
	// update_entitlement_brands -> UpdateEntitlementBrands

	// fetch/content_classification_labels.go
	// get_content_classification_labels -> GetContentClassificationLabels

	// fetch/conduits.go
	// create_conduit -> CreateConduit
	// get_conduits -> GetConduits
	// update_conduit -> UpdateConduit
	// delete_conduit -> DeleteConduit
	// get_conduit_users -> GetConduitUsers
	// update_conduit_users -> UpdateConduitUsers

	// fetch/clips.go
	// create_clip -> CreateClip
	// get_clips -> GetClips

	// fetch/charity.go
	// get_charity_campaign -> GetCharityCampaign
	// get_charity_donations -> GetCharityDonations
	// start_charity_campaign -> StartCharityCampaign
	// stop_charity_campaign -> StopCharityCampaign

	// fetch/channel_points.go
	// get_custom_reward -> GetCustomReward
	// get_custom_rewards -> GetCustomRewards
	// create_custom_reward -> CreateCustomReward
	// update_custom_reward -> UpdateCustomReward
	// delete_custom_reward -> DeleteCustomReward
	// get_custom_reward_redemption -> GetCustomRewardRedemption
	// get_custom_reward_redemptions -> GetCustomRewardRedemptions
	// update_custom_reward_redemption -> UpdateCustomRewardRedemption
	// get_automatic_reward_status -> GetAutomaticRewardStatus

	// fetch/bits.go
	// get_bits_leaderboard -> GetBitsLeaderboard
	// get_bits_leaderboard_by_time -> GetBitsLeaderboardByTime
	// get_extension_bits_leaderboard -> GetExtensionBitsLeaderboard
	// get_clips -> GetClips
	// generate_extension_dashboard_link -> GenerateExtensionDashboardLink

	// fetch/analytics.go
	// get_extension_analytics -> GetExtensionAnalytics
	// get_game_analytics -> GetGameAnalytics

	// fetch/ads.go
	// start_ad_session -> StartAdSession
	// get_ad_schedule -> GetAdSchedule
	// get_extension_ads -> GetExtensionAds
	// post_extension_ad -> PostExtensionAd

}
