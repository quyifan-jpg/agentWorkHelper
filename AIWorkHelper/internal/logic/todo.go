/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package logic

import (
	"AIWorkHelper/internal/model"
	"AIWorkHelper/pkg/token"
	"context"
	"errors"
	"gitee.com/dn-jinmin/tlog"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/svc"
)

// Todo 待办事项业务逻辑接口，定义了待办事项相关的所有业务操作方法
type Todo interface {
	Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.TodoInfoResp, err error)   // 获取待办事项详情
	Create(ctx context.Context, req *domain.Todo) (resp *domain.IdResp, err error)            // 创建待办事项
	Edit(ctx context.Context, req *domain.Todo) (err error)                                   // 编辑待办事项
	Delete(ctx context.Context, req *domain.IdPathReq) (err error)                            // 删除待办事项
	Finish(ctx context.Context, req *domain.FinishedTodoReq) (err error)                      // 完成待办事项
	CreateRecord(ctx context.Context, req *domain.TodoRecord) (err error)                     // 创建待办操作记录
	List(ctx context.Context, req *domain.TodoListReq) (resp *domain.TodoListResp, err error) // 获取待办事项列表
}

// todo 待办事项业务逻辑的默认实现
type todo struct {
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewTodo 创建待办事项业务逻辑实例
func NewTodo(svcCtx *svc.ServiceContext) Todo {
	return &todo{
		svcCtx: svcCtx,
	}
}

// Info 获取待办事项详情，包含执行人信息和创建人信息
func (l *todo) Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.TodoInfoResp, err error) {
	t, err := l.svcCtx.TodoModel.FindOne(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	uids := make([]string, 0, len(t.Executes)) // 获取到关联的用户列表
	for i, _ := range t.Executes {
		uids = append(uids, t.Executes[i].UserId)
	}
	users, err := l.svcCtx.UserModel.ListToMaps(ctx, &domain.UserListReq{Ids: uids}) // 关联用户的用户信息
	if err != nil {
		return nil, err
	}
	userTodoDomains := make([]*domain.UserTodo, 0, len(t.Executes))
	for i, _ := range t.Executes {
		u, ok := users[t.Executes[i].UserId]
		if !ok {
			u = new(model.User)
		}
		userTodoDomains = append(userTodoDomains, t.Executes[i].ToDomain(u.Name))
	}

	if time.Now().Unix() > t.DeadlineAt { // 检查是否超时
		t.TodoStatus = model.TodoTimeout
	}

	creator, ok := users[t.CreatorId] // 创建人
	if !ok {
		return nil, errors.New("用户信息查询失败")
	}

	return &domain.TodoInfoResp{
		ID:          t.ID.Hex(),
		CreatorId:   t.CreatorId,
		CreatorName: creator.Name,
		Title:       t.Title,
		DeadlineAt:  t.DeadlineAt,
		Desc:        t.Desc,
		Records:     t.ToDomainTodoRecords(),
		Status:      int(t.TodoStatus),
		ExecuteIds:  userTodoDomains,
		TodoStatus:  int(t.TodoStatus),
	}, nil
}

// Create 创建待办事项，如果没有指定执行人则默认为创建人
func (l *todo) Create(ctx context.Context, req *domain.Todo) (resp *domain.IdResp, err error) {

	tlog.InfoCtx(ctx, "create todo ", req)
	uid := token.GetUId(ctx) // 获取当前用户ID
	executes := make([]*model.UserTodo, 0, len(req.ExecuteIds))
	for _, id := range req.ExecuteIds {
		executes = append(executes, &model.UserTodo{
			UserId:     id,
			TodoStatus: model.TodoPending, // 新建待办默认为"待处理"状态
		})
	}

	if len(executes) == 0 { // 如果没有指定执行人，默认为创建人
		executes = append(executes, &model.UserTodo{
			UserId:     uid,
			TodoStatus: model.TodoPending, // 新建待办默认为"待处理"状态
		})
	}

	var records []*model.TodoRecord
	copier.Copy(&records, req.Records)

	tlog.InfoCtx(ctx, "create todo insert", req)

	id := primitive.NewObjectID()
	err = l.svcCtx.TodoModel.Insert(ctx, &model.Todo{
		ID:         id,
		CreatorId:  uid,
		Title:      req.Title,
		DeadlineAt: req.DeadlineAt,
		Desc:       req.Desc,
		Records:    records,
		Executes:   executes,
		TodoStatus: model.TodoPending, // 新建待办默认为"待处理"状态
		CreateAt:   time.Now().Unix(),
		UpdateAt:   time.Now().Unix(),
	})
	if err != nil {
		return
	}

	return &domain.IdResp{
		Id: id.Hex(),
	}, nil
}

// Edit 编辑待办事项，更新标题、描述、截止时间和状态
func (l *todo) Edit(ctx context.Context, req *domain.Todo) (err error) {
	// 将字符串ID转换为ObjectID
	oid, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return errors.New("无效的待办事项ID")
	}

	// 查询原有待办事项
	todo, err := l.svcCtx.TodoModel.FindOne(ctx, req.ID)
	if err != nil {
		return err
	}

	// 更新字段
	todo.Title = req.Title
	todo.Desc = req.Desc
	todo.DeadlineAt = req.DeadlineAt
	todo.TodoStatus = model.TodoStatus(req.Status)
	todo.ID = oid

	// 调用Update方法保存
	return l.svcCtx.TodoModel.Update(ctx, todo)
}

