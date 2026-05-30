package dao

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/ent"
	"blog-server/internal/ent/category"
	"blog-server/internal/ent/post"
	"blog-server/internal/ent/tag"
	"blog-server/internal/model"

	"entgo.io/ent/dialect/sql"
)

var ErrPostNotFound = errors.New("post not found")
var ErrPostSlugConflict = errors.New("post slug already exists")
var ErrCategoryNotFound = errors.New("category not found")

type PostDAO interface {
	List(ctx context.Context, filter model.PostListFilter) ([]model.Post, int, error)
	CountPublished(ctx context.Context, filter model.PostCountFilter) (int, error)
	ListPublishedArchive(ctx context.Context) ([]model.ArchivePost, error)
	FindByID(ctx context.Context, id uint64) (*model.Post, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*model.Post, error)
	FindNeighbors(ctx context.Context, publishedAt time.Time) (*model.PostNeighbor, *model.PostNeighbor, error)
	Create(ctx context.Context, input model.PostCreateInput) (*model.Post, error)
	Update(ctx context.Context, id uint64, input model.PostUpdateInput) (*model.Post, error)
	UpdateStatus(ctx context.Context, id uint64, status string, publishedAt *time.Time) (*model.PostStatusUpdate, error)
	Delete(ctx context.Context, id uint64) error
	IncrementViewCount(ctx context.Context, id uint64) error
}

type EntPostDAO struct {
	client *ent.Client
}

func NewEntPostDAO(client *ent.Client) *EntPostDAO {
	return &EntPostDAO{client: client}
}

func (dao *EntPostDAO) List(ctx context.Context, filter model.PostListFilter) ([]model.Post, int, error) {
	query := dao.client.Post.Query().
		Where(post.DeletedAtIsNil()).
		WithCategory().
		WithTags()

	dao.applyListFilters(query, filter)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if filter.WebOnly {
		query = query.
			Order(
				post.ByIsPinned(sql.OrderDesc()),
				post.ByPublishedAt(sql.OrderDesc()),
			)
	} else {
		query = query.Order(ent.Desc(post.FieldUpdatedAt))
	}

	offset := (filter.Page - 1) * filter.PageSize
	rows, err := query.
		Offset(offset).
		Limit(filter.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.Post, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toPostModel(row))
	}
	return items, total, nil
}

func (dao *EntPostDAO) CountPublished(ctx context.Context, filter model.PostCountFilter) (int, error) {
	now := time.Now()
	query := dao.client.Post.Query().
		Where(
			post.DeletedAtIsNil(),
			post.StatusEQ("published"),
			post.PublishedAtNotNil(),
			post.PublishedAtLTE(now),
		)

	if filter.CategoryID != 0 {
		query = query.Where(post.CategoryIdEQ(filter.CategoryID))
	}
	if filter.TagID != 0 {
		query = query.Where(post.HasTagsWith(
			tag.IDEQ(filter.TagID),
			tag.VisibleEQ(true),
			tag.DeletedAtIsNil(),
		))
	}

	return query.Count(ctx)
}

func (dao *EntPostDAO) ListPublishedArchive(ctx context.Context) ([]model.ArchivePost, error) {
	now := time.Now()
	rows, err := dao.client.Post.Query().
		Where(
			post.DeletedAtIsNil(),
			post.StatusEQ("published"),
			post.PublishedAtNotNil(),
			post.PublishedAtLTE(now),
		).
		Order(post.ByPublishedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]model.ArchivePost, 0, len(rows))
	for _, row := range rows {
		if row.PublishedAt == nil {
			continue
		}
		items = append(items, model.ArchivePost{
			ID:          row.ID,
			Title:       row.Title,
			Slug:        row.Slug,
			Cover:       row.Cover,
			Summary:     row.Summary,
			ViewCount:   int(row.ViewCount),
			PublishedAt: *row.PublishedAt,
		})
	}
	return items, nil
}

func (dao *EntPostDAO) FindByID(ctx context.Context, id uint64) (*model.Post, error) {
	row, err := dao.client.Post.Query().
		Where(
			post.IDEQ(id),
			post.DeletedAtIsNil(),
		).
		WithCategory().
		WithTags().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return toPostModel(row), nil
}

