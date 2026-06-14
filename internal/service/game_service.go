package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"blog-server/internal/config"
	"blog-server/internal/dao"
	"blog-server/internal/dicttypes"
	"blog-server/internal/dto"
	"blog-server/internal/model"
	"blog-server/internal/steam"
)

var (
	ErrGameNotFound        = errors.New("game not found")
	ErrSteamNotConfigured  = errors.New("steam not configured")
	ErrSteamLibraryPrivate = steam.ErrGameLibraryPrivate
	ErrSteamInvalidAPIKey  = steam.ErrInvalidAPIKey
	ErrInvalidPlayStatus   = errors.New("invalid play status")
)

var allowedPlayStatuses = map[string]struct{}{
	"playing":   {},
	"completed": {},
	"backlog":   {},
	"dropped":   {},
	"hidden":    {},
}

type GameService struct {
	games     dao.GameDAO
	dictItems dao.DictItemDAO
	auth      *AuthService
	steam     *steam.Client
}

func NewGameService(
	games dao.GameDAO,
	dictItems dao.DictItemDAO,
	auth *AuthService,
	cfg config.Config,
) *GameService {
	return &GameService{
		games:     games,
		dictItems: dictItems,
		auth:      auth,
		steam:     steam.NewClient(cfg.SteamAPIKey, cfg.SteamID, cfg.SteamSyncLang),
	}
}

