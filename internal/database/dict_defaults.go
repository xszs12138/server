package database

import (
	"context"

	"blog-server/internal/dicttypes"
	"blog-server/internal/ent"
	entdictitem "blog-server/internal/ent/dictitem"
)

func migrateLegacyDictTypes(ctx context.Context, client *ent.Client) error {
	_, err := client.DictItem.Update().
		Where(entdictitem.DictTypeEQ("game.genre")).
		SetDictType(dicttypes.GameGenre).
		Save(ctx)
	return err
}

func ensureDefaultDictItems(ctx context.Context, client *ent.Client) error {
	if err := migrateLegacyDictTypes(ctx, client); err != nil {
		return err
	}
	if err := backfillDictItemCodes(ctx, client); err != nil {
		return err
	}
	defaults := []struct {
		dictType string
		value    int
		code     string
		label    string
		sort     int
	}{
		{dictType: dicttypes.Operation, value: 1, code: "login", label: "登录", sort: 1},
		{dictType: dicttypes.Operation, value: 2, code: "logout", label: "退出", sort: 2},
		{dictType: dicttypes.GameGenre, value: 1, code: "rpg", label: "RPG", sort: 1},
		{dictType: dicttypes.GameGenre, value: 2, code: "action", label: "动作", sort: 2},
		{dictType: dicttypes.GameGenre, value: 3, code: "adventure", label: "冒险", sort: 3},
		{dictType: dicttypes.GameGenre, value: 4, code: "indie", label: "独立", sort: 4},
		{dictType: dicttypes.GameGenre, value: 5, code: "strategy", label: "策略", sort: 5},
		{dictType: dicttypes.GameGenre, value: 6, code: "online", label: "联机", sort: 6},
	}

	for _, item := range defaults {
		exists, err := client.DictItem.Query().
			Where(
				entdictitem.DictTypeEQ(item.dictType),
				entdictitem.ValueEQ(item.value),
			).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		code := item.code
		if _, err := client.DictItem.Create().
			SetDictType(item.dictType).
			SetValue(item.value).
			SetCode(code).
			SetLabel(item.label).
			SetEnabled(true).
			SetSort(item.sort).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
