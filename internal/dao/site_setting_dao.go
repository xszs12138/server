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
	UpsertValueByKey(ctx context.Context, key string, value json.RawMessage, description string) error
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

func (dao *EntSiteSettingDAO) UpsertValueByKey(
	ctx context.Context,
	key string,
	value json.RawMessage,
	description string,
) error {
	existing, err := dao.client.SiteSetting.Query().
		Where(entsitesetting.SettingKeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			_, createErr := dao.client.SiteSetting.Create().
				SetSettingKey(key).
				SetValue(value).
				SetDescription(description).
				Save(ctx)
			return createErr
		}
		return err
	}

	update := dao.client.SiteSetting.UpdateOneID(existing.ID).
		SetValue(value)
	if description != "" {
		update = update.SetDescription(description)
	}
	return update.Exec(ctx)
}
