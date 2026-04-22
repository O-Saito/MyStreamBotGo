package twitch

import (
	"fmt"
	twitch "MyStreamBot/services/twitch"
)

type Conduit struct {
	ID            string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ShardCount    int    `json:"shard_count"`
	HealthStatus  string `json:"health_status"`
}

type ConduitShard struct {
	ID              string `json:"id"`
	ConduitID       string `json:"conduit_id"`
	ShardIndex      int    `json:"shard_index"`
	ShardStatus     string `json:"shard_status"`
	TransportMethod string `json:"transport_method"`
	Transport       ConduitTransport `json:"transport"`
}

type ConduitTransport struct {
	Method        string            `json:"method"`
	Callback      string            `json:"callback,omitempty"`
	Secret        string            `json:"secret,omitempty"`
	SessionID     string            `json:"session_id,omitempty"`
	WebSocketURL  string            `json:"websocket_url,omitempty"`
}

type GetConduitsResponse struct {
	Data       []Conduit `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type GetConduitShardsResponse struct {
	Data []ConduitShard `json:"data"`
}

func GetConduits(first int, after string) ([]Conduit, error) {
	url := twitch.HelixBaseURL + "/conduits"
	if first > 0 {
		url += fmt.Sprintf("?first=%d", first)
		if after != "" {
			url += "&after=" + after
		}
	} else if after != "" {
		url += "?after=" + after
	}

	result, err := twitch.ExecuteRequest[GetConduitsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

type CreateConduitRequest struct {
	ShardCount int `json:"shard_count"`
}

func CreateConduits(shardCount int) (*Conduit, error) {
	url := twitch.HelixBaseURL + "/conduits"
	body := CreateConduitRequest{
		ShardCount: shardCount,
	}

	result, err := twitch.ExecuteJSONRequest[GetConduitsResponse, CreateConduitRequest]("POST", url, body, 201)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateConduits(conduitID string, shardCount int) error {
	url := fmt.Sprintf("%s/conduits?conduit_id=%s", twitch.HelixBaseURL, conduitID)
	body := map[string]any{
		"shard_count": shardCount,
	}

	_, err := twitch.ExecuteJSONRequest[map[string]any, map[string]any]("PATCH", url, body, 204)
	return err
}

func DeleteConduit(conduitID string) error {
	url := fmt.Sprintf("%s/conduits?conduit_id=%s", twitch.HelixBaseURL, conduitID)

	_, err := twitch.ExecuteRequestNoParse("DELETE", url, 204)
	return err
}

func GetConduitShards(conduitID string) ([]ConduitShard, error) {
	url := fmt.Sprintf("%s/conduits/shards?conduit_id=%s", twitch.HelixBaseURL, conduitID)

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
	url := fmt.Sprintf("%s/conduits/shards?conduit_id=%s", twitch.HelixBaseURL, conduitID)
	body := map[string]any{
		"shards": shards,
	}

	_, err := twitch.ExecuteJSONRequest[map[string]any, map[string]any]("PATCH", url, body, 204)
	return err
}