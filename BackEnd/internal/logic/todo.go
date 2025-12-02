package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/xerr"
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type TodoLogic interface {
	Create(ctx context.Context, userID uint, req *domain.Todo) (*domain.IdResp, error)
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

func (l *todoLogic) Create(ctx context.Context, userID uint, req *domain.Todo) (*domain.IdResp, error) {
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
		log.Error().Err(err).Msg("failed to create todo")
		return nil, xerr.New(err)
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
			log.Error().Err(err).Msg("failed to create todo executors")
			return nil, xerr.New(err)
		}
	} else {
		// Guide says: "System automatically sets creator as executor"
		// If executeIds is empty, add creator?
		// Let's check guide: "👥 执行人: 自动添加创建者为执行人"
		// My current logic doesn't do this if executeIds is empty.
		// I should add creator as executor if list is empty, or maybe always?
		// Guide says: "executeIds: []" in request, and response has creator as executor.
		// So I should add creator.
		userTodos := []model.UserTodo{
			{
				TodoID:     todo.ID,
				UserID:     userID,
				TodoStatus: 1, // InProgress
			},
		}
		if err := tx.Create(&userTodos).Error; err != nil {
			tx.Rollback()
			return nil, xerr.New(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, xerr.New(err)
	}
	return &domain.IdResp{Id: strconv.Itoa(int(todo.ID))}, nil
}

func (l *todoLogic) Update(ctx context.Context, userID uint, req *domain.Todo) error {
	todo := &model.Todo{}
	if err := l.svcCtx.DB.WithContext(ctx).First(todo, req.ID).Error; err != nil {
		return xerr.New(err)
	}

	// Only creator can update
	// if todo.CreatorID != userID { return errors.New("permission denied") }

	todo.Title = req.Title
	todo.Desc = req.Desc
	todo.DeadlineAt = time.Unix(req.DeadlineAt, 0)

	tx := l.svcCtx.DB.WithContext(ctx).Begin()
	if err := tx.Save(todo).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("failed to update todo")
		return xerr.New(err)
	}

	// Update executors
	if len(req.ExecuteIds) > 0 {
		// 1. Get existing executors to preserve status
		var existing []model.UserTodo
		if err := tx.Where("todo_id = ?", todo.ID).Find(&existing).Error; err != nil {
			tx.Rollback()
			return xerr.New(err)
		}
		statusMap := make(map[uint]int)
		for _, e := range existing {
			statusMap[e.UserID] = e.TodoStatus
		}

		// 2. Delete all
		if err := tx.Where("todo_id = ?", todo.ID).Delete(&model.UserTodo{}).Error; err != nil {
			tx.Rollback()
			return xerr.New(err)
		}

		// 3. Re-add with preserved status
		var userTodos []model.UserTodo
		for _, eidStr := range req.ExecuteIds {
			eid, _ := strconv.ParseUint(eidStr, 10, 64)
			uid := uint(eid)
			status := 0 // Default pending
			if s, ok := statusMap[uid]; ok {
				status = s
			}
			userTodos = append(userTodos, model.UserTodo{
				TodoID:     todo.ID,
				UserID:     uid,
				TodoStatus: status,
			})
		}
		if err := tx.Create(&userTodos).Error; err != nil {
			tx.Rollback()
			log.Error().Err(err).Msg("failed to update todo executors")
			return xerr.New(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return xerr.New(err)
	}
	return nil
}

func (l *todoLogic) Delete(ctx context.Context, userID uint, id string) error {
	if err := l.svcCtx.DB.WithContext(ctx).Delete(&model.Todo{}, id).Error; err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to delete todo")
		return xerr.New(err)
	}
	return nil
}

func (l *todoLogic) Get(ctx context.Context, userID uint, id string) (*domain.TodoInfoResp, error) {
	todo := &model.Todo{}
	if err := l.svcCtx.DB.WithContext(ctx).Preload("Creator").Preload("Executors").Preload("Records").Preload("Records.User").First(todo, id).Error; err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to find todo info")
		return nil, xerr.New(err)
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
		log.Error().Err(err).Msg("failed to find todo ids")
		return nil, xerr.New(err)
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
			log.Error().Err(err).Msg("failed to find todos")
			return nil, xerr.New(err)
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
	// 1. Update UserTodo status
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.UserTodo{}).
		Where("todo_id = ? AND user_id = ?", req.TodoId, userID).
		Update("todo_status", 2).Error; err != nil {
		log.Error().Err(err).Msg("failed to update todo status")
		return err
	}

	// 2. Check if all executors finished
	var count int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.UserTodo{}).
		Where("todo_id = ? AND todo_status != 2", req.TodoId).
		Count(&count).Error; err != nil {
		return xerr.New(err)
	}

	// 3. If all finished (count == 0), update Todo status
	if count == 0 {
		if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Todo{}).
			Where("id = ?", req.TodoId).
			Update("todo_status", 2).Error; err != nil {
			log.Error().Err(err).Msg("failed to update todo overall status")
			return xerr.New(err)
		}
	}

	return nil
}

func (l *todoLogic) CreateRecord(ctx context.Context, userID uint, req *domain.TodoRecord) error {
	todoID, _ := strconv.ParseUint(req.TodoId, 10, 64)
	record := &model.TodoRecord{
		TodoID:  uint(todoID),
		UserID:  userID,
		Content: req.Content,
		Image:   req.Image,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(record).Error; err != nil {
		log.Error().Err(err).Msg("failed to create todo record")
		return xerr.New(err)
	}
	return nil
}
