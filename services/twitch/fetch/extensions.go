package twitch

import (
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"time"
)

var (
	urlAPIExtensions          = twitch.HelixBaseURL + "/extensions"
	urlAPIExtensionConfig     = twitch.HelixBaseURL + "/extensions/config"
	urlAPIExtensionPubSub   = twitch.HelixBaseURL + "/extensions/pubsub"
	urlAPIExtensionLiveChannels = twitch.HelixBaseURL + "/extensions/live"
	urlAPIExtensionSecrets   = twitch.HelixBaseURL + "/extensions/secrets"
	urlAPIExtensionChat     = twitch.HelixBaseURL + "/extensions/chat"
	urlAPIReleasedExtensions = twitch.HelixBaseURL + "/extensions/released"
	urlAPIExtensionBitsProducts = twitch.HelixBaseURL + "/extensions/bits"
)

type ExtensionConfigSegment struct {
	Segment string `json:"segment"`
	Content string `json:"content"`
	Version string `json:"version"`
}

type ExtensionLiveChannel struct {
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName  string    `json:"broadcaster_name"`
	BroadcasterLogin string   `json:"broadcaster_login"`
	GameID          string    `json:"game_id"`
	GameName        string    `json:"game_name"`
	Title           string    `json:"title"`
	ViewerCount     int       `json:"viewer_count"`
	StartedAt      time.Time `json:"started_at"`
}