func (dao *EntPostDAO) FindPublishedBySlug(ctx context.Context, slug string) (*model.Post, error) {
	now := time.Now()
	row, err := dao.client.Post.Query().
		Where(
			post.SlugEQ(slug),
			post.DeletedAtIsNil(),
			post.StatusEQ("published"),
			post.PublishedAtNotNil(),
			post.PublishedAtLTE(now),
		).
		WithCategory(func(q *ent.CategoryQuery) {
			q.Where(category.DeletedAtIsNil())
		}).
		WithTags(func(q *ent.TagQuery) {
			q.Where(tag.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return toPostModel(row), nil
}

func (dao *EntPostDAO) FindNeighbors(ctx context.Context, publishedAt time.Time) (*model.PostNeighbor, *model.PostNeighbor, error) {
	now := time.Now()

	prevRow, err := dao.client.Post.Query().
		Where(
			post.DeletedAtIsNil(),
			post.StatusEQ("published"),
			post.PublishedAtNotNil(),
			post.PublishedAtLTE(now),
			post.PublishedAtLT(publishedAt),
		).
		Order(post.ByPublishedAt(sql.OrderDesc())).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, nil, err
	}

	nextRow, err := dao.client.Post.Query().
		Where(
			post.DeletedAtIsNil(),
			post.StatusEQ("published"),
			post.PublishedAtNotNil(),
			post.PublishedAtLTE(now),
			post.PublishedAtGT(publishedAt),
		).
		Order(post.ByPublishedAt(sql.OrderAsc())).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, nil, err
	}

	var prev *model.PostNeighbor
	if prevRow != nil {
		prev = &model.PostNeighbor{
			Title:       prevRow.Title,
			Slug:        prevRow.Slug,
			Cover:       prevRow.Cover,
			PublishedAt: prevRow.PublishedAt,
		}
	}
	var next *model.PostNeighbor
	if nextRow != nil {
		next = &model.PostNeighbor{
			Title:       nextRow.Title,
			Slug:        nextRow.Slug,
			Cover:       nextRow.Cover,
			PublishedAt: nextRow.PublishedAt,
		}
	}
	return prev, next, nil
}

func (dao *EntPostDAO) Create(ctx context.Context, input model.PostCreateInput) (*model.Post, error) {
	create := dao.client.Post.Create().
		SetTitle(input.Title).
		SetSlug(input.Slug).
		SetCover(input.Cover).
		SetSummary(input.Summary).
		SetContent(input.Content).
		SetStatus(defaultPostStatus(input.Status)).
		SetIsPinned(input.IsPinned).
		SetAuthorID(input.AuthorID)

	if input.ContentType != "" {
		create.SetContentType(input.ContentType)
	}
	if input.CategoryID != nil {
		create.SetCategoryID(*input.CategoryID)
	}
	if input.Status == "published" {
		create.SetPublishedAt(time.Now())
	}

	row, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrPostSlugConflict
		}
		return nil, err
	}

	if len(input.TagIDs) > 0 {
		if err := dao.client.Post.UpdateOneID(row.ID).AddTagIDs(input.TagIDs...).Exec(ctx); err != nil {
			return nil, err
		}
	}

	return dao.FindByID(ctx, row.ID)
}

func (dao *EntPostDAO) Update(ctx context.Context, id uint64, input model.PostUpdateInput) (*model.Post, error) {
	update := dao.client.Post.UpdateOneID(id).
		SetTitle(input.Title).
		SetSlug(input.Slug).
		SetCover(input.Cover).
		SetSummary(input.Summary).
		SetContent(input.Content).
		SetStatus(defaultPostStatus(input.Status)).
		SetIsPinned(input.IsPinned).
		ClearTags()

	if input.ContentType != "" {
		update.SetContentType(input.ContentType)
	} else {
		update.SetContentType("markdown")
	}
	if input.CategoryID != nil {
		update.SetCategoryID(*input.CategoryID)
	} else {
		update.ClearCategoryId()
	}

	if err := update.Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrPostSlugConflict
		}
		return nil, err
	}

	if len(input.TagIDs) > 0 {
		if err := dao.client.Post.UpdateOneID(id).AddTagIDs(input.TagIDs...).Exec(ctx); err != nil {
			return nil, err
		}
	}

	return dao.FindByID(ctx, id)
}

