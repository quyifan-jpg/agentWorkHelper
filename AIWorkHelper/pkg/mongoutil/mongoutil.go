/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package mongoutil 提供 MongoDB 连接和配置工具
package mongoutil

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongodbConfig MongoDB 连接配置
type MongodbConfig struct {
	User        string   // 用户名
	Password    string   // 密码
	Hosts       []string // 主机地址列表
	Port        int      // 端口号
	Database    string   // 数据库名
	Params      string   // 连接参数
	MaxPoolSize uint64   // 连接池最大连接数
}

// MongodbDatabase 创建 MongoDB 数据库连接
func MongodbDatabase(config *MongodbConfig) (*mongo.Database, error) {
	// 连接 MongoDB
	client, err := mongo.Connect(context.TODO(), config.GetApplyURI()...)
	if err != nil {
		return nil, err
	}

	// 测试连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return client.Database(config.Database), nil
}

// GetApplyURI 构建 MongoDB 连接选项
func (t *MongodbConfig) GetApplyURI() []*options.ClientOptions {
	var ops []*options.ClientOptions
	var uri string

	// 构建基础 URI
	uri = "mongodb://"

	// 添加认证信息
	if len(t.User) > 0 && len(t.Password) > 0 {
		uri = fmt.Sprintf("%v%v:%v@", uri, t.User, t.Password)
	}

	// 添加主机列表
	for index, v := range t.Hosts {
		var host string
		if t.Port != 0 {
			host += v + fmt.Sprintf(":%d", t.Port)
		} else {
			host = v
		}
		if index < len(t.Hosts)-1 {
			host += ","
		}
		uri += host
	}

	// 添加数据库名
	uri += fmt.Sprintf("/%s", t.Database)

	// 添加连接参数
	if len(t.Params) > 0 {
		uri = fmt.Sprintf("%v?%v", uri, t.Params)
	}

	// 设置 URI 选项
	ops = append(ops, options.Client().ApplyURI(uri))

	// 设置连接池大小
	if t.MaxPoolSize > 0 {
		ops = append(ops, options.Client().SetMaxPoolSize(t.MaxPoolSize))
	}

	return ops
}
