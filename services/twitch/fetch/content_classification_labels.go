package twitch

import (
	"MyStreamBot/helpers"
	
)

type ContentClassificationLabel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GetContentClassificationLabelsResponse struct {
	Data []ContentClassificationLabel `json:"data"`
}

func GetContentClassificationLabels() ([]ContentClassificationLabel, error) {
	url := HelixBaseURL + "/content_classification_labels"
	result, err := ExecuteRequest[GetContentClassificationLabelsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetContentClassificationLabels: %v", err)
		return nil, err
	}

	return result.Data, nil
}