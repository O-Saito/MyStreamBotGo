package youtube

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strings"
	"time"
)

type YouTubeLiveChatMessagesResponse struct {
	Kind                  string                   `json:"kind"`
	Etag                  string                   `json:"etag"`
	NextPageToken         string                   `json:"nextPageToken"`
	OfflineAt             time.Time                `json:"offlineAt"`
	PollingIntervalMillis int                      `json:"pollingIntervalMillis"`
	Items                 []YouTubeLiveChatMessage `json:"items"`
}

type YouTubeLiveChatMessage struct {
	Kind          string                 `json:"kind"`
	Etag          string                 `json:"etag"`
	ID            string                 `json:"id"`
	Snippet       LiveChatMessageSnippet `json:"snippet"`
	AuthorDetails LiveChatAuthorDetails  `json:"authorDetails"`
}

type LiveChatMessageSnippet struct {
	Type               string                      `json:"type"`
	PublishedAt        time.Time                   `json:"publishedAt"`
	DisplayMessage     string                      `json:"displayMessage"`
	TextMessageDetails *LiveChatTextMessageDetails `json:"textMessageDetails"`
}

type LiveChatTextMessageDetails struct {
	MessageText string `json:"messageText"`
}

type LiveChatAuthorDetails struct {
	ChannelID       string `json:"channelId"`
	DisplayName     string `json:"displayName"`
	ProfileImageURL string `json:"profileImageUrl"`
	IsChatModerator bool   `json:"isChatModerator"`
	IsChatOwner     bool   `json:"isChatOwner"`
	IsChatSponsor   bool   `json:"isChatSponsor"`
	IsVerified      bool   `json:"isVerified"`
}

func defaultUserColor(username string) string {
	colors := []string{
		"#FF0000", "#0000FF", "#008000", "#B22222", "#FF7F50",
		"#9ACD32", "#FF4500", "#2E8B57", "#DAA520", "#D2691E",
		"#5F9EA0", "#1E90FF", "#FF69B4", "#8A2BE2", "#00FF7F",
	}

	h := fnv.New32a()
	h.Write([]byte(strings.ToLower(username)))
	index := int(h.Sum32()) % len(colors)

	return colors[index]
}

func FetchLiveChatMessages(liveChatID, pageToken string) (*YouTubeLiveChatMessagesResponse, error) {
	baseURL := "https://www.googleapis.com/youtube/v3/liveChat/messages"
	req, _ := http.NewRequest("GET", baseURL, nil)

	q := req.URL.Query()
	q.Add("liveChatId", liveChatID)
	q.Add("part", "snippet,authorDetails")
	if pageToken != "" {
		q.Add("pageToken", pageToken)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data YouTubeLiveChatMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func ListenToChat(id string) {
	helpers.Log(helpers.INFO, "Started yt chat listener!")
	page := ""
	for {
		helpers.Logf(helpers.DEBUG, "READING YT CHAT...")
		data, _ := FetchLiveChatMessages(id, page)

		for _, msg := range data.Items {
			messagedata := globals.MessageFromStream{
				Source:    "youtube",
				Channel:   id,
				User:      msg.AuthorDetails.DisplayName,
				UserId:    msg.AuthorDetails.ChannelID,
				MessageId: msg.ID,
				Message:   msg.Snippet.DisplayMessage,
				Metadata: map[string]any{
					"snippet":       msg.Snippet,
					"authorDetails": msg.AuthorDetails,
				},
			}

			state := globals.GetState()
			if messagedata.Metadata["color"] == nil || messagedata.Metadata["color"] == "" {
				userColor := state.GetData("youtube-user-color")
				if userColor == nil {
					userColor = make(map[string]any)
				}
				userColorMap, ok := userColor.(map[string]any)
				if !ok {
					userColorMap = make(map[string]any)
				}
				if userColorMap[messagedata.User] == nil {
					userColorMap[messagedata.User] = defaultUserColor(messagedata.User)
					state.SetData("youtube-user-color", userColorMap)
				}
				if colorVal, ok := userColorMap[messagedata.User].(string); ok {
					messagedata.Metadata["color"] = colorVal
				}
			}

			globals.ChatQueue <- messagedata
		}

		if !data.OfflineAt.IsZero() && data.OfflineAt.After(time.Now()) {
			helpers.Logf(helpers.WARN, "[YT OFF] Chat offline at %v, now is %v", data.OfflineAt, time.Now())
			globals.WsBroadcast <- globals.SocketMessage{
				Type: "youtube-live-offline", Data: map[string]any{"liveId": id, "offlineAt": data.OfflineAt},
			}
			break
		}

		page = data.NextPageToken
		helpers.Logf(helpers.DEBUG, "YT POLL %d", data.PollingIntervalMillis)

		if len(data.Items) == 0 {
			data.PollingIntervalMillis = 60000
		}

		if data.PollingIntervalMillis < 5000 {
			data.PollingIntervalMillis = 5000
		}

		helpers.Logf(helpers.DEBUG, "YT POLL NOW %d", data.PollingIntervalMillis)

		time.Sleep(time.Duration(data.PollingIntervalMillis) * time.Millisecond)
	}
}
