package imagebed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://7bu.top/api/v1"

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = defaultBaseURL
	}
	return &Client{
		baseURL: base,
		token:   strings.TrimSpace(token),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c.token != ""
}

type Links struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type ImageItem struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	OriginName string `json:"origin_name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	HumanDate  string `json:"human_date"`
	Date       string `json:"date"`
	Links      Links  `json:"links"`
}

type ImagesPage struct {
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
	PerPage     int         `json:"per_page"`
	Total       int         `json:"total"`
	Data        []ImageItem `json:"data"`
}

type apiEnvelope[T any] struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type ListImagesParams struct {
	Page       int
	Order      string
	AlbumID    int
	Permission string
}

func (c *Client) ListImages(ctx context.Context, params ListImagesParams) (*ImagesPage, error) {
	if !c.Enabled() {
		return &ImagesPage{}, nil
	}

	query := url.Values{}
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}
	if order := strings.TrimSpace(params.Order); order != "" {
		query.Set("order", order)
	}
	if permission := strings.TrimSpace(params.Permission); permission != "" {
		query.Set("permission", permission)
	}
	if params.AlbumID > 0 {
		query.Set("album_id", strconv.Itoa(params.AlbumID))
	}

	endpoint := c.baseURL + "/images"
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("image bed http %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope apiEnvelope[ImagesPage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if !envelope.Status {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = "image bed request failed"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return &envelope.Data, nil
}
