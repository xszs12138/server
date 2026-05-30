package dto

type WebSocialLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type WebSiteInfo struct {
	ID           uint64          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Logo         string          `json:"logo"`
	Icp          string          `json:"icp"`
	About        string          `json:"about"`
	SocialLinks  []WebSocialLink `json:"socialLinks"`
}

type SiteProfileValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Icp         string `json:"icp"`
}

// AdminSiteSettings 后台站点设置（与 site.profile + site.socialLinks 聚合）
type AdminSiteSettings struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Logo        string          `json:"logo"`
	Icp         string          `json:"icp"`
	About       string          `json:"about"`
	SocialLinks []WebSocialLink `json:"socialLinks"`
}

type SiteAboutValue struct {
	Markdown string `json:"markdown"`
}

type AdminSiteSettingsUpdateRequest struct {
	Name        string          `json:"name" binding:"required,max=120"`
	Description string          `json:"description" binding:"max=500"`
	Logo        string          `json:"logo" binding:"max=512"`
	Icp         string          `json:"icp" binding:"max=64"`
	About       string          `json:"about" binding:"max=100000"`
	SocialLinks []WebSocialLink `json:"socialLinks"`
}
