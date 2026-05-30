package dao

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"blog-server/internal/ent"
	"blog-server/internal/ent/game"
	"blog-server/internal/ent/gamemonthlystat"
	"blog-server/internal/ent/gameplaytimesnapshot"
	"blog-server/internal/model"
)

var ErrGameNotFound = errors.New("game not found")

type GameListFilter struct {
	Page        int
	PageSize    int
	GenreSlug   string
	Status      string
	Sort        string
	OnlyVisible bool
}

type GameDAO interface {
	List(ctx context.Context, filter GameListFilter) ([]model.Game, int, error)
	FindByID(ctx context.Context, id uint64) (*model.Game, error)
	FindBySteamAppID(ctx context.Context, steamAppID uint32) (*model.Game, error)
	UpsertFromSteam(ctx context.Context, item model.Game) (*model.Game, error)
	Update(ctx context.Context, id uint64, item model.Game, withGenres bool) (*model.Game, error)
	CountVisible(ctx context.Context) (int, error)
	ListAllVisible(ctx context.Context) ([]model.Game, error)
	ListRecentVisible(ctx context.Context, limit int) ([]model.Game, error)
	CreateSnapshot(ctx context.Context, steamAppID uint32, playtimeMinutes uint32, snapshotAt time.Time) error
	LatestSnapshotMinutes(ctx context.Context, steamAppID uint32) (uint32, bool, error)
	AddMonthlyMinutes(ctx context.Context, yearMonth string, deltaMinutes uint32) error
	ListMonthlyStats(ctx context.Context, limit int) ([]model.GameMonthlyStat, error)
	SumVisiblePlaytimeMinutes(ctx context.Context) (uint32, error)
}

type EntGameDAO struct {
	client *ent.Client
}

func NewEntGameDAO(client *ent.Client) *EntGameDAO {
	return &EntGameDAO{client: client}
}

