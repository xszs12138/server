package dao

import (
	"context"

	"blog-server/internal/ent"
	entoperationlog "blog-server/internal/ent/operationlog"
	"blog-server/internal/model"
)

type OperationLogDAO interface {
	Create(ctx context.Context, log model.OperationLog) error
	List(ctx context.Context, page int, pageSize int) ([]model.OperationLog, int, error)
}

type EntOperationLogDAO struct {
	client *ent.Client
}

func NewEntOperationLogDAO(client *ent.Client) *EntOperationLogDAO {
	return &EntOperationLogDAO{client: client}
}

func (dao *EntOperationLogDAO) Create(ctx context.Context, log model.OperationLog) error {
	return dao.client.OperationLog.Create().
		SetUserId(log.UserID).
		SetUsername(log.Username).
		SetActionValue(log.ActionValue).
		SetActionLabel(log.ActionLabel).
		SetIP(log.IP).
		SetRegion(log.Region).
		SetUserAgent(log.UserAgent).
		Exec(ctx)
}

func (dao *EntOperationLogDAO) List(ctx context.Context, page int, pageSize int) ([]model.OperationLog, int, error) {
	total, err := dao.client.OperationLog.Query().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	logs, err := dao.client.OperationLog.Query().
		Order(ent.Desc(entoperationlog.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.OperationLog, 0, len(logs))
	for _, log := range logs {
		items = append(items, model.OperationLog{
			ID:          log.ID,
			UserID:      log.UserId,
			Username:    log.Username,
			ActionValue: log.ActionValue,
			ActionLabel: log.ActionLabel,
			IP:          log.IP,
			Region:      log.Region,
			UserAgent:   log.UserAgent,
			CreatedAt:   log.CreatedAt,
		})
	}
	return items, total, nil
}
