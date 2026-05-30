package dicttypes

import (
	"errors"
	"fmt"
)

var ErrUnknownDictType = errors.New("unknown dict type")

// 字典类型常量。文章分类走 categories 表，不在此注册。
const (
	Operation = "operation"
	GameGenre = "game-genre"
)

type Meta struct {
	Key         string
	Name        string
	Description string
	// WebPublic 是否提供博客端只读接口
	WebPublic bool
}

var registry = map[string]Meta{
	Operation: {
		Key:         Operation,
		Name:        "操作类型",
		Description: "后台操作日志动作，如登录、退出",
		WebPublic:   false,
	},
	GameGenre: {
		Key:         GameGenre,
		Name:        "游戏类型",
		Description: "游戏页筛选与游戏标注，选项固定由后台维护",
		WebPublic:   true,
	},
}

func List() []Meta {
	items := make([]Meta, 0, len(registry))
	for _, item := range registry {
		items = append(items, item)
	}
	return items
}

func Get(dictType string) (Meta, bool) {
	item, ok := registry[dictType]
	return item, ok
}

func MustValidate(dictType string) error {
	if _, ok := registry[dictType]; ok {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnknownDictType, dictType)
}

func IsWebPublic(dictType string) bool {
	meta, ok := registry[dictType]
	return ok && meta.WebPublic
}
