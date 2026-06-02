package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"blog-server/internal/dicttypes"
	"blog-server/internal/ent"
	entdictitem "blog-server/internal/ent/dictitem"

	"github.com/go-sql-driver/mysql"
)

func ensureDictItemsCodeColumn(ctx context.Context, db *sql.DB) error {
	if err := addColumnIfNotExists(
		ctx,
		db,
		"dictItems",
		"code",
		"varchar(64) NULL",
	); err != nil {
		return err
	}
	return addUniqueIndexIfNotExists(
		ctx,
		db,
		"dictItems",
		"ukDictItemsTypeCode",
		"`dictType`, `code`",
	)
}

func addColumnIfNotExists(
	ctx context.Context,
	db *sql.DB,
	table string,
	column string,
	definition string,
) error {
	query := fmt.Sprintf(
		"ALTER TABLE `%s` ADD COLUMN `%s` %s",
		table,
		column,
		definition,
	)
	if _, err := db.ExecContext(ctx, query); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && (mysqlErr.Number == 1060 || mysqlErr.Number == 1146) {
			return nil
		}
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}

func addUniqueIndexIfNotExists(
	ctx context.Context,
	db *sql.DB,
	table string,
	indexName string,
	columns string,
) error {
	query := fmt.Sprintf(
		"CREATE UNIQUE INDEX `%s` ON `%s` (%s)",
		indexName,
		table,
		columns,
	)
	if _, err := db.ExecContext(ctx, query); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && (mysqlErr.Number == 1061 || mysqlErr.Number == 1062 || mysqlErr.Number == 1146) {
			return nil
		}
		return fmt.Errorf("create index %s on %s: %w", indexName, table, err)
	}
	return nil
}

func backfillDictItemCodes(ctx context.Context, client *ent.Client) error {
	backfills := []struct {
		dictType string
		value    int
		code     string
	}{
		{dicttypes.Operation, 1, "login"},
		{dicttypes.Operation, 2, "logout"},
		{dicttypes.GameGenre, 1, "rpg"},
		{dicttypes.GameGenre, 2, "action"},
		{dicttypes.GameGenre, 3, "adventure"},
		{dicttypes.GameGenre, 4, "indie"},
		{dicttypes.GameGenre, 5, "strategy"},
		{dicttypes.GameGenre, 6, "online"},
	}

	for _, item := range backfills {
		_, err := client.DictItem.Update().
			Where(
				entdictitem.DictTypeEQ(item.dictType),
				entdictitem.ValueEQ(item.value),
			).
			SetCode(item.code).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