func (dao *EntGameDAO) List(ctx context.Context, filter GameListFilter) ([]model.Game, int, error) {
	query := dao.client.Game.Query()
	if filter.OnlyVisible {
		query = query.Where(game.IsVisibleEQ(true))
	}
	if filter.Status == "playing" {
		query = query.Where(game.PlayStatusEQ("playing"))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.Game, 0, len(rows))
	for _, row := range rows {
		item := toGameModel(row)
		if filter.GenreSlug != "" && filter.GenreSlug != "all" {
			if !gameHasGenreSlug(item.Genres, filter.GenreSlug) {
				continue
			}
		}
		items = append(items, item)
	}

	sortGames(items, filter.Sort)

	total := len(items)
	start := (filter.Page - 1) * filter.PageSize
	if start > total {
		return []model.Game{}, total, nil
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (dao *EntGameDAO) FindByID(ctx context.Context, id uint64) (*model.Game, error) {
	row, err := dao.client.Game.Query().
		Where(game.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}
	item := toGameModel(row)
	return &item, nil
}

func (dao *EntGameDAO) FindBySteamAppID(ctx context.Context, steamAppID uint32) (*model.Game, error) {
	row, err := dao.client.Game.Query().
		Where(game.SteamAppIdEQ(steamAppID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}
	item := toGameModel(row)
	return &item, nil
}

func (dao *EntGameDAO) UpsertFromSteam(ctx context.Context, item model.Game) (*model.Game, error) {
	existing, err := dao.FindBySteamAppID(ctx, item.SteamAppID)
	now := time.Now()
	if errors.Is(err, ErrGameNotFound) {
		row, createErr := dao.client.Game.Create().
			SetSteamAppId(item.SteamAppID).
			SetName(item.Name).
			SetNillableNameZh(item.NameZh).
			SetCover(item.Cover).
			SetGenres(item.Genres).
			SetPlaytimeMinutes(item.PlaytimeMinutes).
			SetPlaytime2WeeksMinutes(item.Playtime2WeeksMinutes).
			SetNillableLastPlayedAt(item.LastPlayedAt).
			SetNillableAchievementUnlocked(item.AchievementUnlocked).
			SetNillableAchievementTotal(item.AchievementTotal).
			SetNillableProgressPercent(item.ProgressPercent).
			SetProgressSource(item.ProgressSource).
			SetPlayStatus(item.PlayStatus).
			SetIsVisible(false).
			SetSort(item.Sort).
			SetNillableSyncedAt(&now).
			Save(ctx)
		if createErr != nil {
			return nil, createErr
		}
		result := toGameModel(row)
		return &result, nil
	}
	if err != nil {
		return nil, err
	}

	update := dao.client.Game.UpdateOneID(existing.ID).
		SetName(item.Name).
		SetCover(item.Cover).
		SetGenres(item.Genres).
		SetPlaytimeMinutes(item.PlaytimeMinutes).
		SetPlaytime2WeeksMinutes(item.Playtime2WeeksMinutes).
		SetNillableLastPlayedAt(item.LastPlayedAt).
		SetNillableSyncedAt(&now)

	if item.NameZh != nil {
		update = update.SetNillableNameZh(item.NameZh)
	}
	if item.AchievementUnlocked != nil {
		update = update.SetNillableAchievementUnlocked(item.AchievementUnlocked)
	}
	if item.AchievementTotal != nil {
		update = update.SetNillableAchievementTotal(item.AchievementTotal)
	}
	if item.ProgressPercent != nil {
		update = update.SetNillableProgressPercent(item.ProgressPercent).
			SetProgressSource(item.ProgressSource)
	} else if existing.ProgressSource != "manual" {
		update = update.SetNillableProgressPercent(nil).SetProgressSource("none")
	}

	row, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	result := toGameModel(row)
	return &result, nil
}

func (dao *EntGameDAO) Update(ctx context.Context, id uint64, item model.Game, withGenres bool) (*model.Game, error) {
	update := dao.client.Game.UpdateOneID(id)
	if item.NameZh != nil {
		update = update.SetNillableNameZh(item.NameZh)
	}
	if item.PlayStatus != "" {
		update = update.SetPlayStatus(item.PlayStatus)
	}
	if item.ProgressPercent != nil {
		update = update.SetNillableProgressPercent(item.ProgressPercent).
			SetProgressSource(item.ProgressSource)
	}
	if withGenres {
		update = update.SetGenres(item.Genres)
	}
	update = update.SetIsVisible(item.IsVisible).SetSort(item.Sort)

	row, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}
	result := toGameModel(row)
	return &result, nil
}

func (dao *EntGameDAO) CountVisible(ctx context.Context) (int, error) {
	return dao.client.Game.Query().
		Where(game.IsVisibleEQ(true)).
		Count(ctx)
}

func (dao *EntGameDAO) ListAllVisible(ctx context.Context) ([]model.Game, error) {
	rows, err := dao.client.Game.Query().
		Where(game.IsVisibleEQ(true)).
		Order(ent.Asc(game.FieldSort), ent.Desc(game.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]model.Game, 0, len(rows))
	for _, row := range rows {
		items = append(items, toGameModel(row))
	}
	return items, nil
}

func (dao *EntGameDAO) ListRecentVisible(ctx context.Context, limit int) ([]model.Game, error) {
	rows, err := dao.client.Game.Query().
		Where(
			game.IsVisibleEQ(true),
			game.LastPlayedAtNotNil(),
		).
		Order(ent.Desc(game.FieldLastPlayedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]model.Game, 0, len(rows))
	for _, row := range rows {
		items = append(items, toGameModel(row))
	}
	return items, nil
}

func (dao *EntGameDAO) CreateSnapshot(ctx context.Context, steamAppID uint32, playtimeMinutes uint32, snapshotAt time.Time) error {
	_, err := dao.client.GamePlaytimeSnapshot.Create().
		SetSteamAppId(steamAppID).
		SetPlaytimeMinutes(playtimeMinutes).
		SetSnapshotAt(snapshotAt).
		Save(ctx)
	return err
}

func (dao *EntGameDAO) LatestSnapshotMinutes(ctx context.Context, steamAppID uint32) (uint32, bool, error) {
	row, err := dao.client.GamePlaytimeSnapshot.Query().
		Where(gameplaytimesnapshot.SteamAppIdEQ(steamAppID)).
		Order(ent.Desc(gameplaytimesnapshot.FieldSnapshotAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return row.PlaytimeMinutes, true, nil
}

func (dao *EntGameDAO) AddMonthlyMinutes(ctx context.Context, yearMonth string, deltaMinutes uint32) error {
	if deltaMinutes == 0 {
		return nil
	}
	existing, err := dao.client.GameMonthlyStat.Query().
		Where(gamemonthlystat.YearMonthEQ(yearMonth)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			_, createErr := dao.client.GameMonthlyStat.Create().
				SetYearMonth(yearMonth).
				SetTotalMinutes(deltaMinutes).
				Save(ctx)
			return createErr
		}
		return err
	}
	return dao.client.GameMonthlyStat.UpdateOneID(existing.ID).
		SetTotalMinutes(existing.TotalMinutes + deltaMinutes).
		Exec(ctx)
}

func (dao *EntGameDAO) ListMonthlyStats(ctx context.Context, limit int) ([]model.GameMonthlyStat, error) {
	rows, err := dao.client.GameMonthlyStat.Query().
		Order(ent.Desc(gamemonthlystat.FieldYearMonth)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]model.GameMonthlyStat, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.GameMonthlyStat{
			YearMonth:    row.YearMonth,
			TotalMinutes: row.TotalMinutes,
		})
	}
	return items, nil
}

func (dao *EntGameDAO) SumVisiblePlaytimeMinutes(ctx context.Context) (uint32, error) {
	rows, err := dao.client.Game.Query().
		Where(game.IsVisibleEQ(true)).
		All(ctx)
	if err != nil {
		return 0, err
	}
	var total uint32
	for _, row := range rows {
		total += row.PlaytimeMinutes
	}
	return total, nil
}

func cloneGenres(genres []string) []string {
	if len(genres) == 0 {
		return []string{}
	}
	return append([]string(nil), genres...)
}

func toGameModel(row *ent.Game) model.Game {
	return model.Game{
		ID:                    row.ID,
		SteamAppID:            row.SteamAppId,
		Name:                  row.Name,
		NameZh:                row.NameZh,
		Cover:                 row.Cover,
		Genres:                cloneGenres(row.Genres),
		PlaytimeMinutes:       row.PlaytimeMinutes,
		Playtime2WeeksMinutes: row.Playtime2WeeksMinutes,
		LastPlayedAt:          row.LastPlayedAt,
		AchievementUnlocked:   row.AchievementUnlocked,
		AchievementTotal:      row.AchievementTotal,
		ProgressPercent:       row.ProgressPercent,
		ProgressSource:        row.ProgressSource,
		PlayStatus:            row.PlayStatus,
		IsVisible:             row.IsVisible,
		Sort:                  row.Sort,
		SyncedAt:              row.SyncedAt,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func gameHasGenreSlug(genres []string, slug string) bool {
	for _, genre := range genres {
		if strings.TrimSpace(genre) == slug {
			return true
		}
	}
	return false
}

func GenreSlug(name string) string {
	value := strings.TrimSpace(name)
	if value == "" {
		return ""
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			if builder.Len() > 0 && !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
			builder.WriteRune(r + ('a' - 'A'))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			builder.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '_' || r == '-' || r == '/' || r == '&':
			if builder.Len() > 0 && !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

func sortGames(items []model.Game, sortKey string) {
	switch sortKey {
	case "playtime":
		sort.Slice(items, func(i, j int) bool {
			if items[i].PlaytimeMinutes == items[j].PlaytimeMinutes {
				return items[i].ID < items[j].ID
			}
			return items[i].PlaytimeMinutes > items[j].PlaytimeMinutes
		})
	case "name":
		sort.Slice(items, func(i, j int) bool {
			left := gameDisplayName(items[i])
			right := gameDisplayName(items[j])
			if left == right {
				return items[i].ID < items[j].ID
			}
			return left < right
		})
	default:
		sort.Slice(items, func(i, j int) bool {
			left := items[i].LastPlayedAt
			right := items[j].LastPlayedAt
			if left == nil && right == nil {
				return items[i].ID > items[j].ID
			}
			if left == nil {
				return false
			}
			if right == nil {
				return true
			}
			if left.Equal(*right) {
				return items[i].ID > items[j].ID
			}
			return left.After(*right)
		})
	}
}

func gameDisplayName(item model.Game) string {
	if item.NameZh != nil && strings.TrimSpace(*item.NameZh) != "" {
		return *item.NameZh
	}
	return item.Name
}
