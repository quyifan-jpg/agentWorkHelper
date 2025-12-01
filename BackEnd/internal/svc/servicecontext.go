package svc

import (
"BackEnd/internal/config"
"BackEnd/internal/middleware"
"BackEnd/internal/model"

"gorm.io/driver/mysql"
"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Jwt    *middleware.Jwt
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 连接数据库
	db, err := gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
	if err != nil {
		// 如果连接失败，尝试创建数据库
		dsnWithoutDB := "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local"
		tempDB, err := gorm.Open(mysql.Open(dsnWithoutDB), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		if err := tempDB.Exec("CREATE DATABASE IF NOT EXISTS aiworkhelper").Error; err != nil {
			panic(err)
		}
		// 重试连接
		db, err = gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
		if err != nil {
			panic(err)
		}
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
&model.User{},
		&model.Department{},
		&model.DepartmentUser{},
		&model.Todo{},
		&model.TodoRecord{},
		&model.UserTodo{},
	); err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
		Jwt:    middleware.NewJwt(c.Auth.Secret),
	}
}
