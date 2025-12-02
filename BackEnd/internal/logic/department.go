package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"context"
	"strconv"

	"gorm.io/gorm"
)

type DepartmentLogic interface {
	Soa(ctx context.Context, req *domain.DepartmentListReq) (*domain.DepartmentSoaResp, error)
	Info(ctx context.Context, req *domain.IdPathReq) (*domain.Department, error)
	Create(ctx context.Context, req *domain.Department) error
	Edit(ctx context.Context, req *domain.Department) error
	Delete(ctx context.Context, req *domain.IdPathReq) error
	SetDepartmentUsers(ctx context.Context, req *domain.SetDepartmentUser) error
	AddDepartmentUser(ctx context.Context, req *domain.AddDepartmentUser) error
	RemoveDepartmentUser(ctx context.Context, req *domain.RemoveDepartmentUser) error
	DepartmentUserInfo(ctx context.Context, req *domain.IdPathReq) (*domain.Department, error)
}

type departmentLogic struct {
	svcCtx *svc.ServiceContext
}

func NewDepartment(svcCtx *svc.ServiceContext) DepartmentLogic {
	return &departmentLogic{
		svcCtx: svcCtx,
	}
}

func (l *departmentLogic) Soa(ctx context.Context, req *domain.DepartmentListReq) (*domain.DepartmentSoaResp, error) {
	var depts []model.Department
	db := l.svcCtx.DB.WithContext(ctx)

	if req.DepId != "" {
		id, _ := strconv.Atoi(req.DepId)
		if id > 0 {
			db = db.Where("id = ?", id)
		}
	}
	if len(req.DepIds) > 0 {
		var ids []int
		for _, idStr := range req.DepIds {
			id, _ := strconv.Atoi(idStr)
			if id > 0 {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			db = db.Where("id IN ?", ids)
		}
	}

	if err := db.Find(&depts).Error; err != nil {
		return nil, err
	}

	// 1. Convert to domain objects and store in map
	deptMap := make(map[string]*domain.Department)
	for _, d := range depts {
		deptMap[strconv.Itoa(int(d.ID))] = &domain.Department{
			Id:         strconv.Itoa(int(d.ID)),
			Name:       d.Name,
			LeaderId:   strconv.Itoa(int(d.LeaderID)),
			ParentId:   strconv.Itoa(int(d.ParentID)),
			ParentPath: d.ParentPath,
			Level:      d.Level,
			Leader:     d.Leader,
			Child:      make([]*domain.Department, 0),
		}
	}

	// 2. Build tree
	var rootDepts []*domain.Department
	for _, d := range deptMap {
		// If parent exists in map, add to parent's child list
		if parent, ok := deptMap[d.ParentId]; ok {
			parent.Child = append(parent.Child, d)
		} else {
			// Otherwise it's a root node (in the context of this query)
			rootDepts = append(rootDepts, d)
		}
	}

	// 3. Construct response
	// If we have a single root and it matches the requested DepId, we might want to return that.
	// But DepartmentSoaResp structure suggests it wraps the result.
	// We will put all root nodes into the Child of the response, or if there is only one root, maybe map it directly?
	// Given DepartmentSoaResp has Id, Name etc, it seems to be a Department node itself.
	// Let's create a virtual root or return the first root if it makes sense.

	// Strategy: Return a "virtual" root containing all found roots as children.
	resp := &domain.DepartmentSoaResp{
		Child: rootDepts,
		Count: int64(len(depts)),
	}

	return resp, nil
}

func (l *departmentLogic) Info(ctx context.Context, req *domain.IdPathReq) (*domain.Department, error) {
	dept := &model.Department{}
	if err := l.svcCtx.DB.WithContext(ctx).First(dept, req.Id).Error; err != nil {
		return nil, err
	}
	return &domain.Department{
		Id:       strconv.Itoa(int(dept.ID)),
		Name:     dept.Name,
		LeaderId: strconv.Itoa(int(dept.LeaderID)),
		ParentId: strconv.Itoa(int(dept.ParentID)),
		// CreatedAt: dept.CreatedAt.Unix(), // API definition doesn't have CreatedAt anymore? Check .api
	}, nil
}

func (l *departmentLogic) Create(ctx context.Context, req *domain.Department) error {
	leaderID, _ := strconv.Atoi(req.LeaderId)
	parentID, _ := strconv.Atoi(req.ParentId)

	dept := &model.Department{
		Name:     req.Name,
		LeaderID: uint(leaderID),
		ParentID: uint(parentID),
	}
	return l.svcCtx.DB.WithContext(ctx).Create(dept).Error
}

func (l *departmentLogic) Edit(ctx context.Context, req *domain.Department) error {
	dept := &model.Department{}
	if err := l.svcCtx.DB.WithContext(ctx).First(dept, req.Id).Error; err != nil {
		return err
	}
	leaderID, _ := strconv.Atoi(req.LeaderId)
	parentID, _ := strconv.Atoi(req.ParentId)

	dept.Name = req.Name
	dept.LeaderID = uint(leaderID)
	dept.ParentID = uint(parentID)
	return l.svcCtx.DB.WithContext(ctx).Save(dept).Error
}

func (l *departmentLogic) Delete(ctx context.Context, req *domain.IdPathReq) error {
	return l.svcCtx.DB.WithContext(ctx).Delete(&model.Department{}, req.Id).Error
}

func (l *departmentLogic) SetDepartmentUsers(ctx context.Context, req *domain.SetDepartmentUser) error {
	// Transaction to replace users
	return l.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deptID, _ := strconv.Atoi(req.DepId)
		// Delete existing
		if err := tx.Where("department_id = ?", deptID).Delete(&model.DepartmentUser{}).Error; err != nil {
			return err
		}
		// Add new
		var users []model.DepartmentUser
		for _, uidStr := range req.UserIds {
			uid, _ := strconv.Atoi(uidStr)
			users = append(users, model.DepartmentUser{
				DepartmentID: uint(deptID),
				UserID:       uint(uid),
			})
		}
		if len(users) > 0 {
			if err := tx.Create(&users).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *departmentLogic) AddDepartmentUser(ctx context.Context, req *domain.AddDepartmentUser) error {
	deptID, _ := strconv.Atoi(req.DepId)
	uid, _ := strconv.Atoi(req.UserId)
	user := model.DepartmentUser{
		DepartmentID: uint(deptID),
		UserID:       uint(uid),
	}
	return l.svcCtx.DB.WithContext(ctx).Create(&user).Error
}

func (l *departmentLogic) RemoveDepartmentUser(ctx context.Context, req *domain.RemoveDepartmentUser) error {
	deptID, _ := strconv.Atoi(req.DepId)
	uid, _ := strconv.Atoi(req.UserId)
	return l.svcCtx.DB.WithContext(ctx).Where("department_id = ? AND user_id = ?", deptID, uid).Delete(&model.DepartmentUser{}).Error
}

func (l *departmentLogic) DepartmentUserInfo(ctx context.Context, req *domain.IdPathReq) (*domain.Department, error) {
	// Find department for user
	var deptUser model.DepartmentUser
	if err := l.svcCtx.DB.WithContext(ctx).Where("user_id = ?", req.Id).First(&deptUser).Error; err != nil {
		return nil, err
	}

	var dept model.Department
	if err := l.svcCtx.DB.WithContext(ctx).First(&dept, deptUser.DepartmentID).Error; err != nil {
		return nil, err
	}

	return &domain.Department{
		Id:       strconv.Itoa(int(dept.ID)),
		Name:     dept.Name,
		LeaderId: strconv.Itoa(int(dept.LeaderID)),
		ParentId: strconv.Itoa(int(dept.ParentID)),
	}, nil
}