// Delete 删除待办事项，只有创建人可以删除
func (l *todo) Delete(ctx context.Context, req *domain.IdPathReq) (err error) {
	uid := token.GetUId(ctx) // 获取当前用户ID

	todo, err := l.svcCtx.TodoModel.FindOne(ctx, req.Id)
	if err != nil {
		return err
	}

	if uid != todo.CreatorId { // 只有创建人可以删除
		return errors.New("你不能删除该待办事项")
	}

	return l.svcCtx.TodoModel.Delete(ctx, req.Id)
}

// Finish 完成待办事项，标记指定用户的待办状态为完成，如果所有执行人都完成则更新整体状态
func (l *todo) Finish(ctx context.Context, req *domain.FinishedTodoReq) (err error) {
	todo, err := l.svcCtx.TodoModel.FindOne(ctx, req.TodoId)
	if err != nil {
		return err
	}

	for i, _ := range todo.Executes { // 标记指定用户的待办状态为完成
		if todo.Executes[i].UserId != req.UserId {
			continue
		}

		todo.Executes[i].TodoStatus = model.TodoFinish
	}

	isAllFinished := true // 检查是否所有执行人都完成了
	for i, _ := range todo.Executes {
		if todo.Executes[i].TodoStatus != model.TodoFinish {
			isAllFinished = false
			break
		}
	}

	return l.svcCtx.TodoModel.UpdateFinished(ctx, todo, isAllFinished)
}

// CreateRecord 创建待办操作记录，自动设置当前用户为记录创建人
func (l *todo) CreateRecord(ctx context.Context, req *domain.TodoRecord) (err error) {

	req.UserId = token.GetUId(ctx) // 自动设置当前用户为记录创建人

	todo, err := l.svcCtx.TodoModel.FindOne(ctx, req.TodoId)
	if err != nil {
		return err
	}

	var record model.TodoRecord
	copier.Copy(&record, req)

	todo.Records = append(todo.Records, &record) // 添加记录到待办事项

	return l.svcCtx.TodoModel.Update(ctx, todo)
}

// List 获取待办事项列表，自动检查并更新超时状态
func (l *todo) List(ctx context.Context, req *domain.TodoListReq) (resp *domain.TodoListResp, err error) {
	tlog.InfoCtx(ctx, "todo list request:", req)
	data, count, err := l.svcCtx.TodoModel.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// 收集所有用户ID（包括创建人和执行人）
	userIds := make(map[string]bool)
	for i := range data {
		userIds[data[i].CreatorId] = true
		for _, exec := range data[i].Executes {
			userIds[exec.UserId] = true
		}
	}

	// 转换为数组
	userIdList := make([]string, 0, len(userIds))
	for uid := range userIds {
		userIdList = append(userIdList, uid)
	}

	// 批量查询用户信息
	users, err := l.svcCtx.UserModel.ListToMaps(ctx, &domain.UserListReq{Ids: userIdList})
	if err != nil {
		return nil, err
	}

	var todoDomains []*domain.Todo
	for i, _ := range data {
		if time.Now().Unix() > data[i].DeadlineAt { // 检查是否超时
			data[i].TodoStatus = model.TodoTimeout
		}

		todoDomain := data[i].ToDomainTodo()
		// 设置创建人名称
		if creator, ok := users[data[i].CreatorId]; ok {
			todoDomain.CreatorName = creator.Name
		}

		// 设置执行人信息（包含用户名和状态）
		executeInfos := make([]string, 0, len(data[i].Executes))
		for _, exec := range data[i].Executes {
			if user, ok := users[exec.UserId]; ok {
				executeInfos = append(executeInfos, user.Name)
			}
		}
		todoDomain.ExecuteIds = executeInfos

		todoDomains = append(todoDomains, todoDomain)
	}

	return &domain.TodoListResp{
		Count: count,
		List:  todoDomains,
	}, nil
}
