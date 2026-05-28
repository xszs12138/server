package service

import (
	"context"
	"encoding/json"
	"errors"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
)

type SiteService struct {
	settings dao.SiteSettingDAO
}

func NewSiteService(settings dao.SiteSettingDAO) *SiteService {
	return &SiteService{settings: settings}
}

func (svc *SiteService) WebGetSite(ctx context.Context) (*dto.WebSiteInfo, error) {
	profileRaw, profileErr := svc.settings.GetValueByKey(ctx, "site.profile")
	socialRaw, socialErr := svc.settings.GetValueByKey(ctx, "site.socialLinks")

	profile := dto.SiteProfileValue{
		Name:        "xszs-blog",
		Description: "个人博客",
	}
	if profileErr == nil {
		if err := json.Unmarshal(profileRaw, &profile); err != nil {
			return nil, err
		}
	} else if !errors.Is(profileErr, dao.ErrSiteSettingNotFound) {
		return nil, profileErr
	}

	socialLinks := []dto.WebSocialLink{}
	if socialErr == nil {
		if err := json.Unmarshal(socialRaw, &socialLinks); err != nil {
			return nil, err
		}
	} else if !errors.Is(socialErr, dao.ErrSiteSettingNotFound) {
		return nil, socialErr
	}

	return &dto.WebSiteInfo{
		ID:          1,
		Name:        profile.Name,
		Description: profile.Description,
		Logo:        profile.Logo,
		Icp:         profile.Icp,
		SocialLinks: socialLinks,
	}, nil
}
