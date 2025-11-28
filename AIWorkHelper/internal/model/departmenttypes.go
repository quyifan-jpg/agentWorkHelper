/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package model

import (
	"AIWorkHelper/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

// Department 部门数据模型
type Department struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`                // 部门ID
	Name       string             `bson:"name" json:"name"`                                 // 部门名称
	ParentId   string             `bson:"parentId,omitempty" json:"parentId,omitempty"`     // 父部门ID
	ParentPath string             `bson:"parentPath,omitempty" json:"parentPath,omitempty"` // 父部门路径
	Level      int                `bson:"level" json:"level"`                               // 部门层级
	LeaderId   string             `bson:"leaderId,omitempty" json:"leaderId,omitempty"`     // 部门负责人ID
	Leader     string             `bson:"leader,omitempty" json:"leader,omitempty"`         // 部门负责人姓名
	Count      int64              `bson:"count" json:"count"`                               // 部门人数
	UpdateAt   int64              `bson:"updateAt,omitempty" json:"updateAt,omitempty"`     // 更新时间
	CreateAt   int64              `bson:"createAt,omitempty" json:"createAt,omitempty"`     // 创建时间
}

// DepartmentParentPath 构建部门父路径
// 将父路径和当前ID拼接，用冒号分隔
func DepartmentParentPath(path string, id string) string {
	return path + ":" + id
}

// ParseParentPath 解析父路径字符串
// 将路径字符串按冒号分割，返回ID数组（去掉第一个空元素）
func ParseParentPath(parentPath string) []string {
	res := strings.Split(parentPath, ":")
	return res[1:] // 去掉第一个空字符串
}

// ToDepartment 将数据模型转换为领域模型
func (d *Department) ToDepartment() *domain.Department {
	return &domain.Department{
		Id:         d.ID.Hex(),    // ObjectID转换为字符串
		Name:       d.Name,
		ParentId:   d.ParentId,
		Level:      d.Level,
		LeaderId:   d.LeaderId,
		ParentPath: d.ParentPath,
	}
}
