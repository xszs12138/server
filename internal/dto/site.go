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
	SocialLinks  []WebSocialLink `json:"socialLinks"`
}

type SiteProfileValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Icp         string `json:"icp"`
}
