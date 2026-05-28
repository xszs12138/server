package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"
)

var ErrCommentNotFound = errors.New("comment not found")
var ErrCommentDisabled = errors.New("comment disabled")
var ErrInvalidCommentParent = errors.New("invalid comment parent")
var ErrCommentRateLimited = errors.New("comment rate limited")
var ErrInvalidCommentStatus = errors.New("invalid comment status")

const commentMaxLength = 2000
const commentRateLimit = time.Minute

type CommentService struct {
	comments  dao.CommentDAO
	posts     dao.PostDAO
	auth      *AuthService
	rateLimit sync.Map
}

func NewCommentService(comments dao.CommentDAO, posts dao.PostDAO, auth *AuthService) *CommentService {
	return &CommentService{
		comments: comments,
		posts:    posts,
		auth:     auth,
	}
}

func (svc *CommentService) WebListByPostSlug(ctx context.Context, slug string) ([]dto.WebCommentItem, error) {
	post, err := svc.posts.FindPublishedBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	rows, err := svc.comments.ListByPost(ctx, post.ID, true)
	if err != nil {
		return nil, err
	}

	return buildWebCommentTree(rows), nil
}

func (svc *CommentService) WebCreate(ctx context.Context, slug string, req dto.WebCommentCreateRequest, ip string, userAgent string) (*dto.WebCommentCreateResponse, error) {
	post, err := svc.posts.FindPublishedBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > commentMaxLength {
		return nil, errors.New("invalid content")
	}
	if err := svc.checkRateLimit(ip); err != nil {
		return nil, err
	}

	status := model.CommentStatusPending
	comment, err := svc.comments.Create(ctx, model.CommentCreateInput{
		PostID:    post.ID,
		ParentID:  req.ParentID,
		Nickname:  strings.TrimSpace(req.Nickname),
		Email:     strings.TrimSpace(req.Email),
		Website:   strings.TrimSpace(req.Website),
		Content:   content,
		Status:    status,
		IP:        ip,
		UserAgent: userAgent,
	})
	if err != nil {
		if errors.Is(err, dao.ErrInvalidCommentParent) {
			return nil, ErrInvalidCommentParent
		}
		return nil, err
	}

	svc.markRateLimit(ip)
	return &dto.WebCommentCreateResponse{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

func (svc *CommentService) AdminList(ctx context.Context, authorization string, page int, pageSize int, postID uint64, status string, keyword string) (*dto.PageResult[dto.AdminCommentListItem], error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}

	rows, total, err := svc.comments.List(ctx, model.CommentListFilter{
		Page:     page,
		PageSize: pageSize,
		PostID:   postID,
		Status:   status,
		Keyword:  keyword,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.AdminCommentListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAdminCommentListItem(&row))
	}
	return &dto.PageResult[dto.AdminCommentListItem]{
		Items: items,
		Total: total,
	}, nil
}

func (svc *CommentService) AdminListByPostID(ctx context.Context, authorization string, postID uint64, page int, pageSize int, status string) (*dto.PageResult[dto.AdminCommentListItem], error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}

	if _, err := svc.posts.FindByID(ctx, postID); err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return svc.AdminList(ctx, authorization, page, pageSize, postID, status, "")
}

func (svc *CommentService) AdminApprove(ctx context.Context, authorization string, id uint64) (*dto.CommentStatusResponse, error) {
	return svc.adminUpdateStatus(ctx, authorization, id, model.CommentStatusApproved)
}

func (svc *CommentService) AdminReject(ctx context.Context, authorization string, id uint64) (*dto.CommentStatusResponse, error) {
	return svc.adminUpdateStatus(ctx, authorization, id, model.CommentStatusRejected)
}

func (svc *CommentService) AdminDelete(ctx context.Context, authorization string, id uint64) error {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return ErrInvalidToken
	}

	err := svc.comments.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return ErrCommentNotFound
		}
		return err
	}
	return nil
}

