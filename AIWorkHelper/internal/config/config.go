/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package config 提供应用程序配置结构定义
package config

import "gitee.com/dn-jinmin/tlog"

// Config 应用程序配置结构
type Config struct {
	Name string // 应用程序名称
	Addr string // 服务监听地址
	Host string // 服务访问地址，用于AI工具内部调用API

	// MongoDB 数据库配置
	Mongo struct {
		User     string   // 用户名
		Password string   // 密码
		Hosts    []string // 主机地址列表
		Port     int      // 端口号
		Database string   // 数据库名
		Params   string   // 连接参数
	}

	// JWT Token 配置
	Jwt struct {
		Secret string // 签名密钥
		Expire int64  // 过期时间（秒）
	}

	Tlog struct {
		Mode  tlog.LogMod // 运行模式
		Label string      // 加载日志输出的标签
	}

	Ws struct {
		Addr string
	}

	Langchain struct {
		Url    string
		ApiKey string
	}

	// Upload 文件上传配置
	Upload struct {
		SavePath string // 文件保存路径
		Host     string // 文件访问主机地址
	}

	Redis struct {
		Addr     string
		Password string
	}
}