func (dao *EntPostDAO) UpdateStatus(ctx context.Context, id uint64, status string, publishedAt *time.Time) (*model.PostStatusUpdate, error) {
	update := dao.client.Post.UpdateOneID(id).SetStatus(status)
	if publishedAt != nil {
		update.SetPublishedAt(*publishedAt)
	}

	if err := update.Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	row, err := dao.client.Post.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &model.PostStatusUpdate{
		ID:     row.ID,
		Status: row.Status,
	}
	if row.PublishedAt != nil {
		t := *row.PublishedAt
		result.PublishedAt = &t
	}
	return result, nil
}

func (dao *EntPostDAO) Delete(ctx context.Context, id uint64) error {
	err := dao.client.Post.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrPostNotFound
		}
		return err
	}
	return nil
}

func (dao *EntPostDAO) IncrementViewCount(ctx context.Context, id uint64) error {
	return dao.client.Post.UpdateOneID(id).
		AddViewCount(1).
		Exec(ctx)
}

func (dao *EntPostDAO) applyListFilters(query *ent.PostQuery, filter model.PostListFilter) {
	if filter.WebOnly {
		now := time.Now()
		query.Where(
			post.StatusEQ("published"),
			post.PublishedAtNotNil(),
			post.PublishedAtLTE(now),
		)
	} else if filter.Status != "" {
		query.Where(post.StatusEQ(filter.Status))
	}

	if filter.CategoryID != 0 {
		query.Where(post.CategoryIdEQ(filter.CategoryID))
	}
	if filter.CategorySlug != "" {
		query.Where(post.HasCategoryWith(
			category.SlugEQ(filter.CategorySlug),
			category.VisibleEQ(true),
			category.DeletedAtIsNil(),
		))
	}
	if filter.TagID != 0 {
		query.Where(post.HasTagsWith(tag.IDEQ(filter.TagID)))
	}
	if filter.TagSlug != "" {
		query.Where(post.HasTagsWith(
			tag.SlugEQ(filter.TagSlug),
			tag.VisibleEQ(true),
			tag.DeletedAtIsNil(),
		))
	}
	if filter.Keyword != "" {
		if filter.WebOnly {
			query.Where(post.Or(
				post.TitleContains(filter.Keyword),
				post.SummaryContains(filter.Keyword),
			))
		} else {
			query.Where(post.Or(
				post.TitleContains(filter.Keyword),
				post.SummaryContains(filter.Keyword),
				post.ContentContains(filter.Keyword),
			))
		}
	}
}

func defaultPostStatus(status string) string {
	switch status {
	case "draft", "published", "archived":
		return status
	default:
		return "draft"
	}
}

func toPostModel(row *ent.Post) *model.Post {
	item := &model.Post{
		ID:          row.ID,
		Title:       row.Title,
		Slug:        row.Slug,
		Cover:       row.Cover,
		Summary:     row.Summary,
		Content:     row.Content,
		ContentType: row.ContentType,
		Status:      row.Status,
		IsPinned:    row.IsPinned,
		ViewCount:   row.ViewCount,
		AuthorID:    row.AuthorId,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.CategoryId != nil {
		id := *row.CategoryId
		item.CategoryID = &id
	}
	if row.PublishedAt != nil {
		t := *row.PublishedAt
		item.PublishedAt = &t
	}
	if row.Edges.Category != nil {
		cat := row.Edges.Category
		item.Category = &model.Category{
			ID:          cat.ID,
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Sort:        cat.Sort,
			Visible:     cat.Visible,
		}
	}
	if len(row.Edges.Tags) > 0 {
		item.Tags = make([]model.Tag, 0, len(row.Edges.Tags))
		for _, tg := range row.Edges.Tags {
			item.Tags = append(item.Tags, model.Tag{
				ID:          tg.ID,
				Name:        tg.Name,
				Slug:        tg.Slug,
				Description: tg.Description,
				Sort:        tg.Sort,
				Visible:     tg.Visible,
			})
		}
	}
	return item
}
