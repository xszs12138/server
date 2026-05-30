package service

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"
)

var ErrPostNotFound = errors.New("post not found")
var ErrPostSlugConflict = errors.New("post slug already exists")
var ErrInvalidPostStatus = errors.New("invalid post status")
var ErrCategoryNotFound = errors.New("category not found")
var ErrTagNotFound = errors.New("tag not found")

type PostService struct {
	posts      dao.PostDAO
	categories dao.CategoryDAO
	tags       dao.TagDAO
	auth       *AuthService
}

func NewPostService(posts dao.PostDAO, categories dao.CategoryDAO, tags dao.TagDAO, auth *AuthService) *PostService {
	return &PostService{
		posts:      posts,
		categories: categories,
		tags:       tags,
		auth:       auth,
	}
}

func (svc *PostService) WebList(ctx context.Context, page int, pageSize int, categorySlug string, tagSlug string, keyword string) (*dto.PageResult[dto.WebPostListItem], error) {
	// 模拟返回时间延长
	// time.Sleep(5 * time.Second)

	posts, total, err := svc.posts.List(ctx, model.PostListFilter{
		Page:         page,
		PageSize:     pageSize,
		CategorySlug: categorySlug,
		TagSlug:      tagSlug,
		Keyword:      keyword,
		WebOnly:      true,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.WebPostListItem, 0, len(posts))
	for _, post := range posts {
		items = append(items, toWebPostListItem(&post))
	}
	return &dto.PageResult[dto.WebPostListItem]{
		Items: items,
		Total: total,
	}, nil
}

func (svc *PostService) WebGetBySlug(ctx context.Context, slug string) (*dto.WebPostDetail, error) {
	post, err := svc.posts.FindPublishedBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if err := svc.posts.IncrementViewCount(ctx, post.ID); err != nil {
		return nil, err
	}
	post.ViewCount++

	var prev *model.PostNeighbor
	var next *model.PostNeighbor
	if post.PublishedAt != nil {
		prev, next, err = svc.posts.FindNeighbors(ctx, *post.PublishedAt)
		if err != nil {
			return nil, err
		}
	}

	detail := toWebPostDetail(post)
	if prev != nil {
		detail.PrevPost = &dto.WebPostNeighbor{
			Title:       prev.Title,
			Slug:        prev.Slug,
			Cover:       prev.Cover,
			PublishedAt: prev.PublishedAt,
		}
	}
	if next != nil {
		detail.NextPost = &dto.WebPostNeighbor{
			Title:       next.Title,
			Slug:        next.Slug,
			Cover:       next.Cover,
			PublishedAt: next.PublishedAt,
		}
	}
	return detail, nil
}

func (svc *PostService) WebArchives(ctx context.Context) ([]dto.WebArchiveYear, error) {
	posts, err := svc.posts.ListPublishedArchive(ctx)
	if err != nil {
		return nil, err
	}
	return buildWebArchives(posts), nil
}

func (svc *PostService) AdminList(ctx context.Context, authorization string, page int, pageSize int, status string, categoryID uint64, tagID uint64, keyword string) (*dto.PageResult[dto.AdminPostListItem], error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}

	posts, total, err := svc.posts.List(ctx, model.PostListFilter{
		Page:       page,
		PageSize:   pageSize,
		Status:     status,
		CategoryID: categoryID,
		TagID:      tagID,
		Keyword:    keyword,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.AdminPostListItem, 0, len(posts))
	for _, post := range posts {
		items = append(items, toAdminPostListItem(&post))
	}
	return &dto.PageResult[dto.AdminPostListItem]{
		Items: items,
		Total: total,
	}, nil
}

func (svc *PostService) AdminGetByID(ctx context.Context, authorization string, id uint64) (*dto.AdminPostDetail, error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}

	post, err := svc.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return toAdminPostDetail(post), nil
}

func (svc *PostService) Create(ctx context.Context, authorization string, req dto.PostRequest) (*dto.PostIDResponse, error) {
	user, err := svc.auth.CurrentUser(ctx, authorization)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if err := svc.validatePostRequest(ctx, req); err != nil {
		return nil, err
	}

	post, err := svc.posts.Create(ctx, model.PostCreateInput{
		Title:       req.Title,
		Slug:        req.Slug,
		Cover:       req.Cover,
		Summary:     req.Summary,
		Content:     req.Content,
		ContentType: defaultContentType(req.ContentType),
		Status:      normalizePostStatus(req.Status),
		IsPinned:    req.IsPinned,
		CategoryID:  req.CategoryID,
		TagIDs:      req.TagIDs,
		AuthorID:    user.ID,
	})
	if err != nil {
		if errors.Is(err, dao.ErrPostSlugConflict) {
			return nil, ErrPostSlugConflict
		}
		return nil, err
	}

	return &dto.PostIDResponse{ID: post.ID}, nil
}

