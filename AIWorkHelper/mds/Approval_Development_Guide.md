# 审批业务开发指南

本文档旨在为新手开发者提供一个清晰、分步的指南，帮助您在现有项目框架下成功开发审批业务功能。我们将遵循从接口定义到后端实现的完整流程。

---

## 第一步：API 定义 (`.api` 文件)

在项目开发中，我们首先需要明确接口的契约，包括路由、请求参数和响应结构。这有助于前后端并行开发，并确保接口的一致性。在本项目中，我们使用 `.api` 文件来定义接口。

### 1.1 文件位置

审批业务的 API 定义文件位于 `doc/approval.api`。

### 1.2 定义数据结构 (`type`)

在 `.api` 文件中，我们使用 `type` 关键字来定义所有与审批相关的请求和响应数据结构。这些结构体最终会自动生成为 Go 代码中的 `domain` 层代码。

**关键数据结构示例：**

- **`Approval`**: 定义了创建审批申请时需要传入的核心数据，包含了请假(`Leave`)、外出(`GoOut`)、补卡(`MakeCard`)等不同类型的嵌套结构体。

  ```api
  type Approval {
      Id       string         `json:"id,omitempty"`
      UserId   string         `json:"userId,omitempty"`
      Type     int            `json:"type,omitempty"` // 审批类型：1=请假 2=外出 3=补卡
      // ... 其他字段
      Leave    *Leave         `json:"leave,omitempty"` // 请假详情
      GoOut    *GoOut         `json:"goOut,omitempty"`   // 外出详情
      MakeCard *MakeCard      `json:"makeCard,omitempty"`// 补卡详情
  }
  ```

- **`ApprovalInfoResp`**: 定义了获取审批详情接口的响应结构，包含了申请人、审批人、审批状态等完整信息。

- **`DisposeReq`**: 定义了处理审批（通过/拒绝/撤销）时需要传入的请求参数。

- **`ApprovalListReq` / `ApprovalListResp`**: 分别定义了获取审批列表的请求参数（如分页）和响应结构。

### 1.3 定义服务接口 (`service`)

在定义完数据结构后，我们使用 `service` 关键字来定义具体的 HTTP 接口。

**核心概念：**

- **`@server` 注解**: 用于配置路由、中间件和业务逻辑层(logic)的对应关系。
- **`group`**: 定义路由前缀，例如 `v1/approval`。
- **`middleware`**: 指定需要使用的中间件，例如 `Jwt` 用于身份验证。
- **`handler`**: 指定该接口对应的 `handler` 层方法。
- **`logic`**: 指定该接口对应的 `logic` 层方法。

**审批服务定义示例：**

```api
@server(
    middleware: Jwt
    group: v1/approval
    logic: Approval
)
service Approval {
    // 获取审批详情
    @server(
        handler: Info
        logic: Approval.Info
    )
    get /:id(IdPathReq) returns (ApprovalInfoResp)

    // 创建审批申请
    @server(
        handler: Create
        logic: Approval.Create
    )
    post / (Approval) returns (IdResp)

    // 处理审批
    @server(
        handler: Dispose
        logic: Approval.Dispose
    )
    put /dispose (DisposeReq)

    // 获取审批列表
    @server(
        handler: List
        logic: Approval.List
    )
    get /list (ApprovalListReq) returns(ApprovalListResp)
}
```

### 1.4 代码生成

完成 `.api` 文件的定义后，项目框架会根据该文件自动生成 `handler`、`logic`、`domain` 等目录下的基础代码框架，我们后续只需要在生成的文件中填充具体的业务逻辑即可。


---

## 第二步：Model 层开发 (数据模型与数据库操作)

Model 层是数据持久化的核心，负责定义数据库中的数据结构以及提供对这些数据的增删改查（CRUD）操作。

### 2.1 文件结构

审批业务的 Model 层主要包含两个文件：

- **`internal/model/approvaltypes.go`**: 定义与数据库集合（Collection）对应的 Go 结构体，以及相关的常量和枚举。
- **`internal/model/approvalmodel.go`**: 实现对数据库进行具体操作的方法。

### 2.2 定义数据结构 (`approvaltypes.go`)

此文件中的结构体是数据库文档（Document）的直接映射。

