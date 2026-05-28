package service

import (
	"context"
	"errors"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"
)

type CategoryService struct {
	categories dao.CategoryDAO
	posts      dao.PostDAO
	auth       *AuthService
}

func NewCategoryService(categories dao.CategoryDAO, posts dao.PostDAO, auth *AuthService) *CategoryService {
	return &CategoryService{
		categories: categories,
		posts:      posts,
		auth:       auth,
	}
}

func (svc *CategoryService) WebList(ctx context.Context) ([]dto.WebCategoryItem, error) {
	items, err := svc.categories.ListVisible(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.WebCategoryItem, 0, len(items))
	for _, item := range items {
		postCount, err := svc.posts.CountPublished(ctx, model.PostCountFilter{
			CategoryID: item.ID,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, dto.WebCategoryItem{
			ID:          item.ID,
			Name:        item.Name,
			Slug:        item.Slug,
			Description: item.Description,
			PostCount:   postCount,
		})
	}
	return result, nil
}

func (svc *CategoryService) AdminList(
	ctx context.Context,
	authorization string,
	page int,
	pageSize int,
	keyword string,
) (*dto.PageResult[dto.AdminCategoryListItem], error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	items, total, err := svc.categories.List(ctx, dao.CategoryListFilter{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminCategoryListItem, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminCategoryListItem(&item))
	}
	return &dto.PageResult[dto.AdminCategoryListItem]{
		Items: result,
		Total: total,
	}, nil
}

func (svc *CategoryService) AdminGetByID(
	ctx context.Context,
	authorization string,
	id uint64,
) (*dto.AdminCategoryDetail, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.categories.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	detail := toAdminCategoryListItem(item)
	return &dto.AdminCategoryDetail{AdminCategoryListItem: detail}, nil
}

func (svc *CategoryService) Create(
	ctx context.Context,
	authorization string,
	req dto.TaxonomyRequest,
) (*dto.TaxonomyIDResponse, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.categories.Create(ctx, toCategoryModel(req))
	if err != nil {
		if errors.Is(err, dao.ErrCategorySlugConflict) {
			return nil, ErrCategorySlugConflict
		}
		return nil, err
	}
	return &dto.TaxonomyIDResponse{ID: item.ID}, nil
}

func (svc *CategoryService) Update(
	ctx context.Context,
	authorization string,
	id uint64,
	req dto.TaxonomyRequest,
) (*dto.TaxonomyIDResponse, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	item, err := svc.categories.Update(ctx, id, toCategoryModel(req))
	if err != nil {
		if errors.Is(err, dao.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		if errors.Is(err, dao.ErrCategorySlugConflict) {
			return nil, ErrCategorySlugConflict
		}
		return nil, err
	}
	return &dto.TaxonomyIDResponse{ID: item.ID}, nil
}

func (svc *CategoryService) Delete(
	ctx context.Context,
	authorization string,
	id uint64,
) error {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return err
	}

	err := svc.categories.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return nil
}

var ErrCategorySlugConflict = errors.New("category slug already exists")

func toCategoryModel(req dto.TaxonomyRequest) model.Category {
	sort := req.Sort
	if sort == 0 {
		sort = 100
	}
	return model.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Sort:        sort,
		Visible:     req.Visible,
	}
}

func toAdminCategoryListItem(item *model.Category) dto.AdminCategoryListItem {
	return dto.AdminCategoryListItem{
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
