package twitch

import (
	twitch "MyStreamBot/services/twitch"
)

type Conduit struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ShardCount     int    `json:"shard_count"`
	HealthStatus   string `json:"health_status"`
}

type ConduitShard struct {
	ID              string          `json:"id"`
	ConduitID       string          `json:"conduit_id"`
	ShardIndex      int             `json:"shard_index"`
	ShardStatus     string          `json:"shard_status"`
	TransportMethod string          `json:"transport_method"`
	Transport       ConduitTransport `json:"transport"`
}

type ConduitTransport struct {
	Method       string `json:"method"`
	Callback     string `json:"callback,omitempty"`
	Secret       string `json:"secret,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	WebSocketURL string `json:"websocket_url,omitempty"`
}

func GetConduits(req *twitch.PaginationRequest) ([]Conduit, error) {
	opts := map[string]any{}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/conduits", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Conduit]]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

type CreateConduitRequest struct {
	ShardCount int `json:"shard_count"`
}

func CreateConduits(shardCount int) (*Conduit, error) {
	url := twitch.BuildURL(twitch.HelixBaseURL+"/conduits", nil)

	body := CreateConduitRequest{
		ShardCount: shardCount,
	}

	result, err := twitch.ExecuteJSONRequest[twitch.PaginationData[Conduit], CreateConduitRequest]("POST", url, body, 201)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateConduits(conduitID string, shardCount int) error {
	opts := map[string]any{
		"conduit_id": conduitID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/conduits", opts)

	body := map[string]any{
		"shard_count": shardCount,
	}

	_, err := twitch.ExecuteJSONRequest[map[string]any, map[string]any]("PATCH", url, body, 204)
	return err
}

func DeleteConduit(conduitID string) error {
	opts := map[string]any{
		"conduit_id": conduitID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/conduits", opts)

	_, err := twitch.ExecuteRequest[map[string]any]("DELETE", url, 204)
	return err
}

type GetConduitShardsResponse struct {
	Data []ConduitShard `json:"data"`
}

func GetConduitShards(conduitID string) ([]ConduitShard, error) {
	opts := map[string]any{
		"conduit_id": conduitID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/conduits/shards", opts)

	result, err := twitch.ExecuteRequest[GetConduitShardsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

type UpdateConduitShardRequest struct {
	Transport ConduitTransport `json:"transport"`
}

func UpdateConduitShards(conduitID string, shards []UpdateConduitShardRequest) error {
	opts := map[string]any{
		"conduit_id": conduitID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/conduits/shards", opts)

	body := map[string]any{
		"shards": shards,
	}

	_, err := twitch.ExecuteJSONRequest[map[string]any, map[string]any]("PATCH", url, body, 204)
	return err
}