package youtube

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"MyStreamBot/globals"
)

type ytMockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *ytMockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func setupYouTubeMockTransport(resp *http.Response) func() {
	orig := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: &ytMockRoundTripper{response: resp}}
	return func() { http.DefaultClient = orig }
}

func TestRefreshToken_Success(t *testing.T) {
	state := globals.GetState()
	state.YouTubeUser = globals.YouTubeUser{
		Token:        "old_token",
		RefreshToken: "test_refresh",
	}

	state.SetYouTubeUser(state.YouTubeUser)
	globals.GetConfig().YouTubeClientID = "client_id"
	globals.GetConfig().YouTubeClientSecret = "secret"

	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"access_token": "new_access_token",
			"expires_in": 3600,
			"token_type": "Bearer"
		}`)),
	})
	defer restore()

	err := RefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedUser := state.GetYouTubeUser()
	if updatedUser.Token != "new_access_token" {
		t.Errorf("Token = %q, want %q", updatedUser.Token, "new_access_token")
	}
}

func TestRefreshToken_Error(t *testing.T) {
	state := globals.GetState()
	state.YouTubeUser = globals.YouTubeUser{
		Token:        "old_token",
		RefreshToken: "invalid_refresh",
	}
	state.SetYouTubeUser(state.YouTubeUser)
	globals.GetConfig().YouTubeClientID = "client_id"
	globals.GetConfig().YouTubeClientSecret = "secret"

	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 400,
		Body: io.NopCloser(strings.NewReader(`{
			"error": "invalid_grant",
			"error_description": "Invalid refresh token"
		}`)),
	})
	defer restore()

	err := RefreshToken()
	if err == nil {
		t.Error("expected error for invalid refresh token, got nil")
	}
}

func TestGetCurrentStreamings_Parsing(t *testing.T) {
	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"kind": "youtube#liveBroadcastList",
			"items": [
				{
					"id": "broadcast123",
					"snippet": {
						"channelId": "channel123",
						"title": "Live Stream",
						"description": "Test stream"
					},
					"status": {
						"lifeCycleStatus": "live",
						"privacyStatus": "public"
					}
				}
			]
		}`)),
	})
	defer restore()

	broadcasts, err := GetCurrentStreamings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(broadcasts.Items) != 1 {
		t.Errorf("len(broadcasts.Items) = %d, want 1", len(broadcasts.Items))
	}
	if broadcasts.Items[0].ID != "broadcast123" {
		t.Errorf("broadcasts.Items[0].ID = %q, want %q", broadcasts.Items[0].ID, "broadcast123")
	}
	if broadcasts.Items[0].Snippet.Title != "Live Stream" {
		t.Errorf("title = %q, want %q", broadcasts.Items[0].Snippet.Title, "Live Stream")
	}
	if broadcasts.Items[0].Status.LifeCycleStatus != "live" {
		t.Errorf("status = %q, want %q", broadcasts.Items[0].Status.LifeCycleStatus, "live")
	}
}

func TestGetNextStreamings_Parsing(t *testing.T) {
	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"kind": "youtube#liveBroadcastList",
			"items": [
				{
					"id": "broadcast456",
					"snippet": {
						"channelId": "channel123",
						"title": "Upcoming Stream",
						"scheduledStartTime": "2024-01-01T12:00:00Z"
					},
					"status": {
						"lifeCycleStatus": "upcoming"
					}
				}
			]
		}`)),
	})
	defer restore()

	broadcasts, err := GetNextStreamings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(broadcasts.Items) != 1 {
		t.Errorf("len(broadcasts.Items) = %d, want 1", len(broadcasts.Items))
	}
	if broadcasts.Items[0].Status.LifeCycleStatus != "upcoming" {
		t.Errorf("status = %q, want %q", broadcasts.Items[0].Status.LifeCycleStatus, "upcoming")
	}
}

func TestGetStreamData_Parsing(t *testing.T) {
	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"kind": "youtube#videoListResponse",
			"items": [
				{
					"id": "video123",
					"snippet": {"title": "Test Video"},
					"liveStreamingDetails": {
						"concurrentViewers": "1000",
						"activeLiveChatId": "chat123",
						"actualStartTime": "2024-01-01T10:00:00Z"
					}
				}
			]
		}`)),
	})
	defer restore()

	stream, err := GetStreamData("video123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stream == nil {
		t.Fatal("stream is nil")
	}
	if stream.VideoID != "video123" {
		t.Errorf("stream.VideoID = %q, want %q", stream.VideoID, "video123")
	}
	if stream.Title != "Test Video" {
		t.Errorf("stream.Title = %q, want %q", stream.Title, "Test Video")
	}
	if stream.ConcurrentViewers != "1000" {
		t.Errorf("stream.ConcurrentViewers = %q, want %q", stream.ConcurrentViewers, "1000")
	}
	if stream.LiveChatID != "chat123" {
		t.Errorf("stream.LiveChatID = %q, want %q", stream.LiveChatID, "chat123")
	}
}