**核心要点：**

- **`bson` 标签**: 用于指定字段在 MongoDB 中存储的名称，例如 `bson:"_id,omitempty"`。
- **常量定义**: 使用 `const` 和 `iota` 定义审批类型（`ApprovalType`）、审批状态（`ApprovalStatus`）等枚举值，使代码更具可读性和可维护性。
- **模型转换方法**: 提供将数据库模型（`model.Approval`）转换为域模型（`domain.ApprovalInfoResp`）的方法，实现数据在不同层之间的解耦。

**`Approval` 结构体示例：**

```go
// internal/model/approvaltypes.go

type (
	// Approval 审批数据模型
	Approval struct {
		ID primitive.ObjectID `bson:"_id,omitempty"` // 数据库ID

		UserId   string         `bson:"userId,omitempty"`     // 申请人用户ID
		No       string         `bson:"no,omitempty"`             // 审批编号
		Type     ApprovalType   `bson:"type,omitempty"`         // 审批类型
		Status   ApprovalStatus `bson:"status,omitempty"`       // 审批状态
		// ... 其他字段
	}

	// Approver 审批人数据模型
	Approver struct {
		UserId   string         `bson:"userId,omitempty"`   // 用户ID
		UserName string         `bson:"userName,omitempty"` // 用户姓名
		Status   ApprovalStatus `bson:"status,omitempty"`  // 审批状态
		Reason   string         `bson:"reason,omitempty"`   // 审批理由
	}
)

// ToDomainApprovalInfo 将数据库模型转换为审批详情响应模型
func (m *Approval) ToDomainApprovalInfo() *domain.ApprovalInfoResp {
    // ... 转换逻辑
}
```

### 2.3 实现数据库操作 (`approvalmodel.go`)

此文件负责实现对 `approval` 集合的 CRUD 操作。

**核心步骤：**

1.  **定义接口 (`ApprovalModel`)**: 首先，定义一个包含所有数据库操作方法的接口，以实现依赖倒置。

    ```go
    // internal/model/approvalmodel.go

    type ApprovalModel interface {
        List(ctx context.Context, req *domain.ApprovalListReq) ([]*Approval, int64, error)
        Insert(ctx context.Context, data *Approval) error
        FindOne(ctx context.Context, id string) (*Approval, error)
        Update(ctx context.Context, data *Approval) error
        Delete(ctx context.Context, id string) error
    }
    ```

2.  **创建实现 (`defaultApprovalModel`)**: 创建一个结构体，并实现 `ApprovalModel` 接口中定义的所有方法。

3.  **实现具体方法**: 使用 `go-mongo-driver` 库提供的方法来实现具体的数据库操作。

    - **`Insert`**: 使用 `m.col.InsertOne()` 方法插入新文档。
    - **`FindOne`**: 使用 `m.col.FindOne()` 方法根据 ID 查询单个文档。
    - **`List`**: 使用 `m.col.Find()` 和 `m.col.CountDocuments()` 方法实现带分页和过滤的列表查询。
    - **`Update`**: 使用 `m.col.UpdateOne()` 方法更新文档。
    - **`Delete`**: 使用 `m.col.DeleteOne()` 方法删除文档。

**`FindOne` 方法示例：**

```go
// internal/model/approvalmodel.go

// FindOne 根据 ID 查找单个审批记录
func (m *defaultApprovalModel) FindOne(ctx context.Context, id string) (*Approval, error) {
	// 将字符串 ID 转换为 ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data Approval
	// 执行查询
	err = m.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&data)
	switch err {
	case nil:
		return &data, nil
	case mongo.ErrNoDocuments:
		// 如果没有找到文档，返回自定义的 ErrNotFound 错误
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
```


---

## 第三步：Logic 层开发 (核心业务逻辑)

Logic 层是业务处理的核心，它负责编排和组合 Model 层的原子操作，以实现复杂的业务功能。所有的业务规则、计算和流程控制都应在这一层实现。

### 3.1 文件位置

审批业务的 Logic 层代码位于 `internal/logic/approval.go`。

### 3.2 接口与实现

与 Model 层类似，Logic 层也遵循接口-实现的模式。

