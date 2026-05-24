package dao

import (
	"context"
	"errors"

	"blog-server/internal/ent"
	entdictitem "blog-server/internal/ent/dictitem"
	"blog-server/internal/model"
)

var ErrDictItemNotFound = errors.New("dict item not found")
var ErrDictItemAlreadyExists = errors.New("dict item already exists")

type DictItemDAO interface {
	Create(ctx context.Context, item model.DictItem) (*model.DictItem, error)
	Delete(ctx context.Context, id uint64) error
	FindByTypeAndValue(ctx context.Context, dictType string, value int) (*model.DictItem, error)
	List(ctx context.Context, dictType string) ([]model.DictItem, error)
	Update(ctx context.Context, id uint64, item model.DictItem) (*model.DictItem, error)
}

type EntDictItemDAO struct {
	client *ent.Client
}

func NewEntDictItemDAO(client *ent.Client) *EntDictItemDAO {
	return &EntDictItemDAO{client: client}
}

func (dao *EntDictItemDAO) Create(ctx context.Context, item model.DictItem) (*model.DictItem, error) {
	created, err := dao.client.DictItem.Create().
		SetDictType(item.DictType).
		SetValue(item.Value).
		SetLabel(item.Label).
		SetEnabled(item.Enabled).
		SetSort(item.Sort).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrDictItemAlreadyExists
		}
		return nil, err
	}
	return toDictItemModel(created), nil
}

func (dao *EntDictItemDAO) Delete(ctx context.Context, id uint64) error {
	err := dao.client.DictItem.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrDictItemNotFound
		}
		return err
	}
	return nil
}

func (dao *EntDictItemDAO) FindByTypeAndValue(ctx context.Context, dictType string, value int) (*model.DictItem, error) {
	item, err := dao.client.DictItem.Query().
		Where(
			entdictitem.DictTypeEQ(dictType),
			entdictitem.ValueEQ(value),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrDictItemNotFound
		}
		return nil, err
	}
	return toDictItemModel(item), nil
}

func (dao *EntDictItemDAO) List(ctx context.Context, dictType string) ([]model.DictItem, error) {
	query := dao.client.DictItem.Query()
	if dictType != "" {
		query.Where(entdictitem.DictTypeEQ(dictType))
	}

	items, err := query.
		Order(ent.Asc(entdictitem.FieldDictType), ent.Asc(entdictitem.FieldSort), ent.Asc(entdictitem.FieldValue)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.DictItem, 0, len(items))
	for _, item := range items {
		result = append(result, *toDictItemModel(item))
	}
	return result, nil
}

func (dao *EntDictItemDAO) Update(ctx context.Context, id uint64, item model.DictItem) (*model.DictItem, error) {
	updated, err := dao.client.DictItem.UpdateOneID(id).
		SetDictType(item.DictType).
		SetValue(item.Value).
		SetLabel(item.Label).
		SetEnabled(item.Enabled).
		SetSort(item.Sort).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrDictItemNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrDictItemAlreadyExists
		}
		return nil, err
	}
	return toDictItemModel(updated), nil
}

func toDictItemModel(item *ent.DictItem) *model.DictItem {
	return &model.DictItem{
		ID:        item.ID,
		DictType:  item.DictType,
		Value:     item.Value,
		Label:     item.Label,
		Enabled:   item.Enabled,
		Sort:      item.Sort,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
