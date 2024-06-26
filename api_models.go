package moonshot

import (
	"context"

	"github.com/northes/gox/httpx"
)

type models struct {
	client *httpx.Client
}

func (c *Client) Models() *models {
	return &models{
		client: c.newHTTPClient(),
	}
}

type ListModelsRequest struct {
}

type ListModelsResponse struct {
	CommonResponse
	Object string                   `json:"object"`
	Data   []*ListModelResponseData `json:"data"`
}

type ListModelResponseData struct {
	Created    int                                 `json:"created"`
	Id         string                              `json:"id"`
	Object     string                              `json:"object"`
	OwnedBy    string                              `json:"owned_by"`
	Permission []*ListModelsResponseDataPermission `json:"permission"`
	Root       string                              `json:"root"`
	Parent     string                              `json:"parent"`
}

type ListModelsResponseDataPermission struct {
	Created            int    `json:"created"`
	Id                 string `json:"id"`
	Object             string `json:"object"`
	AllowCreateEngine  bool   `json:"allow_create_engine"`
	AllowSampling      bool   `json:"allow_sampling"`
	AllowLogprobs      bool   `json:"allow_logprobs"`
	AllowSearchIndices bool   `json:"allow_search_indices"`
	AllowView          bool   `json:"allow_view"`
	AllowFineTuning    bool   `json:"allow_fine_tuning"`
	Organization       string `json:"organization"`
	Group              string `json:"group"`
	IsBlocking         bool   `json:"is_blocking"`
}

func (m *models) List(ctx context.Context) (*ListModelsResponse, error) {
	const path = "/v1/models"
	listModelResp := new(ListModelsResponse)
	resp, err := m.client.AddPath(path).Get()
	if err != nil {
		return nil, err
	}
	if !resp.StatusOK() {
		return nil, StatusCodeToError(resp.Raw().StatusCode)
	}
	err = resp.Unmarshal(listModelResp)
	if err != nil {
		return nil, err
	}
	return listModelResp, nil
}
