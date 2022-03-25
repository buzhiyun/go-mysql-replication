package model

import "github.com/go-mysql-org/go-mysql/schema"

// 表结构信息
type TableInfomation struct {
	TableInfo       *schema.Table
	TableColumnSize int
}

var TableInfo = make(map[string]TableInfomation)
