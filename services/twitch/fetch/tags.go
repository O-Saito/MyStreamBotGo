package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
)

type StreamTag struct {
	TagID                    string            `json:"tag_id"`
	IsAuto                   bool              `json:"is_auto"`
	LocalizationNames        map[string]string `json:"localization_names"`
	LocalizationDescriptions map[string]string `json:"localization_descriptions"`
}

func GetAllStreamTags(tagIDs []string, req *PaginationRequest) (*PaginationData[StreamTag], error) {
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

	url := BuildURL(HelixBaseURL+"/tags/streams", opts)

	result, err := ExecuteRequest[PaginationData[StreamTag]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAllStreamTags: no params")
		return nil, err
	}

	result.GetNext = func() *PaginationData[StreamTag] {
		n, _ := GetAllStreamTags(tagIDs, &PaginationRequest{
			Cursor: result.Pagination.Cursor,
		})
		return n
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

	url := BuildURL(HelixBaseURL+"/streams/tags", opts)

	result, err := ExecuteRequest[struct {
		Data []StreamTag `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamTags: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return result.Data, nil
}
