package service

import (
	"context"
	"strings"

	"blog-server/internal/dto"
	"blog-server/internal/imagebed"
)

type GalleryService struct {
	client        *imagebed.Client
	defaultAlbum  int
	defaultOrder  string
}

func NewGalleryService(client *imagebed.Client, defaultAlbumID int, defaultOrder string) *GalleryService {
	order := strings.TrimSpace(defaultOrder)
	if order == "" {
		order = "newest"
	}
	return &GalleryService{
		client:       client,
		defaultAlbum: defaultAlbumID,
		defaultOrder: order,
	}
}

func (svc *GalleryService) WebListImages(
	ctx context.Context,
	page int,
	pageSize int,
	albumID int,
	order string,
) (*dto.PageResult[dto.WebGalleryImage], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 60 {
		pageSize = 60
	}

	resolvedAlbum := albumID
	if resolvedAlbum <= 0 {
		resolvedAlbum = svc.defaultAlbum
	}

	resolvedOrder := strings.TrimSpace(order)
	if resolvedOrder == "" {
		resolvedOrder = svc.defaultOrder
	}

	pageData, err := svc.client.ListImages(ctx, imagebed.ListImagesParams{
		Page:       page,
		Order:      resolvedOrder,
		AlbumID:    resolvedAlbum,
		Permission: "public",
	})
	if err != nil {
		return nil, err
	}
	if pageData == nil {
		return &dto.PageResult[dto.WebGalleryImage]{
			Items: []dto.WebGalleryImage{},
			Total: 0,
		}, nil
	}

	items := make([]dto.WebGalleryImage, 0, len(pageData.Data))
	for _, row := range pageData.Data {
		url := strings.TrimSpace(row.Links.URL)
		if url == "" {
			continue
		}
		thumb := strings.TrimSpace(row.Links.ThumbnailURL)
		if thumb == "" {
			thumb = url
		}
		name := strings.TrimSpace(row.OriginName)
		if name == "" {
			name = strings.TrimSpace(row.Name)
		}
		items = append(items, dto.WebGalleryImage{
			Key:          row.Key,
			Name:         name,
			Width:        row.Width,
			Height:       row.Height,
			Url:          url,
			ThumbnailUrl: thumb,
			Date:         row.Date,
			HumanDate:    row.HumanDate,
		})
	}

	total := pageData.Total
	if total <= 0 {
		total = len(items)
	}

	return &dto.PageResult[dto.WebGalleryImage]{
		Items: items,
		Total: total,
	}, nil
}
