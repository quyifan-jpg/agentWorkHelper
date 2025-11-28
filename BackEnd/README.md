# AIWorkHelper Backend

这是一个使用 Go 语言开发的 AIWorkHelper 后端服务，基于 **Gin + GORM** 的清晰分层架构。

## 前置要求

1. **Go 1.24+** - [安装 Go](https://golang.org/dl/)
2. **MySQL 5.7+** - 确保 MySQL 服务正在运行
3. **swag 工具** - 用于生成 Swagger 文档（可选，如果已生成可跳过）

## 快速开始

### 1. 安装依赖

```bash
cd BackEnd
go mod download
```

### 2. 配置数据库

编辑 `etc/backend.yaml` 文件，修改 MySQL 连接信息：

```yaml
MySQL:
  DSN: "root:root@tcp(127.0.0.1:3306)/aiworkhelper?charset=utf8mb4&parseTime=True&loc=Local"
```

根据你的实际情况修改：
- `root:root` - 数据库用户名和密码
- `127.0.0.1:3306` - 数据库地址和端口
- `aiworkhelper` - 数据库名称（如果不存在，程序会自动创建）

### 3. 生成 Swagger 文档（如果需要更新）

如果修改了 API 注释，需要重新生成 Swagger 文档：

```bash
# 安装 swag 工具（如果还没安装）
go install github.com/swaggo/swag/cmd/swag@latest

# 生成 Swagger 文档
swag init -g cmd/api/main.go
```

### 4. 启动服务

```bash
# 方式1: 直接运行
go run cmd/api/main.go

# 方式2: 指定配置文件路径
go run cmd/api/main.go -f etc/backend.yaml

# 方式3: 编译后运行
go build -o backend cmd/api/main.go
./backend -f etc/backend.yaml
```

服务启动后，你会看到：
```
Starting server at 0.0.0.0:8889...
```

## API 测试

### 使用 Swagger UI 测试

1. 启动服务后，在浏览器中访问：
   ```
   http://localhost:8889/swagger/index.html
   ```

2. 在 Swagger UI 中你可以：
   - 查看所有可用的 API 端点
   - 查看请求/响应格式
   - 直接在页面上测试 API

### 使用 curl 测试

#### 1. 测试健康检查
```bash
curl http://localhost:8889/ping
```

#### 2. 用户注册
```bash
curl -X POST http://localhost:8889/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

#### 3. 用户登录
```bash
curl -X POST http://localhost:8889/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

登录成功后会返回 JWT token，格式如下：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 4. 获取用户信息（需要认证）
```bash
curl -X GET http://localhost:8889/v1/user/info \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 5. 更新用户资料（需要认证）
```bash
curl -X PUT http://localhost:8889/v1/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "newname"
  }'
```

## API 端点

### 公开接口（无需认证）
- `GET /ping` - 健康检查
- `POST /v1/user/register` - 用户注册
- `POST /v1/user/login` - 用户登录

### 需要认证的接口（需要 JWT Token）
- `GET /v1/user/info` - 获取用户信息
- `PUT /v1/user/profile` - 更新用户资料

### 文档
- `GET /swagger/*any` - Swagger UI 文档

**注意**: 需要认证的接口需要在请求头中添加 `Authorization: Bearer YOUR_JWT_TOKEN`

## 项目结构

```
BackEnd/
├── cmd/
│   └── api/
│       └── main.go              # 应用入口
├── etc/
│   └── backend.yaml            # 配置文件
├── internal/
│   ├── config/                 # 配置结构
│   ├── handler/                # HTTP 处理器层
│   │   ├── api/                # API Handler（统一接口模式）
│   │   │   ├── handler.go     # Handler 管理器
│   │   │   ├── router.go      # Handler 注册
│   │   │   └── user.go        # 用户相关 Handler
│   │   └── result.go          # 统一错误处理
│   ├── middleware/            # 中间件
│   │   ├── jwt.go             # JWT 认证中间件
│   │   └── log.go             # 日志中间件
│   ├── logic/                  # 业务逻辑层
│   │   └── user.go            # 用户业务逻辑
│   ├── model/                  # 数据模型（GORM）
│   │   └── user.go
│   └── svc/                    # 服务上下文（依赖注入）
│       └── servicecontext.go
├── pkg/                        # 公共包
│   ├── httpx/                  # HTTP 工具
│   │   ├── request.go         # 请求绑定和验证
│   │   └── response.go        # 统一响应格式
│   ├── jwt/                    # JWT 工具
│   │   └── jwt.go
│   └── token/                  # Token 上下文工具
│       └── token.go
├── docs/                       # Swagger 文档（自动生成）
└── go.mod                      # Go 模块依赖
```

## 架构说明

项目采用清晰的分层架构：

1. **Handler 层** (`internal/handler/api/`): 处理 HTTP 请求，参数验证，调用 Logic 层
2. **Logic 层** (`internal/logic/`): 业务逻辑处理，数据校验
3. **Model 层** (`internal/model/`): GORM 数据模型定义
4. **ServiceContext** (`internal/svc/`): 依赖注入，统一管理数据库、配置等

### 特性

- ✅ **统一 Handler 接口**: 所有 Handler 实现 `InitRegister` 接口，便于管理
- ✅ **统一响应格式**: 所有 API 返回统一的 JSON 格式
- ✅ **JWT 认证中间件**: 保护需要认证的路由
- ✅ **日志中间件**: 自动记录请求日志
- ✅ **参数验证**: 使用 `validator` 进行请求参数验证

## 常见问题

### 1. 数据库连接失败

确保：
- MySQL 服务正在运行
- 配置文件中的数据库连接信息正确
- 数据库用户有创建数据库的权限

### 2. Swagger 页面无法访问

确保：
- 服务已成功启动
- 访问的 URL 正确：`http://localhost:8889/swagger/index.html`
- 如果文档未生成，运行 `swag init -g cmd/api/main.go`

### 3. 端口被占用

修改 `etc/backend.yaml` 中的 `Port` 字段为其他端口号。

## 开发建议

1. **修改 API 后**：记得更新 Swagger 注释并重新生成文档
2. **数据库迁移**：当前使用 GORM 的 AutoMigrate，生产环境建议使用迁移工具
3. **配置管理**：敏感信息（如数据库密码）建议使用环境变量

## 下一步

- [ ] 添加更多 API 端点
- [ ] 实现 JWT 中间件验证
- [ ] 添加日志记录
- [ ] 添加单元测试
- [ ] 配置生产环境部署

