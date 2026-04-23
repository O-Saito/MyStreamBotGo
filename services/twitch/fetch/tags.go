package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

type StreamTag struct {
	TagID                    string            `json:"tag_id"`
	IsAuto                   bool              `json:"is_auto"`
	LocalizationNames        map[string]string `json:"localization_names"`
	LocalizationDescriptions map[string]string `json:"localization_descriptions"`
}

func GetAllStreamTags(tagIDs []string, req *twitch.PaginationRequest) (*twitch.PaginationData[StreamTag], error) {
	opts := map[string]any{}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	if tagIDs != nil {
		opts["tag_id"] = tagIDs
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/tags/streams", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[StreamTag]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAllStreamTags: no params")
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[StreamTag] {
		GetAllStreamTags(tagIDs, &twitch.PaginationRequest{
			Cursor: result.Pagination.Cursor,
		})
		return result
	}

	return result, nil
}

func GetStreamTags(broadcasterID string) ([]StreamTag, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/streams/tags", opts)

	result, err := twitch.ExecuteRequest[struct {
		Data []StreamTag `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamTags: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return result.Data, nil
}
