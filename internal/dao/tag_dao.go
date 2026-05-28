package dao

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/ent"
	"blog-server/internal/ent/tag"
	"blog-server/internal/model"
)

var ErrTagNotFound = errors.New("tag not found")
var ErrTagSlugConflict = errors.New("tag slug already exists")

type TagListFilter struct {
	Page     int
	PageSize int
	Keyword  string
}

type TagDAO interface {
	CountExisting(ctx context.Context, ids []uint64) (int, error)
	List(ctx context.Context, filter TagListFilter) ([]model.Tag, int, error)
	ListVisible(ctx context.Context) ([]model.Tag, error)
	FindByID(ctx context.Context, id uint64) (*model.Tag, error)
	Create(ctx context.Context, item model.Tag) (*model.Tag, error)
	Update(ctx context.Context, id uint64, item model.Tag) (*model.Tag, error)
	Delete(ctx context.Context, id uint64) error
}

type EntTagDAO struct {
	client *ent.Client
}

func NewEntTagDAO(client *ent.Client) *EntTagDAO {
	return &EntTagDAO{client: client}
}

func (dao *EntTagDAO) CountExisting(ctx context.Context, ids []uint64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return dao.client.Tag.Query().
		Where(
			tag.IDIn(ids...),
			tag.DeletedAtIsNil(),
		).
		Count(ctx)
}

func (dao *EntTagDAO) List(ctx context.Context, filter TagListFilter) ([]model.Tag, int, error) {
	query := dao.client.Tag.Query().Where(tag.DeletedAtIsNil())
	if filter.Keyword != "" {
		query = query.Where(
			tag.Or(
				tag.NameContains(filter.Keyword),
				tag.SlugContains(filter.Keyword),
			),
		)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Order(ent.Asc(tag.FieldSort), ent.Desc(tag.FieldID)).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.Tag, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toTagModel(row))
	}
	return items, total, nil
}

func (dao *EntTagDAO) ListVisible(ctx context.Context) ([]model.Tag, error) {
	rows, err := dao.client.Tag.Query().
		Where(
			tag.DeletedAtIsNil(),
			tag.VisibleEQ(true),
		).
		Order(ent.Asc(tag.FieldSort), ent.Desc(tag.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]model.Tag, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toTagModel(row))
	}
	return items, nil
}

func (dao *EntTagDAO) FindByID(ctx context.Context, id uint64) (*model.Tag, error) {
	row, err := dao.client.Tag.Query().
		Where(
			tag.IDEQ(id),
			tag.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	return toTagModel(row), nil
}

func (dao *EntTagDAO) Create(ctx context.Context, item model.Tag) (*model.Tag, error) {
	row, err := dao.client.Tag.Create().
		SetName(item.Name).
		SetSlug(item.Slug).
		SetDescription(item.Description).
		SetSort(item.Sort).
		SetVisible(item.Visible).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrTagSlugConflict
		}
		return nil, err
	}
	return toTagModel(row), nil
}

func (dao *EntTagDAO) Update(ctx context.Context, id uint64, item model.Tag) (*model.Tag, error) {
	row, err := dao.client.Tag.UpdateOneID(id).
		Where(tag.DeletedAtIsNil()).
		SetName(item.Name).
		SetSlug(item.Slug).
		SetDescription(item.Description).
		SetSort(item.Sort).
		SetVisible(item.Visible).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTagNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrTagSlugConflict
		}
		return nil, err
	}
	return toTagModel(row), nil
}

func (dao *EntTagDAO) Delete(ctx context.Context, id uint64) error {
	err := dao.client.Tag.UpdateOneID(id).
		Where(tag.DeletedAtIsNil()).
		ClearPosts().
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTagNotFound
		}
		return err
	}
	return nil
}

func toTagModel(row *ent.Tag) *model.Tag {
	return &model.Tag{
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
