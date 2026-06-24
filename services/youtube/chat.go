package youtube

import (
	"context"
	"hash/fnv"
	"io"
	"slices"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"MyStreamBot/globals"
	"MyStreamBot/helpers"
)

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

func ListenToChat(channelId, chatId string) {
	globals.GetState().AddYouTubeChat(channelId, chatId)
	helpers.Logf(helpers.INFO, "Started Youtube chat listener! channel: %s; chat: %s", channelId, chatId)

	conn, err := grpc.NewClient("youtube.googleapis.com:443",
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[YT] failed to dial gRPC: %v", err)
		return
	}
	defer conn.Close()

	client := NewV3DataLiveChatMessageServiceClient(conn)

	var pageToken string
	for {
		user := globals.GetState().GetYouTubeUser()

		ctx := metadata.NewOutgoingContext(context.Background(),
			metadata.Pairs("authorization", "Bearer "+user.Token))

		req := &LiveChatMessageListRequest{
			LiveChatId: proto.String(chatId),
			Part:       []string{"snippet", "authorDetails"},
		}
		if pageToken != "" {
			req.PageToken = proto.String(pageToken)
		}

		stream, err := client.StreamList(ctx, req)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[YT] StreamList failed: %v", err)
			_, foundChannel := helpers.Find(globals.GetState().GetYouTubeUser().Channels, func(c *globals.YouTubeChannel) bool {
				return slices.Contains(c.ChatIDs, chatId)
			})
			if !foundChannel {
				break
			}
			time.Sleep(5 * time.Second)
			continue
		}

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				helpers.Logf(helpers.ERROR, "[YT] stream.Recv failed: %v", err)
				break
			}

			pageToken = resp.GetNextPageToken()

			for _, msg := range resp.GetItems() {
				messagedata := globals.MessageFromStream{
					Source:    "youtube",
					Channel:   chatId,
					User:      msg.GetAuthorDetails().GetDisplayName(),
					UserId:    msg.GetAuthorDetails().GetChannelId(),
					MessageId: msg.GetId(),
					Message:   msg.GetSnippet().GetDisplayMessage(),
					Metadata: map[string]any{
						"snippet":       msg.GetSnippet(),
						"authorDetails": msg.GetAuthorDetails(),
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

			offlineAtStr := resp.GetOfflineAt()
			if offlineAtStr != "" {
				offlineAt, err := time.Parse(time.RFC3339, offlineAtStr)
				if err == nil && time.Now().After(offlineAt) {
					helpers.Logf(helpers.WARN, "[YT OFF] Chat offline at %v, now is %v", offlineAt, time.Now())
				globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{
					Type: "youtube-live-offline", Data: map[string]any{"liveId": chatId, "offlineAt": offlineAt},
				}, "WsBroadcast", false)
					return
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
	helpers.Logf(helpers.INFO, "Stopped Youtube chat listener! channel: %s; chat: %s", channelId, chatId)
}
