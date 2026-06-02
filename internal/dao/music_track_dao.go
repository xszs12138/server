package dao

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/ent"
	"blog-server/internal/ent/musictrack"
	"blog-server/internal/model"
)

var ErrMusicTrackNotFound = errors.New("music track not found")

type MusicTrackDAO interface {
	List(ctx context.Context, filter model.MusicTrackListFilter) ([]model.MusicTrack, int, error)
	ListVisible(ctx context.Context) ([]model.MusicTrack, error)
	FindByID(ctx context.Context, id uint64) (*model.MusicTrack, error)
	Create(ctx context.Context, input model.MusicTrackSaveInput) (*model.MusicTrack, error)
	Update(ctx context.Context, id uint64, input model.MusicTrackSaveInput) (*model.MusicTrack, error)
	Delete(ctx context.Context, id uint64) error
}

type EntMusicTrackDAO struct {
	client *ent.Client
}

func NewEntMusicTrackDAO(client *ent.Client) *EntMusicTrackDAO {
	return &EntMusicTrackDAO{client: client}
}

func (dao *EntMusicTrackDAO) List(
	ctx context.Context,
	filter model.MusicTrackListFilter,
) ([]model.MusicTrack, int, error) {
	query := dao.client.MusicTrack.Query().Where(musictrack.DeletedAtIsNil())

	if filter.Keyword != "" {
		query = query.Where(
			musictrack.Or(
				musictrack.NameContains(filter.Keyword),
				musictrack.ArtistContains(filter.Keyword),
			),
		)
	}
	if filter.Visible != nil {
		query = query.Where(musictrack.VisibleEQ(*filter.Visible))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Order(ent.Asc(musictrack.FieldSort), ent.Desc(musictrack.FieldID)).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.MusicTrack, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toMusicTrackModel(row))
	}
	return items, total, nil
}

func (dao *EntMusicTrackDAO) ListVisible(ctx context.Context) ([]model.MusicTrack, error) {
	rows, err := dao.client.MusicTrack.Query().
		Where(
			musictrack.DeletedAtIsNil(),
			musictrack.VisibleEQ(true),
		).
		Order(ent.Asc(musictrack.FieldSort), ent.Desc(musictrack.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]model.MusicTrack, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toMusicTrackModel(row))
	}
	return items, nil
}

func (dao *EntMusicTrackDAO) FindByID(ctx context.Context, id uint64) (*model.MusicTrack, error) {
	row, err := dao.client.MusicTrack.Query().
		Where(
			musictrack.IDEQ(id),
			musictrack.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrMusicTrackNotFound
		}
		return nil, err
	}
	return toMusicTrackModel(row), nil
}

func (dao *EntMusicTrackDAO) Create(
	ctx context.Context,
	input model.MusicTrackSaveInput,
) (*model.MusicTrack, error) {
	row, err := dao.client.MusicTrack.Create().
		SetName(input.Name).
		SetArtist(input.Artist).
		SetAudioUrl(input.AudioURL).
		SetCoverUrl(input.CoverURL).
		SetLrc(input.Lrc).
		SetDurationSeconds(input.DurationSeconds).
		SetSort(input.Sort).
		SetVisible(input.Visible).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toMusicTrackModel(row), nil
}

func (dao *EntMusicTrackDAO) Update(
	ctx context.Context,
	id uint64,
	input model.MusicTrackSaveInput,
) (*model.MusicTrack, error) {
	row, err := dao.client.MusicTrack.UpdateOneID(id).
		Where(musictrack.DeletedAtIsNil()).
		SetName(input.Name).
		SetArtist(input.Artist).
		SetAudioUrl(input.AudioURL).
		SetCoverUrl(input.CoverURL).
		SetLrc(input.Lrc).
		SetDurationSeconds(input.DurationSeconds).
		SetSort(input.Sort).
		SetVisible(input.Visible).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrMusicTrackNotFound
		}
		return nil, err
	}
	return toMusicTrackModel(row), nil
}

func (dao *EntMusicTrackDAO) Delete(ctx context.Context, id uint64) error {
	now := time.Now()
	n, err := dao.client.MusicTrack.Update().
		Where(
			musictrack.IDEQ(id),
			musictrack.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrMusicTrackNotFound
	}
	return nil
}

func toMusicTrackModel(row *ent.MusicTrack) *model.MusicTrack {
	return &model.MusicTrack{
		ID:              row.ID,
		Name:            row.Name,
		Artist:          row.Artist,
		AudioURL:        row.AudioUrl,
		CoverURL:        row.CoverUrl,
		Lrc:             row.Lrc,
		DurationSeconds: row.DurationSeconds,
		Sort:            row.Sort,
		Visible:         row.Visible,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