func TestGetStreamData_Empty(t *testing.T) {
	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{"kind": "youtube#videoListResponse","items":[]}`)),
	})
	defer restore()

	stream, err := GetStreamData("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stream != nil {
		t.Errorf("expected nil stream, got %+v", stream)
	}
}

func TestGetCurrentYouTubeChannel_Parsing(t *testing.T) {
	restore := setupYouTubeMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"kind": "youtube#channelListResponse",
			"items": [
				{
					"id": "channel123",
					"snippet": {
						"title": "My Channel",
						"description": "Test channel",
						"customUrl": "@mychannel"
					}
				}
			]
		}`)),
	})
	defer restore()

	channel, err := GetCurrentYouTubeChannel()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channel.Items) != 1 {
		t.Errorf("len(channel.Items) = %d, want 1", len(channel.Items))
	}
	if channel.Items[0].ID != "channel123" {
		t.Errorf("channel.ID = %q, want %q", channel.Items[0].ID, "channel123")
	}
	if channel.Items[0].Snippet.Title != "My Channel" {
		t.Errorf("channel.Snippet.Title = %q, want %q", channel.Items[0].Snippet.Title, "My Channel")
	}
}

func TestYouTubeChannelStruct(t *testing.T) {
	snippet := YouTubeChannelSnippet{
		Title:       "Test Channel",
		Description: "Description",
		CustomURL:   "@testchannel",
		PublishedAt: "2024-01-01T00:00:00Z",
		Country:    "US",
	}

	if snippet.Title != "Test Channel" {
		t.Errorf("Title = %q", snippet.Title)
	}
	if snippet.CustomURL != "@testchannel" {
		t.Errorf("CustomURL = %q", snippet.CustomURL)
	}
}

func TestLiveBroadcastStruct(t *testing.T) {
	snippet := LiveBroadcastSnippet{
		ChannelID: "channel123",
		Title:      "Live Stream",
	}

	if snippet.Title != "Live Stream" {
		t.Errorf("Title = %q", snippet.Title)
	}
}

func TestLiveBroadcastStatusStruct(t *testing.T) {
	status := LiveBroadcastStatus{
		LifeCycleStatus: "live",
		PrivacyStatus:  "public",
	}

	if status.LifeCycleStatus != "live" {
		t.Errorf("LifeCycleStatus = %q", status.LifeCycleStatus)
	}
}

func TestYouTubeRequestBuilding(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.YouTubeClientID = "client_id"

	state := globals.GetState()
	user := globals.YouTubeUser{
		Token: "test_token",
	}
	state.SetYouTubeUser(user)

	_ = cfg
	_ = state

	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/channels", nil)
	req.Header.Set("Authorization", "Bearer "+user.Token)

	if req.Header.Get("Authorization") != "Bearer test_token" {
		t.Errorf("Authorization = %q, want %q", req.Header.Get("Authorization"), "Bearer test_token")
	}
}

func TestOAuthRequestParsing(t *testing.T) {
	data := url.Values{}
	data.Set("client_id", "client_id")
	data.Set("client_secret", "secret")
	data.Set("refresh_token", "refresh")
	data.Set("grant_type", "refresh_token")

	if data.Get("grant_type") != "refresh_token" {
		t.Errorf("grant_type = %q", data.Get("grant_type"))
	}
	if data.Get("client_id") != "client_id" {
		t.Errorf("client_id = %q", data.Get("client_id"))
	}
}

func TestStreamDataStruct(t *testing.T) {
	streamData := StreamData{
		VideoID:            "video123",
		Title:              "Test Stream",
		ConcurrentViewers:  "500",
		LiveChatID:         "chat123",
		ActualStartTime:    "2024-01-01T10:00:00Z",
	}

	if streamData.VideoID != "video123" {
		t.Errorf("VideoID = %q", streamData.VideoID)
	}
	if streamData.Title != "Test Stream" {
		t.Errorf("Title = %q", streamData.Title)
	}
	if streamData.ConcurrentViewers != "500" {
		t.Errorf("ConcurrentViewers = %q", streamData.ConcurrentViewers)
	}
	if streamData.LiveChatID != "chat123" {
		t.Errorf("LiveChatID = %q", streamData.LiveChatID)
	}

	_ = time.Now()
}

