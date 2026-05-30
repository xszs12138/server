package service

import (
	"context"
	"encoding/json"
	"errors"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
)

const liveBroadcastSettingKey = "live.broadcast"

var (
	ErrInvalidLivePlatform = errors.New("invalid live platform")
)

var livePlatformLabels = map[string]string{
	dto.LivePlatformBilibili: "哔哩哔哩",
}

type LiveBroadcaster interface {
	BroadcastLive(item *dto.LiveBroadcast)
}

type LiveService struct {
	settings    dao.SiteSettingDAO
	auth        *AuthService
	broadcaster LiveBroadcaster
}

func NewLiveService(
	settings dao.SiteSettingDAO,
	auth *AuthService,
	broadcaster LiveBroadcaster,
) *LiveService {
	return &LiveService{
		settings:    settings,
		auth:        auth,
		broadcaster: broadcaster,
	}
}

func (svc *LiveService) WebGetLive(ctx context.Context) (*dto.LiveBroadcast, error) {
	item, err := svc.readBroadcast(ctx)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (svc *LiveService) AdminGetLive(ctx context.Context, authorization string) (*dto.LiveBroadcast, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}
	return svc.readBroadcast(ctx)
}

func (svc *LiveService) AdminUpdateLive(
	ctx context.Context,
	authorization string,
	req dto.LiveBroadcastUpdateRequest,
) (*dto.LiveBroadcast, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}
	if req.Platform != dto.LivePlatformBilibili {
		return nil, ErrInvalidLivePlatform
	}

	value := dto.LiveBroadcast{
		IsLive:      req.IsLive,
		Platform:    req.Platform,
		RoomTitle:   req.RoomTitle,
		StreamTitle: req.StreamTitle,
		Subtitle:    req.Subtitle,
		Cover:       req.Cover,
		StreamURL:   req.StreamURL,
	}
	value.PlatformLabel = livePlatformLabels[value.Platform]

	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if err := svc.settings.UpsertValueByKey(ctx, liveBroadcastSettingKey, raw, "直播配置"); err != nil {
		return nil, err
	}
	if svc.broadcaster != nil {
		svc.broadcaster.BroadcastLive(&value)
	}
	return &value, nil
}

func (svc *LiveService) readBroadcast(ctx context.Context) (*dto.LiveBroadcast, error) {
	raw, err := svc.settings.GetValueByKey(ctx, liveBroadcastSettingKey)
	if err != nil {
		if errors.Is(err, dao.ErrSiteSettingNotFound) {
			return defaultLiveBroadcast(), nil
		}
		return nil, err
	}

	var item dto.LiveBroadcast
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	if item.Platform == "" {
		item.Platform = dto.LivePlatformBilibili
	}
	item.PlatformLabel = livePlatformLabels[item.Platform]
	return &item, nil
}

func defaultLiveBroadcast() *dto.LiveBroadcast {
	return &dto.LiveBroadcast{
		IsLive:        false,
		Platform:      dto.LivePlatformBilibili,
		PlatformLabel: livePlatformLabels[dto.LivePlatformBilibili],
	}
}