func (svc *PostService) Update(ctx context.Context, authorization string, id uint64, req dto.PostRequest) (*dto.PostIDResponse, error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}
	if err := svc.validatePostRequest(ctx, req); err != nil {
		return nil, err
	}

	existing, err := svc.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	status := normalizePostStatus(req.Status)
	post, err := svc.posts.Update(ctx, id, model.PostUpdateInput{
		Title:       req.Title,
		Slug:        req.Slug,
		Cover:       req.Cover,
		Summary:     req.Summary,
		Content:     req.Content,
		ContentType: defaultContentType(req.ContentType),
		Status:      status,
		IsPinned:    req.IsPinned,
		CategoryID:  req.CategoryID,
		TagIDs:      req.TagIDs,
	})
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		if errors.Is(err, dao.ErrPostSlugConflict) {
			return nil, ErrPostSlugConflict
		}
		return nil, err
	}

	if status == "published" && existing.PublishedAt == nil {
		now := time.Now()
		if _, err := svc.posts.UpdateStatus(ctx, id, status, &now); err != nil {
			return nil, err
		}
	}

	return &dto.PostIDResponse{ID: post.ID}, nil
}

func (svc *PostService) Delete(ctx context.Context, authorization string, id uint64) error {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return ErrInvalidToken
	}

	err := svc.posts.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return ErrPostNotFound
		}
		return err
	}
	return nil
}