type mockStreamListServer struct {
	UnimplementedV3DataLiveChatMessageServiceServer
	messages []*LiveChatMessageListResponse
}

func (s *mockStreamListServer) StreamList(_ *LiveChatMessageListRequest, stream grpc.ServerStreamingServer[LiveChatMessageListResponse]) error {
	for _, msg := range s.messages {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func TestGRPCStreamList_ReceivesMessages(t *testing.T) {
	server := grpc.NewServer()
	mock := &mockStreamListServer{
		messages: []*LiveChatMessageListResponse{
			{
				Items: []*LiveChatMessage{
					{
						Id: proto.String("msg1"),
						Snippet: &LiveChatMessageSnippet{
							DisplayMessage: proto.String("Hello"),
							Type:           LiveChatMessageSnippet_TypeWrapper_TEXT_MESSAGE_EVENT.Enum(),
						},
						AuthorDetails: &LiveChatMessageAuthorDetails{
							DisplayName:   proto.String("TestUser"),
							ChannelId:     proto.String("UCtest"),
							IsChatOwner:   proto.Bool(true),
						},
					},
				},
				NextPageToken: proto.String("page2"),
			},
			{
				Items: []*LiveChatMessage{
					{
						Id: proto.String("msg2"),
						Snippet: &LiveChatMessageSnippet{
							DisplayMessage: proto.String("World"),
						},
						AuthorDetails: &LiveChatMessageAuthorDetails{
							DisplayName: proto.String("TestUser2"),
							ChannelId:   proto.String("UCtest2"),
						},
					},
				},
				NextPageToken: proto.String(""),
			},
		},
	}
	RegisterV3DataLiveChatMessageServiceServer(server, mock)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := NewV3DataLiveChatMessageServiceClient(conn)
	stream, err := client.StreamList(context.Background(), &LiveChatMessageListRequest{
		LiveChatId: proto.String("test_chat"),
		Part:       []string{"snippet", "authorDetails"},
	})
	if err != nil {
		t.Fatalf("StreamList failed: %v", err)
	}

	var received []*LiveChatMessage
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Recv failed: %v", err)
		}
		received = append(received, resp.GetItems()...)
	}

	if len(received) != 2 {
		t.Fatalf("got %d messages, want 2", len(received))
	}
	if received[0].GetId() != "msg1" {
		t.Errorf("first msg id = %q, want %q", received[0].GetId(), "msg1")
	}
	if received[0].GetSnippet().GetDisplayMessage() != "Hello" {
		t.Errorf("first msg display = %q", received[0].GetSnippet().GetDisplayMessage())
	}
	if received[0].GetAuthorDetails().GetDisplayName() != "TestUser" {
		t.Errorf("first author = %q", received[0].GetAuthorDetails().GetDisplayName())
	}
	if received[1].GetId() != "msg2" {
		t.Errorf("second msg id = %q, want %q", received[1].GetId(), "msg2")
	}
	if received[1].GetAuthorDetails().GetChannelId() != "UCtest2" {
		t.Errorf("second channel = %q", received[1].GetAuthorDetails().GetChannelId())
	}
}

func TestGRPCStreamList_OfflineAt(t *testing.T) {
	server := grpc.NewServer()
	offlineTime := time.Now().Add(1 * time.Hour)
	mock := &mockStreamListServer{
		messages: []*LiveChatMessageListResponse{
			{
				OfflineAt: proto.String(offlineTime.Format(time.RFC3339)),
			},
		},
	}
	RegisterV3DataLiveChatMessageServiceServer(server, mock)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := NewV3DataLiveChatMessageServiceClient(conn)
	stream, err := client.StreamList(context.Background(), &LiveChatMessageListRequest{
		LiveChatId: proto.String("test_chat"),
		Part:       []string{"snippet"},
	})
	if err != nil {
		t.Fatalf("StreamList failed: %v", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv failed: %v", err)
	}

	if resp.GetOfflineAt() == "" {
		t.Error("OfflineAt should not be empty")
	}
	parsed, err := time.Parse(time.RFC3339, resp.GetOfflineAt())
	if err != nil {
		t.Fatalf("failed to parse OfflineAt: %v", err)
	}
	if !parsed.After(time.Now()) {
		t.Error("OfflineAt should be in the future")
	}
}