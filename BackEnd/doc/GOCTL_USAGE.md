# goctl 使用指南

本项目使用 `goctl` 来生成代码，提高开发效率。

## 工作流程

### 1. 定义 API（.api 文件）

在 `doc/` 目录下创建或修改 `.api` 文件：

- `base.api` - 基础数据结构（通用类型）
- `user.api` - 用户相关接口定义
- `api.api` - 主入口文件（导入其他 .api 文件）

### 2. 生成代码

使用 goctl 生成代码：

```bash
# 验证 .api 文件语法
goctl api validate --api doc/api.api

# 格式化 .api 文件
goctl api format --dir doc

# 生成代码（只生成 types 层，不生成 handler/logic）
goctl api go --api doc/api.api --dir . --style gozero
```

**注意**：goctl 默认生成 go-zero 风格的代码，但我们只使用它生成 `types` 层的数据结构。

### 3. 使用生成的 types

生成的 `internal/types/types.go` 文件包含所有请求/响应结构体，可以直接在 Handler 中使用：

```go
import "BackEnd/internal/types"

func (h *User) Register(ctx *gin.Context) {
	var req types.RegisterReq  // 使用生成的类型
	// ...
}
```

### 4. 手动编写 Handler 和 Logic

Handler 和 Logic 层需要手动编写（使用 Gin 框架），参考 `internal/handler/api/user.go` 的模式。

## 项目结构

```
BackEnd/
├── doc/                    # API 定义文件
│   ├── api.api            # 主入口
│   ├── base.api           # 基础类型
│   └── user.api           # 用户接口
├── internal/
│   ├── types/             # goctl 生成（数据结构）
│   │   └── types.go       # 自动生成，不要手动编辑
│   ├── handler/api/        # 手动编写（Gin Handler）
│   └── logic/             # 手动编写（业务逻辑）
```

## 开发新功能

1. **定义 API**：在 `doc/xxx.api` 中定义接口和数据结构
2. **生成 types**：运行 `goctl api go --api doc/api.api --dir .`
3. **编写 Handler**：在 `internal/handler/api/` 中创建 Handler
4. **编写 Logic**：在 `internal/logic/` 中实现业务逻辑
5. **注册路由**：在 `internal/handler/api/router.go` 中注册

## 注意事项

- ✅ **使用 goctl 生成**：`internal/types/types.go`（数据结构）
- ❌ **不使用 goctl 生成**：Handler 和 Logic 层（手动编写，使用 Gin）
- 🔄 **更新 API**：修改 `.api` 文件后，重新运行 goctl 生成 types

## 示例

### 定义新的 API

在 `doc/todo.api` 中：

```api
type CreateTodoReq {
    Title string `json:"title" binding:"required"`
}

@server(
    group: v1/todo
    logic: Todo
    middleware: Jwt
)
service Todo {
    @handler Create
    post /(CreateTodoReq)
}
```

### 生成代码

```bash
goctl api go --api doc/api.api --dir .
```

### 使用生成的类型

```go
import "BackEnd/internal/types"

func (h *Todo) Create(ctx *gin.Context) {
	var req types.CreateTodoReq
	// ...
}
```

