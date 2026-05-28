package dao

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/ent"
	"blog-server/internal/ent/comment"
	"blog-server/internal/model"

	"entgo.io/ent/dialect/sql"
)

var ErrCommentNotFound = errors.New("comment not found")
var ErrInvalidCommentParent = errors.New("invalid comment parent")

type CommentDAO interface {
	ListByPost(ctx context.Context, postID uint64, webOnly bool) ([]model.Comment, error)
	List(ctx context.Context, filter model.CommentListFilter) ([]model.Comment, int, error)
	FindByID(ctx context.Context, id uint64) (*model.Comment, error)
	Create(ctx context.Context, input model.CommentCreateInput) (*model.Comment, error)
	UpdateStatus(ctx context.Context, id uint64, status string) (*model.Comment, error)
	Delete(ctx context.Context, id uint64) error
}

type EntCommentDAO struct {
	client *ent.Client
}

func NewEntCommentDAO(client *ent.Client) *EntCommentDAO {
	return &EntCommentDAO{client: client}
}

func (dao *EntCommentDAO) ListByPost(ctx context.Context, postID uint64, webOnly bool) ([]model.Comment, error) {
	query := dao.client.Comment.Query().
		Where(
			comment.PostIdEQ(postID),
			comment.DeletedAtIsNil(),
		).
		Order(comment.ByCreatedAt(sql.OrderAsc()))

	if webOnly {
		query = query.Where(comment.StatusEQ(model.CommentStatusApproved))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]model.Comment, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toCommentModel(row))
	}
	return items, nil
}

func (dao *EntCommentDAO) List(ctx context.Context, filter model.CommentListFilter) ([]model.Comment, int, error) {
	query := dao.client.Comment.Query().
		Where(comment.DeletedAtIsNil()).
		WithPost()

	if filter.PostID != 0 {
		query = query.Where(comment.PostIdEQ(filter.PostID))
	}
	if filter.Status != "" {
		query = query.Where(comment.StatusEQ(filter.Status))
	}
	if filter.Keyword != "" {
		query = query.Where(comment.Or(
			comment.NicknameContains(filter.Keyword),
			comment.ContentContains(filter.Keyword),
		))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	rows, err := query.
		Order(ent.Desc(comment.FieldCreatedAt)).
		Offset(offset).
		Limit(filter.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.Comment, 0, len(rows))
	for _, row := range rows {
		item := toCommentModel(row)
		if row.Edges.Post != nil {
			item.PostTitle = row.Edges.Post.Title
			item.PostSlug = row.Edges.Post.Slug
		}
		items = append(items, *item)
	}
	return items, total, nil
}

func (dao *EntCommentDAO) FindByID(ctx context.Context, id uint64) (*model.Comment, error) {
	row, err := dao.client.Comment.Query().
		Where(
			comment.IDEQ(id),
			comment.DeletedAtIsNil(),
		).
		WithPost().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return toCommentModel(row), nil
}

func (dao *EntCommentDAO) Create(ctx context.Context, input model.CommentCreateInput) (*model.Comment, error) {
	if input.ParentID != nil {
		parent, err := dao.client.Comment.Query().
			Where(
				comment.IDEQ(*input.ParentID),
				comment.PostIdEQ(input.PostID),
				comment.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrInvalidCommentParent
			}
			return nil, err
		}
		if parent.ParentId != nil {
			return nil, ErrInvalidCommentParent
		}
	}

	create := dao.client.Comment.Create().
		SetPostId(input.PostID).
		SetNickname(input.Nickname).
		SetContent(input.Content).
		SetStatus(input.Status).
		SetIP(input.IP).
		SetUserAgent(input.UserAgent)

	if input.Email != "" {
		create.SetEmail(input.Email)
	}
	if input.Website != "" {
		create.SetWebsite(input.Website)
	}
	if input.ParentID != nil {
		create.SetParentId(*input.ParentID)
	}
	if input.AdminUserID != nil {
		create.SetAdminUserId(*input.AdminUserID)
	}

	row, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return toCommentModel(row), nil
}

func (dao *EntCommentDAO) UpdateStatus(ctx context.Context, id uint64, status string) (*model.Comment, error) {
	err := dao.client.Comment.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return dao.FindByID(ctx, id)
}

func (dao *EntCommentDAO) Delete(ctx context.Context, id uint64) error {
	err := dao.client.Comment.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCommentNotFound
		}
		return err
	}
	return nil
}

func toCommentModel(row *ent.Comment) *model.Comment {
	item := &model.Comment{
		ID:        row.ID,
		PostID:    row.PostId,
		Nickname:  row.Nickname,
		Content:   row.Content,
		Status:    row.Status,
		IP:        row.IP,
		UserAgent: row.UserAgent,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.ParentId != nil {
		id := *row.ParentId
		item.ParentID = &id
	}
	if row.Email != nil {
		item.Email = *row.Email
	}
	if row.Website != nil {
		item.Website = *row.Website
	}
	if row.AdminUserId != nil {
		id := *row.AdminUserId
		item.AdminUserID = &id
	}
	return item
}
