package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"
)

var ErrMusicTrackNotFound = errors.New("music track not found")

type MusicService struct {
	tracks dao.MusicTrackDAO
	auth   *AuthService
}

func NewMusicService(tracks dao.MusicTrackDAO, auth *AuthService) *MusicService {
	return &MusicService{
		tracks: tracks,
		auth:   auth,
	}
}

func (svc *MusicService) WebPlaylist(ctx context.Context) (*dto.WebMusicPlaylistResponse, error) {
	items, err := svc.tracks.ListVisible(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.WebMusicTrackItem, 0, len(items))
	for _, item := range items {
		result = append(result, toWebMusicTrackItem(&item))
	}
	return &dto.WebMusicPlaylistResponse{Items: result}, nil
}

func (svc *MusicService) AdminList(
	ctx context.Context,
	authorization string,
	page int,
	pageSize int,
	keyword string,
	visible *bool,
) (*dto.PageResult[dto.AdminMusicTrackListItem], error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	items, total, err := svc.tracks.List(ctx, model.MusicTrackListFilter{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
		Visible:  visible,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminMusicTrackListItem, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminMusicTrackListItem(&item))
	}
	return &dto.PageResult[dto.AdminMusicTrackListItem]{
		Items: result,
		Total: total,
	}, nil
}

func (svc *MusicService) AdminGetByID(
	ctx context.Context,
	authorization string,
	id uint64,
) (*dto.AdminMusicTrackDetail, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.tracks.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrMusicTrackNotFound) {
			return nil, ErrMusicTrackNotFound
		}
		return nil, err
	}
	return toAdminMusicTrackDetail(item), nil
}

func (svc *MusicService) Create(
	ctx context.Context,
	authorization string,
	req dto.MusicTrackRequest,
) (*dto.AdminMusicTrackDetail, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	input := normalizeMusicTrackRequest(req)
	item, err := svc.tracks.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	return toAdminMusicTrackDetail(item), nil
}

func (svc *MusicService) Update(
	ctx context.Context,
	authorization string,
	id uint64,
	req dto.MusicTrackRequest,
) (*dto.AdminMusicTrackDetail, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	input := normalizeMusicTrackRequest(req)
	item, err := svc.tracks.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, dao.ErrMusicTrackNotFound) {
			return nil, ErrMusicTrackNotFound
		}
		return nil, err
	}
	return toAdminMusicTrackDetail(item), nil
}

func (svc *MusicService) Delete(ctx context.Context, authorization string, id uint64) error {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return err
	}

	err := svc.tracks.Delete(ctx, id)
	if errors.Is(err, dao.ErrMusicTrackNotFound) {
		return ErrMusicTrackNotFound
	}
	return err
}

func normalizeMusicTrackRequest(req dto.MusicTrackRequest) model.MusicTrackSaveInput {
	sort := req.Sort
	if sort <= 0 {
		sort = 100
	}
	return model.MusicTrackSaveInput{
		Name:            strings.TrimSpace(req.Name),
		Artist:          strings.TrimSpace(req.Artist),
		AudioURL:        strings.TrimSpace(req.AudioURL),
		CoverURL:        strings.TrimSpace(req.CoverURL),
		Lrc:             strings.TrimSpace(req.Lrc),
		DurationSeconds: max(0, req.DurationSeconds),
		Sort:            sort,
		Visible:         req.Visible,
	}
}

func toAdminMusicTrackListItem(item *model.MusicTrack) dto.AdminMusicTrackListItem {
	return dto.AdminMusicTrackListItem{
		ID:              item.ID,
		Name:            item.Name,
		Artist:          item.Artist,
		AudioURL:        item.AudioURL,
		CoverURL:        item.CoverURL,
		DurationSeconds: item.DurationSeconds,
		Sort:            item.Sort,
		Visible:         item.Visible,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func toAdminMusicTrackDetail(item *model.MusicTrack) *dto.AdminMusicTrackDetail {
	return &dto.AdminMusicTrackDetail{
		AdminMusicTrackListItem: toAdminMusicTrackListItem(item),
		Lrc:                     item.Lrc,
	}
}

func toWebMusicTrackItem(item *model.MusicTrack) dto.WebMusicTrackItem {
	web := dto.WebMusicTrackItem{
		ID:     strconv.FormatUint(item.ID, 10),
		Name:   item.Name,
		Artist: item.Artist,
		URL:    item.AudioURL,
	}
	if item.CoverURL != "" {
		web.Cover = item.CoverURL
	}
	if item.Lrc != "" {
		web.Lrc = item.Lrc
	}
	return web
}
