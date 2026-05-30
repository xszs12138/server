package dto

type WebGalleryImage struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Url          string `json:"url"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Date         string `json:"date"`
	HumanDate    string `json:"humanDate"`
}