func (svc *GameService) WebList(
	ctx context.Context,
	page int,
	pageSize int,
	genreSlug string,
	status string,
	sortKey string,
) (*dto.PageResult[dto.WebGameListItem], error) {
	if sortKey == "" {
		sortKey = "recent"
	}
	items, total, err := svc.games.List(ctx, dao.GameListFilter{
		Page:        page,
		PageSize:    pageSize,
		GenreSlug:   genreSlug,
		Status:      status,
		Sort:        sortKey,
		OnlyVisible: true,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.WebGameListItem, 0, len(items))
	for _, item := range items {
		result = append(result, toWebGameListItem(item))
	}
	return &dto.PageResult[dto.WebGameListItem]{
		Items: result,
		Total: total,
	}, nil
}

func (svc *GameService) WebListGenres(ctx context.Context) ([]dto.WebGameGenreItem, error) {
	games, err := svc.games.ListAllVisible(ctx)
	if err != nil {
		return nil, err
	}

	dictItems, err := svc.dictItems.ListEnabled(ctx, dicttypes.GameGenre)
	if err != nil {
		return nil, err
	}

	counter := map[string]int{}
	for _, item := range games {
		for _, code := range item.Genres {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			counter[code]++
		}
	}

	result := []dto.WebGameGenreItem{
		{Slug: "all", Name: "全部", Count: len(games)},
	}
	for _, item := range dictItems {
		code := ""
		if item.Code != nil {
			code = strings.TrimSpace(*item.Code)
		}
		if code == "" {
			continue
		}
		result = append(result, dto.WebGameGenreItem{
			Slug:  code,
			Name:  item.Label,
			Count: counter[code],
		})
	}
	return result, nil
}

func (svc *GameService) WebSidebar(ctx context.Context) (*dto.WebGameSidebar, error) {
	recent, err := svc.games.ListRecentVisible(ctx, 5)
	if err != nil {
		return nil, err
	}
	recentItems := make([]dto.WebGameRecentItem, 0, len(recent))
	for _, item := range recent {
		recentItems = append(recentItems, dto.WebGameRecentItem{
			ID:              item.ID,
			SteamAppID:      item.SteamAppID,
			Name:            item.Name,
			NameZh:          gameDisplayName(item),
			Cover:           item.Cover,
			PlaytimeHours:   minutesToHours(item.PlaytimeMinutes),
			ProgressPercent: item.ProgressPercent,
		})
	}

	statsRows, err := svc.games.ListMonthlyStats(ctx, 6)
	if err != nil {
		return nil, err
	}
	months := make([]dto.WebGamePlaytimeMonth, 0, len(statsRows))
	var totalMinutes uint32
	for i := len(statsRows) - 1; i >= 0; i-- {
		row := statsRows[i]
		totalMinutes += row.TotalMinutes
		months = append(months, dto.WebGamePlaytimeMonth{
			Month: row.YearMonth,
			Hours: minutesToHours(row.TotalMinutes),
		})
	}

	visibleTotal, err := svc.games.SumVisiblePlaytimeMinutes(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.WebGameSidebar{
		RecentGames: recentItems,
		PlaytimeStats: dto.WebGamePlaytimeStats{
			TotalHours: minutesToHours(visibleTotal),
			Months:     months,
		},
	}, nil
}

func (svc *GameService) AdminList(
	ctx context.Context,
	authorization string,
	page int,
	pageSize int,
) (*dto.PageResult[dto.AdminGameListItem], error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	items, total, err := svc.games.List(ctx, dao.GameListFilter{
		Page:        page,
		PageSize:    pageSize,
		OnlyVisible: false,
		Sort:        "recent",
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminGameListItem, 0, len(items))
	for _, item := range items {
		webItem := toWebGameListItem(item)
		result = append(result, dto.AdminGameListItem{
			WebGameListItem:     webItem,
			ProgressSource:      item.ProgressSource,
			IsVisible:           item.IsVisible,
			PlaytimeMinutes:     item.PlaytimeMinutes,
			AchievementUnlocked: item.AchievementUnlocked,
			AchievementTotal:    item.AchievementTotal,
			SyncedAt:            item.SyncedAt,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
		})
	}
	return &dto.PageResult[dto.AdminGameListItem]{
		Items: result,
		Total: total,
	}, nil
}

func (svc *GameService) AdminGetByID(
	ctx context.Context,
	authorization string,
	id uint64,
) (*dto.AdminGameListItem, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.games.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrGameNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	webItem := toWebGameListItem(*item)
	result := &dto.AdminGameListItem{
		WebGameListItem:     webItem,
		ProgressSource:      item.ProgressSource,
		IsVisible:           item.IsVisible,
		PlaytimeMinutes:     item.PlaytimeMinutes,
		AchievementUnlocked: item.AchievementUnlocked,
		AchievementTotal:    item.AchievementTotal,
		SyncedAt:            item.SyncedAt,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
	return result, nil
}

func (svc *GameService) AdminUpdate(
	ctx context.Context,
	authorization string,
	id uint64,
	req dto.AdminGameUpdateRequest,
) (*dto.AdminGameListItem, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	existing, err := svc.games.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrGameNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	update := model.Game{
		PlayStatus: existing.PlayStatus,
		IsVisible:  existing.IsVisible,
		Sort:       existing.Sort,
	}
	if req.NameZh != nil {
		value := *req.NameZh
		update.NameZh = &value
	}
	if req.PlayStatus != nil {
		if _, ok := allowedPlayStatuses[*req.PlayStatus]; !ok {
			return nil, ErrInvalidPlayStatus
		}
		update.PlayStatus = *req.PlayStatus
	}
	if req.ProgressPercent != nil {
		if *req.ProgressPercent > 100 {
			return nil, ErrInvalidPlayStatus
		}
		update.ProgressPercent = req.ProgressPercent
		update.ProgressSource = "manual"
	}
	if req.IsVisible != nil {
		update.IsVisible = *req.IsVisible
	}
	if req.Sort != nil {
		update.Sort = *req.Sort
	}
	if req.Genres != nil {
		validCodes, err := svc.enabledGameGenreCodes(ctx)
		if err != nil {
			return nil, err
		}
		normalized := make([]string, 0, len(req.Genres))
		for _, code := range req.Genres {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if !validCodes[code] {
				return nil, fmt.Errorf("无效的游戏类型: %s", code)
			}
			normalized = append(normalized, code)
		}
		update.Genres = normalized
	}

	updated, err := svc.games.Update(ctx, id, update, req.Genres != nil)
	if err != nil {
		return nil, err
	}
	return svc.AdminGetByID(ctx, authorization, updated.ID)
}

func (svc *GameService) AdminSync(
	ctx context.Context,
	authorization string,
) (*dto.GameSyncResult, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}
	if !svc.steam.Configured() {
		log.Printf("[steam] sync aborted: apiKey or steamId not configured")
		return nil, ErrSteamNotConfigured
	}

	log.Printf("[steam] sync start")
	owned, err := svc.steam.GetOwnedGames(ctx)
	if err != nil {
		log.Printf("[steam] sync GetOwnedGames failed: %v", err)
		return nil, err
	}
	log.Printf("[steam] sync owned games count=%d", len(owned))

	recentMap := map[uint32]time.Time{}
	if recent, recentErr := svc.steam.GetRecentlyPlayedGames(ctx, 10); recentErr == nil {
		now := time.Now()
		for index, item := range recent {
			recentMap[item.AppID] = now.Add(-time.Duration(index) * time.Minute)
		}
		log.Printf("[steam] sync recently played count=%d", len(recent))
	} else {
		log.Printf("[steam] sync GetRecentlyPlayedGames skipped: %v", recentErr)
	}

	syncedAt := time.Now()
	yearMonth := syncedAt.Format("2006-01")
	syncedCount := 0

	for _, ownedGame := range owned {
		var lastPlayed *time.Time
		if playedAt, ok := recentMap[ownedGame.AppID]; ok {
			lastPlayed = &playedAt
		}

		existing, _ := svc.games.FindBySteamAppID(ctx, ownedGame.AppID)
		shouldFetchAchievements := ownedGame.Playtime2Weeks > 0
		if existing != nil && existing.IsVisible {
			shouldFetchAchievements = true
		}
		var progressPercent *uint8
		var progressSource string
		var unlocked, total *uint32
		if shouldFetchAchievements {
			progressPercent, progressSource, unlocked, total = svc.resolveProgress(
				ctx,
				ownedGame.AppID,
				ownedGame.HasCommunityVisibleStats,
			)
		} else {
			progressSource = "none"
		}

		displayName := strings.TrimSpace(ownedGame.Name)
		if displayName == "" {
			displayName = fmt.Sprintf("Steam App %d", ownedGame.AppID)
		}

		var nameZhPtr *string
		if existing != nil && existing.NameZh != nil && strings.TrimSpace(*existing.NameZh) != "" {
			nameZhPtr = existing.NameZh
		}

		genres := []string{}
		if existing != nil && len(existing.Genres) > 0 {
			genres = existing.Genres
		}

		// 仅对博客可见游戏补充商店 genres/中文名，避免全库逐条请求商店接口
		if existing != nil && existing.IsVisible {
			if details, detailsErr := svc.steam.GetAppDetails(ctx, ownedGame.AppID); detailsErr == nil {
				if details.NameZh != "" {
					zh := details.NameZh
					nameZhPtr = &zh
				}
				if len(details.Genres) > 0 {
					genres = details.Genres
				}
			} else {
				log.Printf("[steam] sync GetAppDetails appid=%d failed: %v", ownedGame.AppID, detailsErr)
			}
		}

		item := model.Game{
			SteamAppID:            ownedGame.AppID,
			Name:                  displayName,
			NameZh:                nameZhPtr,
			Cover:                 steam.BuildCoverURL(ownedGame.AppID, ""),
			Genres:                genres,
			PlaytimeMinutes:       ownedGame.PlaytimeForever,
			Playtime2WeeksMinutes: ownedGame.Playtime2Weeks,
			LastPlayedAt:          lastPlayed,
			AchievementUnlocked:   unlocked,
			AchievementTotal:      total,
			ProgressPercent:       progressPercent,
			ProgressSource:        progressSource,
			PlayStatus:            "backlog",
			Sort:                  100,
		}

		if _, err := svc.games.UpsertFromSteam(ctx, item); err != nil {
			log.Printf("[steam] sync upsert failed appid=%d name=%q err=%v", ownedGame.AppID, displayName, err)
			return nil, fmt.Errorf("保存游戏 %s (appid=%d) 失败: %w", displayName, ownedGame.AppID, err)
		}
		syncedCount++

		if previous, ok, snapErr := svc.games.LatestSnapshotMinutes(ctx, ownedGame.AppID); snapErr == nil && ok {
			if ownedGame.PlaytimeForever > previous {
				delta := ownedGame.PlaytimeForever - previous
				if addErr := svc.games.AddMonthlyMinutes(ctx, yearMonth, delta); addErr != nil {
					log.Printf("[steam] sync monthly stats failed appid=%d err=%v", ownedGame.AppID, addErr)
					return nil, fmt.Errorf("更新月度统计失败 (appid=%d): %w", ownedGame.AppID, addErr)
				}
			}
		}
		if snapErr := svc.games.CreateSnapshot(ctx, ownedGame.AppID, ownedGame.PlaytimeForever, syncedAt); snapErr != nil {
			log.Printf("[steam] sync snapshot failed appid=%d err=%v", ownedGame.AppID, snapErr)
			return nil, fmt.Errorf("写入时长快照失败 (appid=%d): %w", ownedGame.AppID, snapErr)
		}
	}

	visibleCount, err := svc.games.CountVisible(ctx)
	if err != nil {
		log.Printf("[steam] sync CountVisible failed: %v", err)
		return nil, err
	}

	log.Printf("[steam] sync done synced=%d visible=%d", syncedCount, visibleCount)
	return &dto.GameSyncResult{
		SyncedCount:  syncedCount,
		VisibleCount: visibleCount,
		SyncedAt:     syncedAt,
	}, nil
}

func (svc *GameService) resolveProgress(
	ctx context.Context,
	appID uint32,
	hasCommunityVisible bool,
) (*uint8, string, *uint32, *uint32) {
	if !hasCommunityVisible {
		return nil, "none", nil, nil
	}
	stats, err := svc.steam.GetAchievementStats(ctx, appID)
	if err != nil || stats == nil || stats.Total == 0 {
		return nil, "none", nil, nil
	}
	percent := uint8(math.Round(float64(stats.Unlocked) / float64(stats.Total) * 100))
	unlocked := stats.Unlocked
	total := stats.Total
	return &percent, "achievement", &unlocked, &total
}

func toWebGameListItem(item model.Game) dto.WebGameListItem {
	label := ""
	if item.PlayStatus == "playing" {
		label = "正在游玩"
	}
	genres := item.Genres
	if genres == nil {
		genres = []string{}
	}
	return dto.WebGameListItem{
		ID:                  item.ID,
		SteamAppID:          item.SteamAppID,
		Name:                item.Name,
		NameZh:              gameDisplayName(item),
		Cover:               item.Cover,
		Genres:              genres,
		PlayStatus:          item.PlayStatus,
		PlayStatusLabel:     label,
		ProgressPercent:     item.ProgressPercent,
		PlaytimeHours:       minutesToHours(item.PlaytimeMinutes),
		Playtime2WeeksHours: minutesToHours(item.Playtime2WeeksMinutes),
		LastPlayedAt:        item.LastPlayedAt,
	}
}

func gameDisplayName(item model.Game) string {
	if item.NameZh != nil && *item.NameZh != "" {
		return *item.NameZh
	}
	return item.Name
}

func minutesToHours(minutes uint32) float64 {
	return math.Round(float64(minutes)/60*10) / 10
}

func (svc *GameService) enabledGameGenreCodes(ctx context.Context) (map[string]bool, error) {
	items, err := svc.dictItems.ListEnabled(ctx, dicttypes.GameGenre)
	if err != nil {
		return nil, err
	}
	codes := make(map[string]bool, len(items))
	for _, item := range items {
		if item.Code == nil {
			continue
		}
		code := strings.TrimSpace(*item.Code)
		if code != "" {
			codes[code] = true
		}
	}
	return codes, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
