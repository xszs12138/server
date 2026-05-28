package service

import (
	"context"
	"errors"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"
)

type TagService struct {
	tags  dao.TagDAO
	posts dao.PostDAO
	auth  *AuthService
}

func NewTagService(tags dao.TagDAO, posts dao.PostDAO, auth *AuthService) *TagService {
	return &TagService{
		tags:  tags,
		posts: posts,
		auth:  auth,
	}
}

func (svc *TagService) WebList(ctx context.Context) ([]dto.WebTagItem, error) {
	items, err := svc.tags.ListVisible(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.WebTagItem, 0, len(items))
	for _, item := range items {
		postCount, err := svc.posts.CountPublished(ctx, model.PostCountFilter{
			TagID: item.ID,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, dto.WebTagItem{
			ID:        item.ID,
			Name:      item.Name,
			Slug:      item.Slug,
			PostCount: postCount,
		})
	}
	return result, nil
}

func (svc *TagService) AdminList(
	ctx context.Context,
	authorization string,
	page int,
	pageSize int,
	keyword string,
) (*dto.PageResult[dto.AdminTagListItem], error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	items, total, err := svc.tags.List(ctx, dao.TagListFilter{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminTagListItem, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminTagListItem(&item))
	}
	return &dto.PageResult[dto.AdminTagListItem]{
		Items: result,
		Total: total,
	}, nil
}

func (svc *TagService) AdminGetByID(
	ctx context.Context,
	authorization string,
	id uint64,
) (*dto.AdminTagDetail, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrTagNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	detail := toAdminTagListItem(item)
	return &dto.AdminTagDetail{AdminTagListItem: detail}, nil
}

func (svc *TagService) Create(
	ctx context.Context,
	authorization string,
	req dto.TaxonomyRequest,
) (*dto.TaxonomyIDResponse, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.tags.Create(ctx, toTagModel(req))
	if err != nil {
		if errors.Is(err, dao.ErrTagSlugConflict) {
			return nil, ErrTagSlugConflict
		}
		return nil, err
	}
	return &dto.TaxonomyIDResponse{ID: item.ID}, nil
}

func (svc *TagService) Update(
	ctx context.Context,
	authorization string,
	id uint64,
	req dto.TaxonomyRequest,
) (*dto.TaxonomyIDResponse, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.tags.Update(ctx, id, toTagModel(req))
	if err != nil {
		if errors.Is(err, dao.ErrTagNotFound) {
			return nil, ErrTagNotFound
		}
		if errors.Is(err, dao.ErrTagSlugConflict) {
			return nil, ErrTagSlugConflict
		}
		return nil, err
	}
	return &dto.TaxonomyIDResponse{ID: item.ID}, nil
}

func (svc *TagService) Delete(
	ctx context.Context,
	authorization string,
	id uint64,
) error {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return err
	}

	err := svc.tags.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrTagNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	return nil
}

var ErrTagSlugConflict = errors.New("tag slug already exists")

func toTagModel(req dto.TaxonomyRequest) model.Tag {
	sort := req.Sort
	if sort == 0 {
		sort = 100
	}
	return model.Tag{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Sort:        sort,
		Visible:     req.Visible,
	}
}

func toAdminTagListItem(item *model.Tag) dto.AdminTagListItem {
	return dto.AdminTagListItem{
		ID:          item.ID,
		Name:        item.Name,
		Slug:        item.Slug,
		Description: item.Description,
		Sort:        item.Sort,
		Visible:     item.Visible,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
