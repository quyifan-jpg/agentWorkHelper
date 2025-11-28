/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotFound           = mongo.ErrNoDocuments
	ErrInvalidObjectId    = errors.New("invalid objectId")
	ErrNotUser            = errors.New("查询不到该用户")
	ErrDepartmentNotFound = errors.New("查询不到该部门")
)
