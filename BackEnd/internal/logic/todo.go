package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"context"
	"strconv"
	"time"
)

type TodoLogic interface {
	Create(ctx context.Context, userID uint, req *domain.Todo) error
	Update(ctx context.Context, userID uint, req *domain.Todo) error
	Delete(ctx context.Context, userID uint, id string) error
	Get(ctx context.Context, userID uint, id string) (*domain.TodoInfoResp, error)
	List(ctx context.Context, userID uint, req *domain.TodoListReq) (*domain.TodoListResp, error)
	Finish(ctx context.Context, userID uint, req *domain.FinishedTodoReq) error
	CreateRecord(ctx context.Context, userID uint, req *domain.TodoRecord) error
}

type todoLogic struct {
	svcCtx *svc.ServiceContext
}

func NewTodo(svcCtx *svc.ServiceContext) TodoLogic {
	return &todoLogic{
		svcCtx: svcCtx,
	}
}

func (l *todoLogic) Create(ctx context.Context, userID uint, req *domain.Todo) error {
	todo := &model.Todo{
		CreatorID:  userID,
		Title:      req.Title,
		Desc:       req.Desc,
		DeadlineAt: time.Unix(req.DeadlineAt, 0),
		Status:     0,
	}

	tx := l.svcCtx.DB.WithContext(ctx).Begin()
	if err := tx.Create(todo).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Add executors
	if len(req.ExecuteIds) > 0 {
		var userTodos []model.UserTodo
		for _, eidStr := range req.ExecuteIds {
			eid, _ := strconv.ParseUint(eidStr, 10, 64)
			userTodos = append(userTodos, model.UserTodo{
				TodoID: todo.ID,
				UserID: uint(eid),
			})
		}
		if err := tx.Create(&userTodos).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (l *todoLogic) Update(ctx context.Context, userID uint, req *domain.Todo) error {
	todo := &model.Todo{}
	if err := l.svcCtx.DB.WithContext(ctx).First(todo, req.ID).Error; err != nil {
		return err
	}

	// Only creator can update
	// if todo.CreatorID != userID { return errors.New("permission denied") }

	todo.Title = req.Title
	todo.Desc = req.Desc
	todo.DeadlineAt = time.Unix(req.DeadlineAt, 0)

	tx := l.svcCtx.DB.WithContext(ctx).Begin()
	if err := tx.Save(todo).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update executors (simplified: delete all and re-add)
	if len(req.ExecuteIds) > 0 {
		if err := tx.Where("todo_id = ?", todo.ID).Delete(&model.UserTodo{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		var userTodos []model.UserTodo
		for _, eidStr := range req.ExecuteIds {
			eid, _ := strconv.ParseUint(eidStr, 10, 64)
			userTodos = append(userTodos, model.UserTodo{
				TodoID: todo.ID,
				UserID: uint(eid),
			})
		}
		if err := tx.Create(&userTodos).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (l *todoLogic) Delete(ctx context.Context, userID uint, id string) error {
	return l.svcCtx.DB.WithContext(ctx).Delete(&model.Todo{}, id).Error
}

func (l *todoLogic) Get(ctx context.Context, userID uint, id string) (*domain.TodoInfoResp, error) {
	todo := &model.Todo{}
	if err := l.svcCtx.DB.WithContext(ctx).Preload("Creator").Preload("Executors").Preload("Records").Preload("Records.User").First(todo, id).Error; err != nil {
		return nil, err
	}

	resp := &domain.TodoInfoResp{
		ID:          strconv.Itoa(int(todo.ID)),
		CreatorId:   strconv.Itoa(int(todo.CreatorID)),
		CreatorName: todo.Creator.Name,
		Title:       todo.Title,
		Desc:        todo.Desc,
		DeadlineAt:  todo.DeadlineAt.Unix(),
		Status:      todo.Status,
		TodoStatus:  todo.TodoStatus,
	}

	for _, rec := range todo.Records {
		resp.Records = append(resp.Records, &domain.TodoRecord{
			TodoId:   strconv.Itoa(int(rec.TodoID)),
			UserId:   strconv.Itoa(int(rec.UserID)),
			UserName: rec.User.Name,
			Content:  rec.Content,
			Image:    rec.Image,
			CreateAt: rec.CreatedAt.Unix(),
		})
	}

	for _, exec := range todo.Executors {
		resp.ExecuteIds = append(resp.ExecuteIds, &domain.UserTodo{
			UserId:   strconv.Itoa(int(exec.ID)),
			UserName: exec.Name,
		})
	}

	return resp, nil
}

func (l *todoLogic) List(ctx context.Context, userID uint, req *domain.TodoListReq) (*domain.TodoListResp, error) {
	var todos []model.Todo
	// 1. Find IDs
	var ids []uint
	db := l.svcCtx.DB.WithContext(ctx).Model(&model.Todo{}).
		Joins("LEFT JOIN user_todos ON user_todos.todo_id = todos.id").
		Where("todos.creator_id = ? OR user_todos.user_id = ?", userID, userID)

	if req.StartTime > 0 {
		db = db.Where("todos.created_at >= ?", time.Unix(req.StartTime, 0))
	}
	if req.EndTime > 0 {
		db = db.Where("todos.created_at <= ?", time.Unix(req.EndTime, 0))
	}

	if err := db.Distinct("todos.id").Pluck("todos.id", &ids).Error; err != nil {
		return nil, err
	}

	// 2. Count
	count := int64(len(ids))

	// Set default pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Count
	if limit <= 0 {
		limit = 10
	}

	// 3. Find Todos with pagination
	start := (page - 1) * limit
	end := start + limit
	if start < len(ids) {
		if end > len(ids) {
			end = len(ids)
		}
		pageIds := ids[start:end]
		if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Todo{}).
			Preload("Creator").
			Where("id IN ?", pageIds).
			Find(&todos).Error; err != nil {
			return nil, err
		}
	}

	list := make([]*domain.Todo, 0, len(todos))
	for _, t := range todos {
		list = append(list, &domain.Todo{
			ID:          strconv.Itoa(int(t.ID)),
			CreatorId:   strconv.Itoa(int(t.CreatorID)),
			CreatorName: t.Creator.Name,
			Title:       t.Title,
			Desc:        t.Desc,
			DeadlineAt:  t.DeadlineAt.Unix(),
			Status:      t.Status,
			TodoStatus:  t.TodoStatus,
		})
	}

	return &domain.TodoListResp{Count: count, List: list}, nil
}

func (l *todoLogic) Finish(ctx context.Context, userID uint, req *domain.FinishedTodoReq) error {
	// Update UserTodo status
	return l.svcCtx.DB.WithContext(ctx).Model(&model.UserTodo{}).
		Where("todo_id = ? AND user_id = ?", req.TodoId, userID).
		Update("todo_status", 1).Error
}

func (l *todoLogic) CreateRecord(ctx context.Context, userID uint, req *domain.TodoRecord) error {
	todoID, _ := strconv.ParseUint(req.TodoId, 10, 64)
	record := &model.TodoRecord{
		TodoID:  uint(todoID),
		UserID:  userID,
		Content: req.Content,
		Image:   req.Image,
	}
	return l.svcCtx.DB.WithContext(ctx).Create(record).Error
}
