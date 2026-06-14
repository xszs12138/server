package steam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidAPIKey = errors.New("Steam API Key 无效")
	// ErrGameLibraryPrivate 资料或游戏详情未公开时，GetOwnedGames 会返回空 response。
	ErrGameLibraryPrivate = errors.New(
		"无法读取 Steam 游戏库：请在 Steam 隐私设置中将「游戏详情」设为公开（个人资料公开不够）",
	)
)

const (
	ownedGamesURL          = "https://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/"
	recentlyPlayedURL      = "https://api.steampowered.com/IPlayerService/GetRecentlyPlayedGames/v0001/"
	playerAchievementsURL  = "https://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v1/"
	schemaForGameURL       = "https://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2/"
	storeAppDetailsURL     = "https://store.steampowered.com/api/appdetails"
)

type Client struct {
	apiKey     string
	steamID    string
	lang       string
	httpClient *http.Client
}

func NewClient(apiKey, steamID, lang string) *Client {
	if lang == "" {
		lang = "schinese"
	}
	if proxy := strings.TrimSpace(os.Getenv("HTTPS_PROXY")); proxy != "" {
		log.Printf("[steam] client init: HTTPS_PROXY=%s steamId=%s apiKeyConfigured=%t", proxy, maskSteamID(steamID), apiKey != "")
	} else if proxy := strings.TrimSpace(os.Getenv("HTTP_PROXY")); proxy != "" {
		log.Printf("[steam] client init: HTTP_PROXY=%s steamId=%s apiKeyConfigured=%t", proxy, maskSteamID(steamID), apiKey != "")
	} else {
		log.Printf("[steam] client init: no HTTP(S)_PROXY steamId=%s apiKeyConfigured=%t", maskSteamID(steamID), apiKey != "")
	}
	return &Client{
		apiKey:  apiKey,
		steamID: steamID,
		lang:    lang,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

func (c *Client) Configured() bool {
	return c.apiKey != "" && c.steamID != ""
}

type OwnedGame struct {
	AppID                uint32 `json:"appid"`
	Name                 string `json:"name"`
	PlaytimeForever      uint32 `json:"playtime_forever"`
	Playtime2Weeks       uint32 `json:"playtime_2weeks"`
	ImgIconURL           string `json:"img_icon_url"`
	HasCommunityVisibleStats bool `json:"has_community_visible_stats"`
}

type RecentGame struct {
	AppID           uint32 `json:"appid"`
	Name            string `json:"name"`
	Playtime2Weeks  uint32 `json:"playtime_2weeks"`
	PlaytimeForever uint32 `json:"playtime_forever"`
}

type AppDetails struct {
	Name         string
	HeaderImage  string
	Genres       []string
	NameZh       string
}

type AchievementStats struct {
	Unlocked uint32
	Total    uint32
}

func (c *Client) GetOwnedGames(ctx context.Context) ([]OwnedGame, error) {
	values := url.Values{}
	values.Set("key", c.apiKey)
	values.Set("steamid", c.steamID)
	values.Set("include_appinfo", "1")
	values.Set("include_played_free_games", "1")
	values.Set("format", "json")

	var payload struct {
		Response struct {
			GameCount int `json:"game_count"`
			Games     []OwnedGame `json:"games"`
			Error     *struct {
				ErrorCode int    `json:"errorcode"`
				ErrorDesc string `json:"errordesc"`
			} `json:"error"`
		} `json:"response"`
	}
	if err := c.getJSON(ctx, ownedGamesURL, values, &payload); err != nil {
		return nil, err
	}
	if payload.Response.Error != nil {
		desc := strings.TrimSpace(payload.Response.Error.ErrorDesc)
		if payload.Response.Error.ErrorCode == 2 || strings.Contains(strings.ToLower(desc), "invalid api key") {
			return nil, ErrInvalidAPIKey
		}
		if desc == "" {
			desc = "Steam API 返回错误"
		}
		return nil, fmt.Errorf("%s (code %d)", desc, payload.Response.Error.ErrorCode)
	}

	games := payload.Response.Games
	if games == nil {
		games = []OwnedGame{}
	}
	if len(games) == 0 && payload.Response.GameCount == 0 {
		return nil, ErrGameLibraryPrivate
	}
	return games, nil
}

func (c *Client) GetRecentlyPlayedGames(ctx context.Context, count int) ([]RecentGame, error) {
	values := url.Values{}
	values.Set("key", c.apiKey)
	values.Set("steamid", c.steamID)
	values.Set("count", strconv.Itoa(count))

	var payload struct {
		Response struct {
			Games []RecentGame `json:"games"`
		} `json:"response"`
	}
	if err := c.getJSON(ctx, recentlyPlayedURL, values, &payload); err != nil {
		return nil, err
	}
	if payload.Response.Games == nil {
		return []RecentGame{}, nil
	}
	return payload.Response.Games, nil
}

func (c *Client) GetAchievementStats(ctx context.Context, appID uint32) (*AchievementStats, error) {
	values := url.Values{}
	values.Set("key", c.apiKey)
	values.Set("steamid", c.steamID)
	values.Set("appid", strconv.FormatUint(uint64(appID), 10))

	var payload struct {
		PlayerStats struct {
			Achievements []struct {
				Achieved int `json:"achieved"`
			} `json:"achievements"`
		} `json:"playerstats"`
	}
	if err := c.getJSON(ctx, playerAchievementsURL, values, &payload); err != nil {
		return nil, err
	}
	if len(payload.PlayerStats.Achievements) == 0 {
		return nil, nil
	}

	unlocked := uint32(0)
	for _, item := range payload.PlayerStats.Achievements {
		if item.Achieved == 1 {
			unlocked++
		}
	}
	total := uint32(len(payload.PlayerStats.Achievements))
	return &AchievementStats{Unlocked: unlocked, Total: total}, nil
}

func (c *Client) GetAppDetails(ctx context.Context, appID uint32) (AppDetails, error) {
	values := url.Values{}
	values.Set("appids", strconv.FormatUint(uint64(appID), 10))
	values.Set("l", c.lang)

	var payload map[string]struct {
		Success bool `json:"success"`
		Data    struct {
			Name        string `json:"name"`
			HeaderImage string `json:"header_image"`
			Genres      []struct {
				Description string `json:"description"`
			} `json:"genres"`
		} `json:"data"`
	}
	if err := c.getJSON(ctx, storeAppDetailsURL, values, &payload); err != nil {
		return AppDetails{}, err
	}
	if payload == nil {
		return AppDetails{}, nil
	}

	key := strconv.FormatUint(uint64(appID), 10)
	item, ok := payload[key]
	if !ok || !item.Success {
		return AppDetails{}, nil
	}

	genres := make([]string, 0, len(item.Data.Genres))
	for _, genre := range item.Data.Genres {
		if genre.Description != "" {
			genres = append(genres, genre.Description)
		}
	}
	return AppDetails{
		Name:        item.Data.Name,
		HeaderImage: item.Data.HeaderImage,
		Genres:      genres,
		NameZh:      item.Data.Name,
	}, nil
}

// GetAppDetailsBatch 商店接口不支持可靠的逗号批量查询，按 appID 逐个拉取。
func (c *Client) GetAppDetailsBatch(ctx context.Context, appIDs []uint32) (map[uint32]AppDetails, error) {
	result := make(map[uint32]AppDetails, len(appIDs))
	if len(appIDs) == 0 {
		return result, nil
	}

	for _, appID := range appIDs {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		details, err := c.GetAppDetails(ctx, appID)
		if err != nil {
			return result, err
		}
		if details.Name != "" || details.HeaderImage != "" || len(details.Genres) > 0 {
			result[appID] = details
		}
	}
	return result, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, values url.Values, dest any) error {
	reqURL := endpoint
	if len(values) > 0 {
		reqURL = endpoint + "?" + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		log.Printf("[steam] build request failed endpoint=%s err=%v", endpoint, err)
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[steam] request failed endpoint=%s err=%v (timeout/proxy/network?)", endpoint, err)
		return fmt.Errorf("steam request %s: %w", endpoint, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[steam] read body failed endpoint=%s status=%d err=%v", endpoint, res.StatusCode, err)
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		snippet := truncateForLog(string(body), 256)
		log.Printf("[steam] bad status endpoint=%s status=%d body=%q", endpoint, res.StatusCode, snippet)
		return fmt.Errorf("steam api status %d: %s", res.StatusCode, snippet)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		snippet := truncateForLog(string(body), 256)
		log.Printf("[steam] json decode failed endpoint=%s err=%v body=%q", endpoint, err, snippet)
		return fmt.Errorf("steam json decode %s: %w", endpoint, err)
	}
	return nil
}

func maskSteamID(steamID string) string {
	steamID = strings.TrimSpace(steamID)
	if len(steamID) <= 4 {
		return "****"
	}
	return steamID[:4] + "****" + steamID[len(steamID)-2:]
}

func truncateForLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func BuildCoverURL(appID uint32, headerImage string) string {
	if headerImage != "" {
		return headerImage
	}
	return fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/steam/apps/%d/header.jpg", appID)
}
