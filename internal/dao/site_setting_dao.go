package dao

import (
	"context"
	"encoding/json"
	"errors"

	"blog-server/internal/ent"
	entsitesetting "blog-server/internal/ent/sitesetting"
)

var ErrSiteSettingNotFound = errors.New("site setting not found")

type SiteSettingDAO interface {
	GetValueByKey(ctx context.Context, key string) (json.RawMessage, error)
}

type EntSiteSettingDAO struct {
	client *ent.Client
}

func NewEntSiteSettingDAO(client *ent.Client) *EntSiteSettingDAO {
	return &EntSiteSettingDAO{client: client}
}

func (dao *EntSiteSettingDAO) GetValueByKey(ctx context.Context, key string) (json.RawMessage, error) {
	row, err := dao.client.SiteSetting.Query().
		Where(entsitesetting.SettingKeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSiteSettingNotFound
		}
		return nil, err
	}
	return append(json.RawMessage(nil), row.Value...), nil
}