1.  **定义接口 (`Approval`)**: 定义所有审批业务逻辑的方法。

    ```go
    // internal/logic/approval.go

    type Approval interface {
        Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.ApprovalInfoResp, err error)
        Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error)
        Dispose(ctx context.Context, req *domain.DisposeReq) (err error)
        List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error)
    }
    ```

2.  **创建实现 (`approval`)**: 创建一个包含服务上下文（`svcCtx`）的结构体，并实现 `Approval` 接口。

    ```go
    type approval struct {
        svcCtx *svc.ServiceContext
    }
    ```

### 3.3 实现核心业务逻辑

在 `approval` 结构体的方法中，我们将实现具体的业务逻辑。

#### `Create` 方法逻辑：

`Create` 方法是审批业务中最复杂的逻辑之一，其核心步骤如下：

1.  **获取用户信息**: 从上下文（`ctx`）中获取当前操作用户的 ID。
2.  **初始化审批记录**: 创建一个新的 `model.Approval` 实例，并填充基础信息（如审批编号、类型、状态等）。
3.  **处理不同类型的审批**: 使用 `switch` 语句根据审批类型（请假、外出、补卡）分别处理，填充特定于该类型的详细信息，并生成审批摘要（`abstract`）。
4.  **构建审批流程**: 这是最关键的一步。
    -   获取申请人所在的部门信息。
    -   解析该部门的父级路径，以确定审批层级。
    -   从直属上级开始，逐级向上查找各级部门的负责人（`LeaderId`）。
    -   将各级负责人依次添加到审批人列表（`Approvers`）中。
    -   设置当前审批人为第一级审批人（即直属上级）。
5.  **保存到数据库**: 调用 `l.svcCtx.ApprovalModel.Insert()` 方法将构建好的审批记录保存到数据库。

**构建审批流程伪代码：**

```go
// internal/logic/approval.go - Create() 方法

// ... (前序步骤)

// 获取申请人所在部门
dep := GetDepartmentByUserId(uid)

// 获取所有父级部门
parentDeps := GetParentDepartments(dep.ParentPath)

// 添加直属上级为第一审批人
approvals = append(approvals, dep.Leader)

// 循环添加各级父部门负责人
for _, pDep := range parentDeps {
    approvals = append(approvals, pDep.Leader)
}

// ... (保存数据库)
```

#### `Dispose` 方法逻辑：

`Dispose` 方法负责处理审批的通过、拒绝和撤销操作。

1.  **权限验证**:
    -   如果是 **撤销** 操作，验证当前用户是否为该审批的 **申请人**。
    -   如果是 **通过/拒绝** 操作，验证当前用户是否为该审批的 **当前审批人** (`approval.ApprovalId`)。
2.  **状态检查**: 检查审批单当前状态，防止对已完成（通过/拒绝/撤销）的审批进行重复操作。
3.  **处理审批流程**:
    -   如果审批 **通过** 且存在下一级审批人，则将 `ApprovalId` 更新为下一级审批人的 ID，并将 `ApprovalIdx` 加一。
    -   如果审批 **通过** 且是最后一级，或审批被 **拒绝**，则更新整个审批单的最终状态。
4.  **更新数据库**: 调用 `l.svcCtx.ApprovalModel.Update()` 方法保存更改。


---

## 第四步：Handler 层开发 (HTTP 请求处理与路由)

Handler 层是连接外部 HTTP 请求和内部业务逻辑（Logic 层）的桥梁。它的主要职责是：

1.  解析和验证 HTTP 请求参数。
2.  调用相应的 Logic 层方法来处理业务逻辑。
3.  将 Logic 层的处理结果（成功或失败）格式化为标准的 HTTP 响应返回给客户端。

### 4.1 文件位置

审批业务的 Handler 层代码位于 `internal/handler/api/approval.go`。

### 4.2 结构体与初始化

- **`Approval` 结构体**: 包含服务上下文（`svcCtx`）和 `logic.Approval` 接口，以便调用业务逻辑。
- **`NewApproval` 函数**: 用于创建 `Approval` Handler 的实例。

```go
// internal/handler/api/approval.go

// Approval 审批处理器结构体
type Approval struct {
	svcCtx   *svc.ServiceContext // 服务上下文
	approval logic.Approval      // 审批业务逻辑接口
}

// NewApproval 创建审批处理器实例
func NewApproval(svcCtx *svc.ServiceContext, approval logic.Approval) *Approval {
	// ...
}
```

