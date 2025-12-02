package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"context"
	"strconv"
)

type DepartmentLogic interface {
	Create(ctx context.Context, req *domain.CreateDepartmentReq) error
	Update(ctx context.Context, req *domain.UpdateDepartmentReq) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*domain.Department, error)
	List(ctx context.Context, req *domain.DepartmentListReq) (*domain.DepartmentListResp, error)
	AddUser(ctx context.Context, req *domain.AddDepartmentUserReq) error
	RemoveUser(ctx context.Context, deptID string, userID string) error
}

type departmentLogic struct {
	svcCtx *svc.ServiceContext
}

func NewDepartment(svcCtx *svc.ServiceContext) DepartmentLogic {
	return &departmentLogic{
		svcCtx: svcCtx,
	}
}

func (l *departmentLogic) Create(ctx context.Context, req *domain.CreateDepartmentReq) error {
	dept := &model.Department{
		Name:     req.Name,
		LeaderID: req.LeaderId,
		ParentID: req.ParentId,
	}
	return l.svcCtx.DB.WithContext(ctx).Create(dept).Error
}

func (l *departmentLogic) Update(ctx context.Context, req *domain.UpdateDepartmentReq) error {
	dept := &model.Department{}
	if err := l.svcCtx.DB.WithContext(ctx).First(dept, req.Id).Error; err != nil {
		return err
	}
	dept.Name = req.Name
	dept.LeaderID = req.LeaderId
	dept.ParentID = req.ParentId
	return l.svcCtx.DB.WithContext(ctx).Save(dept).Error
}

func (l *departmentLogic) Delete(ctx context.Context, id string) error {
	return l.svcCtx.DB.WithContext(ctx).Delete(&model.Department{}, id).Error
}

func (l *departmentLogic) Get(ctx context.Context, id string) (*domain.Department, error) {
	dept := &model.Department{}
	if err := l.svcCtx.DB.WithContext(ctx).First(dept, id).Error; err != nil {
		return nil, err
	}
	return &domain.Department{
		Id:        dept.ID,
		Name:      dept.Name,
		LeaderId:  dept.LeaderID,
		ParentId:  dept.ParentID,
		CreatedAt: dept.CreatedAt.Unix(),
	}, nil
}

func (l *departmentLogic) List(ctx context.Context, req *domain.DepartmentListReq) (*domain.DepartmentListResp, error) {
	var depts []model.Department
	db := l.svcCtx.DB.WithContext(ctx)
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if err := db.Find(&depts).Error; err != nil {
		return nil, err
	}

	list := make([]*domain.Department, 0, len(depts))
	for _, d := range depts {
		list = append(list, &domain.Department{
			Id:        d.ID,
			Name:      d.Name,
			LeaderId:  d.LeaderID,
			ParentId:  d.ParentID,
			CreatedAt: d.CreatedAt.Unix(),
		})
	}
	return &domain.DepartmentListResp{List: list}, nil
}

func (l *departmentLogic) AddUser(ctx context.Context, req *domain.AddDepartmentUserReq) error {
	// 批量插入
	var users []model.DepartmentUser
	for _, uid := range req.UserIds {
		users = append(users, model.DepartmentUser{
			DepartmentID: req.DepartmentId,
			UserID:       uid,
		})
	}
	return l.svcCtx.DB.WithContext(ctx).Create(&users).Error
}

func (l *departmentLogic) RemoveUser(ctx context.Context, deptIDStr, userIDStr string) error {
	deptID, _ := strconv.ParseUint(deptIDStr, 10, 64)
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)
	return l.svcCtx.DB.WithContext(ctx).Where("department_id = ? AND user_id = ?", deptID, userID).Delete(&model.DepartmentUser{}).Error
}
