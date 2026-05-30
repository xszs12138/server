package service

import (
	"context"
	"encoding/json"
	"errors"

	"blog-server/internal/dao"
)

const dictTypeMetaSettingKey = "dict.typeMeta"

type dictTypeMetaOverride struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func (svc *DictService) loadTypeMeta(ctx context.Context) (map[string]dictTypeMetaOverride, error) {
	raw, err := svc.settings.GetValueByKey(ctx, dictTypeMetaSettingKey)
	if err != nil {
		if errors.Is(err, dao.ErrSiteSettingNotFound) {
			return map[string]dictTypeMetaOverride{}, nil
		}
		return nil, err
	}
	meta := map[string]dictTypeMetaOverride{}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (svc *DictService) saveTypeMeta(ctx context.Context, meta map[string]dictTypeMetaOverride) error {
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return svc.settings.UpsertValueByKey(ctx, dictTypeMetaSettingKey, json.RawMessage(raw), "字典类型展示配置")
}