### 4.3 路由注册 (`InitRegister`)

`InitRegister` 方法负责将 `.api` 文件中定义的路由和对应的 Handler 方法进行绑定。

- **`engine.Group`**: 创建一个路由组，并应用 `Jwt` 中间件，确保该组下的所有路由都需要进行身份验证。
- **`g.GET`, `g.POST`, `g.PUT`**: 将具体的 HTTP 方法和路径映射到相应的 Handler 方法（如 `h.Info`, `h.Create`）。

```go
// internal/handler/api/approval.go

// InitRegister 初始化审批相关的路由注册
func (h *Approval) InitRegister(engine *gin.Engine) {
	// 创建审批路由组，添加JWT中间件进行身份验证
	g := engine.Group("v1/approval", h.svcCtx.Jwt.Handler)
	g.GET("/:id", h.Info)        // 获取审批详情
	g.POST("", h.Create)         // 创建审批申请
	g.PUT("/dispose", h.Dispose) // 处理审批
	g.GET("/list", h.List)       // 获取审批列表
}
```

### 4.4 实现 Handler 方法

每个 Handler 方法都遵循一个标准的处理模式：

1.  **定义请求结构体**: 定义一个变量来接收请求参数。
2.  **绑定和验证参数**: 使用 `httpx.BindAndValidate(ctx, &req)` 来自动解析请求（JSON Body, Query Params, Path Params）到结构体中，并根据定义的 `validate` 标签进行参数校验。
3.  **调用 Logic 层**: 如果参数验证通过，则调用 `h.approval` 对应的方法，并将上下文（`ctx`）和请求参数传递进去。
4.  **返回响应**:
    -   如果 Logic 层返回错误，使用 `httpx.FailWithErr(ctx, err)` 返回统一的错误响应。
    -   如果 Logic 层处理成功，使用 `httpx.OkWithData(ctx, res)` 或 `httpx.Ok(ctx)` 返回成功的响应。

**`Info` 方法示例：**

```go
// internal/handler/api/approval.go

// Info 获取审批详情接口
func (h *Approval) Info(ctx *gin.Context) {
	var req domain.IdPathReq
	// 1. 绑定并验证请求参数
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 2. 调用业务逻辑获取审批详情
	res, err := h.approval.Info(ctx.Request.Context(), &req)

    // 3. 返回响应
    if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}
```


---

## 第五步：开发流程总结

通过以上步骤，我们完成了一个功能模块从接口定义到业务实现的完整开发流程。总结一下，开发一个新功能（如审批业务）的推荐步骤如下：

1.  **定义 API (`.api`)**:
    -   在 `doc/` 目录下创建或修改 `.api` 文件。
    -   定义所有需要的数据结构（`type`）。
    -   定义服务和路由（`service`），并关联 `handler` 和 `logic`。

2.  **生成代码**:
    -   运行框架提供的代码生成工具，自动创建 `handler`, `logic`, `domain` 等目录下的基础代码。

3.  **实现 Model 层**:
    -   在 `internal/model/` 目录下创建 `...types.go` 文件，定义数据库模型和常量。
    -   创建 `...model.go` 文件，实现对数据库的增删改查（CRUD）接口。

4.  **实现 Logic 层**:
    -   在 `internal/logic/` 目录下找到对应的文件。
    -   注入 `svcCtx`，通过它来调用 Model 层的数据库操作。
    -   编写核心业务逻辑，如数据处理、流程控制、权限校验等。

5.  **实现 Handler 层**:
    -   在 `internal/handler/api/` 目录下找到对应的文件。
    -   在相应的 Handler 方法中，完成参数绑定、校验，并调用 Logic 层的方法。
    -   根据 Logic 层的返回结果，向客户端返回标准格式的 JSON 响应。

6.  **注册与启动**:
    -   确保新的 Handler 已经在 `internal/svc/servicecontext.go` 和 `internal/handler/api/router.go` 中被正确初始化和注册。
    -   启动服务，进行接口测试。

遵循以上步骤，即使是新手开发者也能清晰、高效地在本项目中开发新的功能模块。