func (svc *PostService) UpdateStatus(ctx context.Context, authorization string, id uint64, status string) (*dto.PostStatusResponse, error) {
	if _, err := svc.auth.CurrentUser(ctx, authorization); err != nil {
		return nil, ErrInvalidToken
	}
	if !isValidPostStatus(status) {
		return nil, ErrInvalidPostStatus
	}

	existing, err := svc.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	var publishedAt *time.Time
	if status == "published" && existing.PublishedAt == nil {
		now := time.Now()
		publishedAt = &now
	}

	result, err := svc.posts.UpdateStatus(ctx, id, status, publishedAt)
	if err != nil {
		if errors.Is(err, dao.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return &dto.PostStatusResponse{
		ID:          result.ID,
		Status:      result.Status,
		PublishedAt: result.PublishedAt,
	}, nil
}

func (svc *PostService) validatePostRequest(ctx context.Context, req dto.PostRequest) error {
	if req.Status != "" && !isValidPostStatus(normalizePostStatus(req.Status)) {
		return ErrInvalidPostStatus
	}
	if req.CategoryID != nil {
		ok, err := svc.categories.Exists(ctx, *req.CategoryID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrCategoryNotFound
		}
	}
	if len(req.TagIDs) > 0 {
		count, err := svc.tags.CountExisting(ctx, req.TagIDs)
		if err != nil {
			return err
		}
		if count != len(req.TagIDs) {
			return ErrTagNotFound
		}
	}
	return nil
}

func isValidPostStatus(status string) bool {
	switch status {
	case "draft", "published", "archived":
		return true
	default:
		return false
	}
}

func normalizePostStatus(status string) string {
	if status == "" {
		return "draft"
	}
	return status
}

func defaultContentType(contentType string) string {
	if contentType == "" {
		return "markdown"
	}
	return contentType
}

func toWebPostListItem(post *model.Post) dto.WebPostListItem {
	item := dto.WebPostListItem{
		ID:          post.ID,
		Title:       post.Title,
		Slug:        post.Slug,
		Cover:       post.Cover,
		Summary:     post.Summary,
		IsPinned:    post.IsPinned,
		ViewCount:   post.ViewCount,
		PublishedAt: post.PublishedAt,
		Tags:        toWebPostTags(post.Tags, true),
	}
	if post.Category != nil && post.Category.Visible {
		item.Category = &dto.WebPostCategory{
			ID:   post.Category.ID,
			Name: post.Category.Name,
			Slug: post.Category.Slug,
		}
	}
	return item
}

func toWebPostDetail(post *model.Post) *dto.WebPostDetail {
	item := &dto.WebPostDetail{
		ID:          post.ID,
		Title:       post.Title,
		Slug:        post.Slug,
		Cover:       post.Cover,
		Summary:     post.Summary,
		Content:     post.Content,
		ContentType: post.ContentType,
		ViewCount:   post.ViewCount,
		PublishedAt: post.PublishedAt,
		Tags:        toWebPostTags(post.Tags, true),
	}
	if post.Category != nil && post.Category.Visible {
		item.Category = &dto.WebPostCategory{
			ID:   post.Category.ID,
			Name: post.Category.Name,
			Slug: post.Category.Slug,
		}
	}
	return item
}

func toWebPostTags(tags []model.Tag, visibleOnly bool) []dto.WebPostTag {
	if len(tags) == 0 {
		return []dto.WebPostTag{}
	}
	items := make([]dto.WebPostTag, 0, len(tags))
	for _, tag := range tags {
		if visibleOnly && !tag.Visible {
			continue
		}
		items = append(items, dto.WebPostTag{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}
	return items
}

func toAdminPostListItem(post *model.Post) dto.AdminPostListItem {
	item := dto.AdminPostListItem{
		ID:          post.ID,
		Title:       post.Title,
		Slug:        post.Slug,
		Cover:       post.Cover,
		Summary:     post.Summary,
		Status:      post.Status,
		IsPinned:    post.IsPinned,
		ViewCount:   post.ViewCount,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		PublishedAt: post.PublishedAt,
		Tags:        toAdminPostTags(post.Tags),
	}
	if post.Category != nil {
		item.Category = &dto.AdminPostCategory{
			ID:   post.Category.ID,
			Name: post.Category.Name,
		}
	}
	return item
}

func toAdminPostDetail(post *model.Post) *dto.AdminPostDetail {
	tagIDs := make([]uint64, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tagIDs = append(tagIDs, tag.ID)
	}
	return &dto.AdminPostDetail{
		ID:          post.ID,
		Title:       post.Title,
		Slug:        post.Slug,
		Cover:       post.Cover,
		Summary:     post.Summary,
		Content:     post.Content,
		ContentType: post.ContentType,
		Status:      post.Status,
		IsPinned:    post.IsPinned,
		ViewCount:   post.ViewCount,
		CategoryID:  post.CategoryID,
		TagIDs:      tagIDs,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		PublishedAt: post.PublishedAt,
	}
}

func toAdminPostTags(tags []model.Tag) []dto.AdminPostTag {
	if len(tags) == 0 {
		return []dto.AdminPostTag{}
	}
	items := make([]dto.AdminPostTag, 0, len(tags))
	for _, tag := range tags {
		items = append(items, dto.AdminPostTag{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return items
}

func buildWebArchives(posts []model.ArchivePost) []dto.WebArchiveYear {
	type monthKey struct {
		year  int
		month int
	}

	yearOrder := make([]int, 0)
	yearIndex := make(map[int]int)
	monthOrder := make(map[int][]monthKey)
	monthPosts := make(map[monthKey][]dto.WebArchivePost)

	for _, post := range posts {
		year := post.PublishedAt.Year()
		month := int(post.PublishedAt.Month())
		key := monthKey{year: year, month: month}

		if _, ok := yearIndex[year]; !ok {
			yearIndex[year] = len(yearOrder)
			yearOrder = append(yearOrder, year)
			monthOrder[year] = make([]monthKey, 0)
		}
		if len(monthPosts[key]) == 0 {
			monthOrder[year] = append(monthOrder[year], key)
		}
		monthPosts[key] = append(monthPosts[key], dto.WebArchivePost{
			ID:          post.ID,
			Title:       post.Title,
			Slug:        post.Slug,
			Cover:       post.Cover,
			Summary:     post.Summary,
			ViewCount:   post.ViewCount,
			PublishedAt: post.PublishedAt,
		})
	}

	result := make([]dto.WebArchiveYear, 0, len(yearOrder))
	for _, year := range yearOrder {
		months := make([]dto.WebArchiveMonth, 0, len(monthOrder[year]))
		for _, key := range monthOrder[year] {
			months = append(months, dto.WebArchiveMonth{
				Month: key.month,
				Posts: monthPosts[key],
			})
		}
		result = append(result, dto.WebArchiveYear{
			Year:   year,
			Months: months,
		})
	}
	return result
}
