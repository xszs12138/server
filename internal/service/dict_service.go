package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"blog-server/internal/dao"
	"blog-server/internal/dicttypes"
	"blog-server/internal/dto"
	"blog-server/internal/model"
)

type DictService struct {
	items    dao.DictItemDAO
	settings dao.SiteSettingDAO
	auth     *AuthService
}

func NewDictService(items dao.DictItemDAO, settings dao.SiteSettingDAO, auth *AuthService) *DictService {
	return &DictService{items: items, settings: settings, auth: auth}
}

func (svc *DictService) AdminListTypes(ctx context.Context, authorization string) ([]dto.DictTypeItem, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}

	meta, err := svc.loadTypeMeta(ctx)
	if err != nil {
		return nil, err
	}

	types := dicttypes.List()
	result := make([]dto.DictTypeItem, 0, len(types))
	for _, item := range types {
		name := item.Name
		description := item.Description
		enabled := true
		if override, ok := meta[item.Key]; ok {
			if strings.TrimSpace(override.Name) != "" {
				name = strings.TrimSpace(override.Name)
			}
			if strings.TrimSpace(override.Description) != "" {
				description = strings.TrimSpace(override.Description)
			}
			if override.Enabled != nil {
				enabled = *override.Enabled
			}
		}

		var createdAt *time.Time
		items, listErr := svc.items.List(ctx, item.Key)
		if listErr != nil {
			return nil, listErr
		}
		for i := range items {
			if createdAt == nil || items[i].CreatedAt.Before(*createdAt) {
				t := items[i].CreatedAt
				createdAt = &t
			}
		}

		result = append(result, dto.DictTypeItem{
			Key:         item.Key,
			Name:        name,
			Description: description,
			Enabled:     enabled,
			CreatedAt:   createdAt,
			WebPublic:   item.WebPublic,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result, nil
}

func (svc *DictService) AdminUpdateType(
	ctx context.Context,
	authorization string,
	dictType string,
	req dto.DictTypeUpdateRequest,
) error {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return err
	}
	if err := dicttypes.MustValidate(dictType); err != nil {
		return err
	}

	meta, err := svc.loadTypeMeta(ctx)
	if err != nil {
		return err
	}
	enabled := req.Enabled
	meta[dictType] = dictTypeMetaOverride{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Enabled:     &enabled,
	}
	return svc.saveTypeMeta(ctx, meta)
}

func (svc *DictService) AdminListItems(
	ctx context.Context,
	authorization string,
	dictType string,
) ([]dto.DictItem, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}
	if dictType != "" {
		if err := dicttypes.MustValidate(dictType); err != nil {
			return nil, err
		}
	}

	items, err := svc.items.List(ctx, dictType)
	if err != nil {
		return nil, err
	}
	return toDictItemDTOs(items), nil
}

func (svc *DictService) WebListItems(ctx context.Context, dictType string) ([]dto.WebDictItem, error) {
	if err := dicttypes.MustValidate(dictType); err != nil {
		return nil, err
	}
	if !dicttypes.IsWebPublic(dictType) {
		return nil, errors.New("该字典类型不对博客开放")
	}
	meta, err := svc.loadTypeMeta(ctx)
	if err != nil {
		return nil, err
	}
	if override, ok := meta[dictType]; ok && override.Enabled != nil && !*override.Enabled {
		return []dto.WebDictItem{}, nil
	}

	items, err := svc.items.ListEnabled(ctx, dictType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.WebDictItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.WebDictItem{
			Value: item.Value,
			Code:  dictCodeString(item.Code),
			Label: item.Label,
			Sort:  item.Sort,
		})
	}
	return result, nil
}

func (svc *DictService) AdminCreateItem(
	ctx context.Context,
	authorization string,
	dictType string,
	req dto.DictItemRequest,
) (*dto.DictItem, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}
	if strings.TrimSpace(dictType) != "" {
		req.DictType = dictType
	}
	if err := dicttypes.MustValidate(req.DictType); err != nil {
		return nil, err
	}
	item, err := svc.items.Create(ctx, toDictItemModel(req))
	if err != nil {
		if errors.Is(err, dao.ErrDictItemAlreadyExists) {
			return nil, ErrDictItemAlreadyExists
		}
		return nil, err
	}
	return toDictItemDTO(item), nil
}

func (svc *DictService) AdminUpdateItem(
	ctx context.Context,
	authorization string,
	id uint64,
	req dto.DictItemRequest,
) (*dto.DictItem, error) {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return nil, err
	}
	if err := dicttypes.MustValidate(req.DictType); err != nil {
		return nil, err
	}

	item, err := svc.items.Update(ctx, id, toDictItemModel(req))
	if err != nil {
		if errors.Is(err, dao.ErrDictItemNotFound) {
			return nil, ErrDictItemNotFound
		}
		if errors.Is(err, dao.ErrDictItemAlreadyExists) {
			return nil, ErrDictItemAlreadyExists
		}
		return nil, err
	}
	return toDictItemDTO(item), nil
}

func (svc *DictService) AdminDeleteItem(ctx context.Context, authorization string, id uint64) error {
	if err := svc.auth.EnsureAuthenticated(ctx, authorization); err != nil {
		return err
	}
	err := svc.items.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrDictItemNotFound) {
			return ErrDictItemNotFound
		}
		return err
	}
	return nil
}

func toDictItemModel(req dto.DictItemRequest) model.DictItem {
	code := strings.TrimSpace(req.Code)
	var codePtr *string
	if code != "" {
		codePtr = &code
	}
	return model.DictItem{
		DictType: req.DictType,
		Value:    req.Value,
		Code:     codePtr,
		Label:    strings.TrimSpace(req.Label),
		Enabled:  req.Enabled,
		Sort:     req.Sort,
	}
}

func toDictItemDTO(item *model.DictItem) *dto.DictItem {
	return &dto.DictItem{
		ID:        item.ID,
		DictType:  item.DictType,
		Value:     item.Value,
		Code:      dictCodeString(item.Code),
		Label:     item.Label,
		Enabled:   item.Enabled,
		Sort:      item.Sort,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func toDictItemDTOs(items []model.DictItem) []dto.DictItem {
	result := make([]dto.DictItem, 0, len(items))
	for i := range items {
		result = append(result, *toDictItemDTO(&items[i]))
	}
	return result
}

func dictCodeString(code *string) string {
	if code == nil {
		return ""
	}
	return *code
}
