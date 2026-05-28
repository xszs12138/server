package dao

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/ent"
	"blog-server/internal/ent/category"
	"blog-server/internal/ent/post"
	"blog-server/internal/model"
)

var ErrCategorySlugConflict = errors.New("category slug already exists")

type CategoryListFilter struct {
	Page     int
	PageSize int
	Keyword  string
}

type CategoryDAO interface {
	Exists(ctx context.Context, id uint64) (bool, error)
	List(ctx context.Context, filter CategoryListFilter) ([]model.Category, int, error)
	ListVisible(ctx context.Context) ([]model.Category, error)
	FindByID(ctx context.Context, id uint64) (*model.Category, error)
	Create(ctx context.Context, item model.Category) (*model.Category, error)
	Update(ctx context.Context, id uint64, item model.Category) (*model.Category, error)
	Delete(ctx context.Context, id uint64) error
}

type EntCategoryDAO struct {
	client *ent.Client
}

func NewEntCategoryDAO(client *ent.Client) *EntCategoryDAO {
	return &EntCategoryDAO{client: client}
}

func (dao *EntCategoryDAO) Exists(ctx context.Context, id uint64) (bool, error) {
	return dao.client.Category.Query().
		Where(
			category.IDEQ(id),
			category.DeletedAtIsNil(),
		).
		Exist(ctx)
}

func (dao *EntCategoryDAO) List(ctx context.Context, filter CategoryListFilter) ([]model.Category, int, error) {
	query := dao.client.Category.Query().Where(category.DeletedAtIsNil())
	if filter.Keyword != "" {
		query = query.Where(
			category.Or(
				category.NameContains(filter.Keyword),
				category.SlugContains(filter.Keyword),
			),
		)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Order(ent.Asc(category.FieldSort), ent.Desc(category.FieldID)).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.Category, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toCategoryModel(row))
	}
	return items, total, nil
}

func (dao *EntCategoryDAO) ListVisible(ctx context.Context) ([]model.Category, error) {
	rows, err := dao.client.Category.Query().
		Where(
			category.DeletedAtIsNil(),
			category.VisibleEQ(true),
		).
		Order(ent.Asc(category.FieldSort), ent.Desc(category.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]model.Category, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toCategoryModel(row))
	}
	return items, nil
}

func (dao *EntCategoryDAO) FindByID(ctx context.Context, id uint64) (*model.Category, error) {
	row, err := dao.client.Category.Query().
		Where(
			category.IDEQ(id),
			category.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return toCategoryModel(row), nil
}

func (dao *EntCategoryDAO) Create(ctx context.Context, item model.Category) (*model.Category, error) {
	row, err := dao.client.Category.Create().
		SetName(item.Name).
		SetSlug(item.Slug).
		SetDescription(item.Description).
		SetSort(item.Sort).
		SetVisible(item.Visible).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrCategorySlugConflict
		}
		return nil, err
	}
	return toCategoryModel(row), nil
}

func (dao *EntCategoryDAO) Update(ctx context.Context, id uint64, item model.Category) (*model.Category, error) {
	row, err := dao.client.Category.UpdateOneID(id).
		Where(category.DeletedAtIsNil()).
		SetName(item.Name).
		SetSlug(item.Slug).
		SetDescription(item.Description).
		SetSort(item.Sort).
		SetVisible(item.Visible).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCategoryNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrCategorySlugConflict
		}
		return nil, err
	}
	return toCategoryModel(row), nil
}

func (dao *EntCategoryDAO) Delete(ctx context.Context, id uint64) error {
	exists, err := dao.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCategoryNotFound
	}

	if err := dao.client.Post.Update().
		Where(
			post.CategoryIdEQ(id),
			post.DeletedAtIsNil(),
		).
		ClearCategoryId().
		Exec(ctx); err != nil {
		return err
	}

	err = dao.client.Category.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCategoryNotFound
		}
		return err
	}
	return nil
}

func toCategoryModel(row *ent.Category) *model.Category {
	return &model.Category{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		Sort:        row.Sort,
		Visible:     row.Visible,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