type ExtensionSecret struct {
	ActiveAt  time.Time `json:"active_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Secret   string   `json:"secret"`
}

type Extension struct {
	Authors                []ExtensionAuthor         `json:"authors"`
	BitsEnabled            bool                      `json:"bits_enabled"`
	CanInstall             bool                      `json:"can_install"`
	ConfigurationLocation string                    `json:"configuration_location"`
	DefaultPanel          string                    `json:"default_panel"`
	Descriptions          ExtensionDescriptions      `json:"descriptions"`
	EULA                  string                    `json:"eula"`
	IconURL               string                    `json:"icon_url"`
	IconURLSmall          string                    `json:"icon_url_small"`
	IdleIconURL           string                    `json:"idle_icon_url"`
	IconURLMedium         string                    `json:"icon_url_medium"`
	LargeIconURL          string                    `json:"large_icon_url"`
	Screenshots           []ExtensionScreenshot     `json:"screenshots"`
	AllowListEntry       ExtensionAllowListEntry    `json:"allow_list_entry"`
	Installed            bool                     `json:"installed"`
	LatestVersion        string                   `json:"latest_version"`
	Name                 string                   `json:"name"`
	RequiredConfiguration string                  `json:"required_configuration"`
	SecretVersion        int                      `json:"secret_version"`
	State                string                   `json:"state"`
	SupportURL           string                   `json:"support_url"`
	Tags                 []string                 `json:"tags"`
	ViewTokens           []string                 `json:"view_tokens"`
	Version              string                   `json:"version"`
	Views                ExtensionViews            `json:"views"`
}

type ExtensionAuthor struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ExtensionDescriptions struct {
	Text string `json:"text"`
	HTML string `json:"html"`
}

type ExtensionScreenshot struct {
	ImageURL string `json:"image_url"`
	ThumbURL string `json:"thumb_url"`
}

type ExtensionAllowListEntry struct {
	CanEmbed   bool     `json:"can_embed"`
	ChannelID  string   `json:"channel_id"`
	Javascript string   `json:"javascript"`
	Partner    bool     `json:"partner"`
	Secret     string   `json:"secret"`
	Screenshots []string `json:"screenshots"`
}

type ExtensionViews struct {
	Panel        ExtensionViewConfig `json:"panel"`
	VideoOverlay ExtensionViewConfig `json:"video_overlay"`
	Component   ExtensionViewConfig `json:"component"`
}

type ExtensionViewConfig struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	URL    string `json:"url"`
}

type ExtensionBitsProduct struct {
	Sku            string `json:"sku"`
	Cost           int    `json:"cost"`
	InDevelopment bool   `json:"in_development"`
	Name          string `json:"name"`
	Description  string `json:"description"`
	IsEnabled     bool   `json:"is_enabled"`
}

type ExtensionConfigResponse struct {
	Data []ExtensionConfigSegment `json:"data"`
}

type ExtensionLiveChannelsResponse struct {
	Data []ExtensionLiveChannel `json:"data"`
}

type ExtensionSecretsResponse struct {
	Data []ExtensionSecret `json:"data"`
}

type ExtensionsResponse struct {
	Data []Extension `json:"data"`
}

type ExtensionBitsProductsResponse struct {
	Data []ExtensionBitsProduct `json:"data"`
}

type VoidResponse struct{}

func GetExtensionConfigurationSegment(extensionID, segment string) (*ExtensionConfigSegment, error) {
	url := fmt.Sprintf("%s?extension_id=%s&segment=%s", urlAPIExtensionConfig, extensionID, segment)
	result, err := twitch.ExecuteRequest[ExtensionConfigResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, nil
	}
	return &result.Data[0], nil
}

type SetExtensionConfigRequest struct {
	ExtensionID string `json:"extension_id"`
	Segment   string `json:"segment"`
	Content   string `json:"content"`
	Version   string `json:"version"`
}

func SetExtensionConfigurationSegment(extensionID, segment, content, version string) error {
	req := SetExtensionConfigRequest{
		ExtensionID: extensionID,
		Segment:     segment,
		Content:     content,
		Version:     version,
	}
	_, err := twitch.ExecuteJSONRequest[VoidResponse, SetExtensionConfigRequest]("PUT", urlAPIExtensionConfig, req, 204)
	return err
}

type SetRequiredConfigRequest struct {
	ExtensionID           string `json:"extension_id"`
	RequiredConfiguration string `json:"required_configuration"`
}

func SetExtensionRequiredConfiguration(extensionID, requiredConfiguration string) error {
	req := SetRequiredConfigRequest{
		ExtensionID:           extensionID,
		RequiredConfiguration: requiredConfiguration,
	}
	_, err := twitch.ExecuteJSONRequest[VoidResponse, SetRequiredConfigRequest]("PUT", urlAPIExtensions+"/required_configuration", req, 204)
	return err
}

type PubSubMessageRequest struct {
	ExtensionID string `json:"extension_id"`
	Target      string `json:"target"`
	Message    string `json:"message"`
}

func SendExtensionPubSubMessage(extensionID, target string, message string) error {
	req := PubSubMessageRequest{
		ExtensionID: extensionID,
		Target:      target,
		Message:     message,
	}
	_, err := twitch.ExecuteJSONRequest[VoidResponse, PubSubMessageRequest]("POST", urlAPIExtensionPubSub, req, 204)
	return err
}

func GetExtensionLiveChannels(extensionID string) ([]ExtensionLiveChannel, error) {
	url := fmt.Sprintf("%s?extension_id=%s", urlAPIExtensionLiveChannels, extensionID)
	result, err := twitch.ExecuteRequest[ExtensionLiveChannelsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func GetExtensionSecrets(extensionID string) ([]ExtensionSecret, error) {
	url := fmt.Sprintf("%s?extension_id=%s", urlAPIExtensionSecrets, extensionID)
	result, err := twitch.ExecuteRequest[ExtensionSecretsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

type CreateSecretRequest struct {
	ExtensionID string `json:"extension_id"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

func CreateExtensionSecret(extensionID string, extendsInSecs int) error {
	req := CreateSecretRequest{ExtensionID: extensionID}
	if extendsInSecs > 0 {
		req.ExpiresIn = extendsInSecs
	}
	_, err := twitch.ExecuteJSONRequest[VoidResponse, CreateSecretRequest]("POST", urlAPIExtensionSecrets, req, 201)
	return err
}

type ChatMessageRequest struct {
	ExtensionID    string `json:"extension_id"`
	BroadcasterID  string `json:"broadcaster_id"`
	Text          string `json:"text"`
}

func SendExtensionChatMessage(extensionID, broadcasterID, text string) error {
	req := ChatMessageRequest{
		ExtensionID:   extensionID,
		BroadcasterID: broadcasterID,
		Text:          text,
	}
	_, err := twitch.ExecuteJSONRequest[VoidResponse, ChatMessageRequest]("POST", urlAPIExtensionChat, req, 204)
	return err
}

func GetExtensions(extensionID string) ([]Extension, error) {
	url := urlAPIExtensions
	if extensionID != "" {
		url = fmt.Sprintf("%s?extension_id=%s", url, extensionID)
	}
	result, err := twitch.ExecuteRequest[ExtensionsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func GetReleasedExtensions(extensionType string) ([]Extension, error) {
	url := urlAPIReleasedExtensions
	if extensionType != "" {
		url = fmt.Sprintf("%s?extension_type=%s", url, extensionType)
	}
	result, err := twitch.ExecuteRequest[ExtensionsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func GetExtensionBitsProducts(extensionID string) ([]ExtensionBitsProduct, error) {
	url := fmt.Sprintf("%s?extension_id=%s", urlAPIExtensionBitsProducts, extensionID)
	result, err := twitch.ExecuteRequest[ExtensionBitsProductsResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

type UpdateExtensionBitsProductRequest struct {
	Sku            string `json:"sku"`
	Cost           int    `json:"cost"`
	IsEnabled      bool   `json:"is_enabled"`
	InDevelopment bool   `json:"in_development"`
}

func UpdateExtensionBitsProduct(extensionID string, req UpdateExtensionBitsProductRequest) error {
	url := fmt.Sprintf("%s?extension_id=%s", urlAPIExtensionBitsProducts, extensionID)
	_, err := twitch.ExecuteJSONRequest[VoidResponse, UpdateExtensionBitsProductRequest]("PUT", url, req, 204)
	return err
}