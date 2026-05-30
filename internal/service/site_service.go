package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
)

var ErrInvalidSiteSettings = errors.New("invalid site settings")

const (
	siteProfileKey     = "site.profile"
	siteSocialLinksKey = "site.socialLinks"
	siteAboutKey       = "site.about"
	maxAboutMarkdown   = 100_000
)

type SiteService struct {
	settings dao.SiteSettingDAO
	auth     *AuthService
}

func NewSiteService(settings dao.SiteSettingDAO, auth *AuthService) *SiteService {
	return &SiteService{
		settings: settings,
		auth:     auth,
	}
}

func (svc *SiteService) WebGetSite(ctx context.Context) (*dto.WebSiteInfo, error) {
	profile, socialLinks, about, err := svc.readSiteConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.WebSiteInfo{
		ID:          1,
		Name:        profile.Name,
		Description: profile.Description,
		Logo:        profile.Logo,
		Icp:         profile.Icp,
		About:       about,
		SocialLinks: socialLinks,
	}, nil
}

func (svc *SiteService) AdminGetSiteSettings(
	ctx context.Context,
	authorization string,
) (*dto.AdminSiteSettings, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	profile, socialLinks, about, err := svc.readSiteConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.AdminSiteSettings{
		Name:        profile.Name,
		Description: profile.Description,
		Logo:        profile.Logo,
		Icp:         profile.Icp,
		About:       about,
		SocialLinks: socialLinks,
	}, nil
}

func (svc *SiteService) AdminUpdateSiteSettings(
	ctx context.Context,
	authorization string,
	req dto.AdminSiteSettingsUpdateRequest,
) (*dto.AdminSiteSettings, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidSiteSettings
	}

	socialLinks, err := normalizeSocialLinks(req.SocialLinks)
	if err != nil {
		return nil, err
	}

	about := req.About
	if len(about) > maxAboutMarkdown {
		return nil, ErrInvalidSiteSettings
	}

	profile := dto.SiteProfileValue{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Logo:        strings.TrimSpace(req.Logo),
		Icp:         strings.TrimSpace(req.Icp),
	}

	profileRaw, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	if err := svc.settings.UpsertValueByKey(ctx, siteProfileKey, profileRaw, "站点基础信息"); err != nil {
		return nil, err
	}

	socialRaw, err := json.Marshal(socialLinks)
	if err != nil {
		return nil, err
	}
	if err := svc.settings.UpsertValueByKey(ctx, siteSocialLinksKey, socialRaw, "社交链接"); err != nil {
		return nil, err
	}

	aboutValue := dto.SiteAboutValue{Markdown: about}
	aboutRaw, err := json.Marshal(aboutValue)
	if err != nil {
		return nil, err
	}
	if err := svc.settings.UpsertValueByKey(ctx, siteAboutKey, aboutRaw, "关于页 Markdown"); err != nil {
		return nil, err
	}

	return &dto.AdminSiteSettings{
		Name:        profile.Name,
		Description: profile.Description,
		Logo:        profile.Logo,
		Icp:         profile.Icp,
		About:       about,
		SocialLinks: socialLinks,
	}, nil
}

func (svc *SiteService) readSiteConfig(ctx context.Context) (dto.SiteProfileValue, []dto.WebSocialLink, string, error) {
	profileRaw, profileErr := svc.settings.GetValueByKey(ctx, siteProfileKey)
	socialRaw, socialErr := svc.settings.GetValueByKey(ctx, siteSocialLinksKey)

	profile := dto.SiteProfileValue{
		Name:        "xszs-blog",
		Description: "个人博客",
	}
	if profileErr == nil {
		if err := json.Unmarshal(profileRaw, &profile); err != nil {
			return dto.SiteProfileValue{}, nil, "", err
		}
	} else if !errors.Is(profileErr, dao.ErrSiteSettingNotFound) {
		return dto.SiteProfileValue{}, nil, "", profileErr
	}

	socialLinks := []dto.WebSocialLink{}
	if socialErr == nil {
		if err := json.Unmarshal(socialRaw, &socialLinks); err != nil {
			return dto.SiteProfileValue{}, nil, "", err
		}
	} else if !errors.Is(socialErr, dao.ErrSiteSettingNotFound) {
		return dto.SiteProfileValue{}, nil, "", socialErr
	}

	about, err := svc.readAboutMarkdown(ctx)
	if err != nil {
		return dto.SiteProfileValue{}, nil, "", err
	}

	return profile, socialLinks, about, nil
}

func (svc *SiteService) readAboutMarkdown(ctx context.Context) (string, error) {
	raw, err := svc.settings.GetValueByKey(ctx, siteAboutKey)
	if errors.Is(err, dao.ErrSiteSettingNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	var about dto.SiteAboutValue
	if err := json.Unmarshal(raw, &about); err != nil {
		return "", err
	}
	return about.Markdown, nil
}

func normalizeSocialLinks(links []dto.WebSocialLink) ([]dto.WebSocialLink, error) {
	result := make([]dto.WebSocialLink, 0, len(links))
	for _, item := range links {
		name := strings.TrimSpace(item.Name)
		linkURL := strings.TrimSpace(item.URL)
		if name == "" && linkURL == "" {
			continue
		}
		if name == "" || linkURL == "" {
			return nil, ErrInvalidSiteSettings
		}
		parsed, err := url.ParseRequestURI(linkURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, ErrInvalidSiteSettings
		}
		result = append(result, dto.WebSocialLink{
			Name: name,
			URL:  linkURL,
		})
	}
	return result, nil
}