func (svc *CommentService) AdminReply(ctx context.Context, authorization string, parentID uint64, req dto.AdminCommentReplyRequest) (*dto.WebCommentCreateResponse, error) {
	user, err := svc.auth.CurrentUser(ctx, authorization)
	if err != nil {
		return nil, ErrInvalidToken
	}

	parent, err := svc.comments.FindByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	if parent.ParentID != nil {
		return nil, ErrInvalidCommentParent
	}

	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > commentMaxLength {
		return nil, errors.New("invalid content")
	}

	adminID := user.ID
	comment, err := svc.comments.Create(ctx, model.CommentCreateInput{
		PostID:      parent.PostID,
		ParentID:    &parentID,
		Nickname:    user.Nickname,
		Content:     content,
		Status:      model.CommentStatusApproved,
		IP:          "admin",
		UserAgent:   "admin",
		AdminUserID: &adminID,
	})
	if err != nil {
		if errors.Is(err, dao.ErrInvalidCommentParent) {
			return nil, ErrInvalidCommentParent
		}
		return nil, err
	}

	return &dto.WebCommentCreateResponse{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

func (svc *CommentService) adminUpdateStatus(ctx context.Context, authorization string, id uint64, status string) (*dto.CommentStatusResponse, error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}
	if !isValidCommentStatus(status) {
		return nil, ErrInvalidCommentStatus
	}

	comment, err := svc.comments.UpdateStatus(ctx, id, status)
	if err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return &dto.CommentStatusResponse{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

func (svc *CommentService) checkRateLimit(ip string) error {
	if ip == "" {
		return nil
	}
	if value, ok := svc.rateLimit.Load(ip); ok {
		if last, ok := value.(time.Time); ok && time.Since(last) < commentRateLimit {
			return ErrCommentRateLimited
		}
	}
	return nil
}

func (svc *CommentService) markRateLimit(ip string) {
	if ip == "" {
		return
	}
	svc.rateLimit.Store(ip, time.Now())
}

func isValidCommentStatus(status string) bool {
	switch status {
	case model.CommentStatusPending, model.CommentStatusApproved, model.CommentStatusRejected, model.CommentStatusSpam:
		return true
	default:
		return false
	}
}

func buildWebCommentTree(rows []model.Comment) []dto.WebCommentItem {
	topLevel := make([]model.Comment, 0)
	replies := make(map[uint64][]model.Comment)

	for _, row := range rows {
		if row.ParentID == nil {
			topLevel = append(topLevel, row)
			continue
		}
		replies[*row.ParentID] = append(replies[*row.ParentID], row)
	}

	items := make([]dto.WebCommentItem, 0, len(topLevel))
	for _, row := range topLevel {
		item := dto.WebCommentItem{
			ID:        row.ID,
			Nickname:  row.Nickname,
			Website:   row.Website,
			Content:   row.Content,
			CreatedAt: row.CreatedAt,
			Replies:   []dto.WebCommentReply{},
		}
		for _, reply := range replies[row.ID] {
			item.Replies = append(item.Replies, dto.WebCommentReply{
				ID:        reply.ID,
				Nickname:  reply.Nickname,
				Website:   reply.Website,
				Content:   reply.Content,
				CreatedAt: reply.CreatedAt,
			})
		}
		items = append(items, item)
	}
	return items
}

func toAdminCommentListItem(row *model.Comment) dto.AdminCommentListItem {
	return dto.AdminCommentListItem{
		ID: row.ID,
		Post: dto.AdminCommentPost{
			ID:    row.PostID,
			Title: row.PostTitle,
			Slug:  row.PostSlug,
		},
		ParentID:  row.ParentID,
		Nickname:  row.Nickname,
		Email:     row.Email,
		Website:   row.Website,
		Content:   row.Content,
		Status:    row.Status,
		IP:        row.IP,
		CreatedAt: row.CreatedAt,
	}
}
