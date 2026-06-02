package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"blog-server/internal/config"
	"blog-server/internal/ent"
	entsitesetting "blog-server/internal/ent/sitesetting"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-sql-driver/mysql"
)

func Open(ctx context.Context, cfg config.Config) (*ent.Client, error) {
	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.MySQL, db)))

	if cfg.AutoMigrate {
		// 旧库列重命名（表不存在时忽略）
		if err := renameLegacyColumns(ctx, db); err != nil {
			client.Close()
			return nil, err
		}
		// 须先建表；全新库不能在 Schema.Create 之前 ALTER dictItems
		if err := client.Schema.Create(ctx); err != nil {
			client.Close()
			return nil, err
		}
		// 旧库升级：补 dictItems.code 列与唯一索引
		if err := ensureDictItemsCodeColumn(ctx, db); err != nil {
			client.Close()
			return nil, err
		}
		if err := backfillDictItemCodes(ctx, client); err != nil {
			client.Close()
			return nil, err
		}
		if err := dropLegacyColumns(ctx, db); err != nil {
			client.Close()
			return nil, err
		}
		if err := ensureDefaultSiteSettings(ctx, client); err != nil {
			client.Close()
			return nil, err
		}
		if err := ensureDefaultDictItems(ctx, client); err != nil {
			client.Close()
			return nil, err
		}
	}

	return client, nil
}

func ensureDefaultSiteSettings(ctx context.Context, client *ent.Client) error {
	defaults := []struct {
		key         string
		value       json.RawMessage
		description string
	}{
		{
			key:         "site.profile",
			value:       json.RawMessage(`{"name":"xszs-blog","description":"个人博客","logo":"","icp":""}`),
			description: "站点基础信息",
		},
		{
			key:         "site.socialLinks",
			value:       json.RawMessage(`[]`),
			description: "社交链接",
		},
		{
			key:         "site.about",
			value:       json.RawMessage(`{"markdown":"# 关于我\n\n在这里写下你的介绍。"}`),
			description: "关于页 Markdown",
		},
		{
			key: "live.broadcast",
			value: json.RawMessage(
				`{"isLive":false,"platform":"bilibili","roomTitle":"","streamTitle":"","subtitle":"","cover":"","streamUrl":""}`,
			),
			description: "直播配置",
		},
	}

	for _, item := range defaults {
		exists, err := client.SiteSetting.Query().
			Where(entsitesetting.SettingKeyEQ(item.key)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := client.SiteSetting.Create().
			SetSettingKey(item.key).
			SetValue(item.value).
			SetDescription(item.description).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func renameLegacyColumns(ctx context.Context, db *sql.DB) error {
	legacyColumns := []struct {
		table      string
		oldColumn  string
		newColumn  string
		definition string
	}{
		{table: "dictItems", oldColumn: "code", newColumn: "value", definition: "int NOT NULL"},
		{table: "dictItems", oldColumn: "name", newColumn: "label", definition: "varchar(64) NOT NULL"},
		{table: "operationLogs", oldColumn: "action", newColumn: "actionLabel", definition: "varchar(64) NOT NULL"},
		{table: "operationLogs", oldColumn: "actionCode", newColumn: "actionValue", definition: "int NOT NULL"},
		{table: "operationLogs", oldColumn: "actionName", newColumn: "actionLabel", definition: "varchar(64) NOT NULL"},
	}

	for _, legacy := range legacyColumns {
		if err := renameColumnIfExists(ctx, db, legacy.table, legacy.oldColumn, legacy.newColumn, legacy.definition); err != nil {
			return err
		}
	}
	return nil
}

func dropLegacyColumns(ctx context.Context, db *sql.DB) error {
	legacyColumns := []struct {
		table  string
		column string
	}{
		// 勿删除 dictItems.code：旧库 rename 后需保留新的 slug 字段
		{table: "dictItems", column: "name"},
		{table: "operationLogs", column: "action"},
		{table: "operationLogs", column: "actionCode"},
		{table: "operationLogs", column: "actionName"},
	}

	for _, legacy := range legacyColumns {
		if err := dropColumnIfExists(ctx, db, legacy.table, legacy.column); err != nil {
			return err
		}
	}
	return nil
}

func renameColumnIfExists(ctx context.Context, db *sql.DB, table string, oldColumn string, newColumn string, definition string) error {
	query := fmt.Sprintf("ALTER TABLE `%s` CHANGE COLUMN `%s` `%s` %s", table, oldColumn, newColumn, definition)
	if _, err := db.ExecContext(ctx, query); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && (mysqlErr.Number == 1054 || mysqlErr.Number == 1060 || mysqlErr.Number == 1146) {
			return nil
		}
		return fmt.Errorf("rename legacy column %s.%s to %s: %w", table, oldColumn, newColumn, err)
	}
	return nil
}

func dropColumnIfExists(ctx context.Context, db *sql.DB, table string, column string) error {
	query := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, column)
	if _, err := db.ExecContext(ctx, query); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && (mysqlErr.Number == 1091 || mysqlErr.Number == 1146) {
			return nil
		}
		return fmt.Errorf("drop legacy column %s.%s: %w", table, column, err)
	}
	return nil
}
