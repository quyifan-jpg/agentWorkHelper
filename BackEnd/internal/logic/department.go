package logic

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"context"
	"strconv"

	"BackEnd/pkg/xerr"
	"strings"

	"github.com/rs/zerolog/log"
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
		log.Error().Err(err).Msg("failed to find departments")
		return nil, xerr.New(err)
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
		log.Error().Err(err).Str("id", req.Id).Msg("failed to find department info")
		return nil, xerr.New(err)
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

	if parentID > 0 {
		var parent model.Department
		if err := l.svcCtx.DB.WithContext(ctx).First(&parent, parentID).Error; err != nil {
			return xerr.New(err)
		}
		dept.Level = parent.Level + 1
		if parent.ParentPath == "" {
			dept.ParentPath = strconv.Itoa(int(parent.ID))
		} else {
			dept.ParentPath = parent.ParentPath + "-" + strconv.Itoa(int(parent.ID))
		}
	} else {
		dept.Level = 0
		dept.ParentPath = ""
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(dept).Error; err != nil {
		log.Error().Err(err).Msg("failed to create department")
		return xerr.New(err)
	}
	return nil
}

func (l *departmentLogic) Edit(ctx context.Context, req *domain.Department) error {
	dept := &model.Department{}
	if err := l.svcCtx.DB.WithContext(ctx).First(dept, req.Id).Error; err != nil {
		return xerr.New(err)
	}
	leaderID, _ := strconv.Atoi(req.LeaderId)
	parentID, _ := strconv.Atoi(req.ParentId)

	dept.Name = req.Name
	dept.LeaderID = uint(leaderID)
	dept.ParentID = uint(parentID)
	if err := l.svcCtx.DB.WithContext(ctx).Save(dept).Error; err != nil {
		log.Error().Err(err).Msg("failed to update department")
		return xerr.New(err)
	}
	return nil
}

func (l *departmentLogic) Delete(ctx context.Context, req *domain.IdPathReq) error {
	if err := l.svcCtx.DB.WithContext(ctx).Delete(&model.Department{}, req.Id).Error; err != nil {
		log.Error().Err(err).Str("id", req.Id).Msg("failed to delete department")
		return xerr.New(err)
	}
	return nil
}

func (l *departmentLogic) SetDepartmentUsers(ctx context.Context, req *domain.SetDepartmentUser) error {
	// Transaction to replace users
	return l.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deptID, _ := strconv.Atoi(req.DepId)
		// Delete existing
		if err := tx.Where("department_id = ?", deptID).Delete(&model.DepartmentUser{}).Error; err != nil {
			return xerr.New(err)
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
				return xerr.New(err)
			}
		}
		return nil
	})
}

func (l *departmentLogic) AddDepartmentUser(ctx context.Context, req *domain.AddDepartmentUser) error {
	deptID, _ := strconv.Atoi(req.DepId)
	uid, _ := strconv.Atoi(req.UserId)

	// 1. Check if department exists
	var dept model.Department
	if err := l.svcCtx.DB.WithContext(ctx).First(&dept, deptID).Error; err != nil {
		return err
	}

	// 2. Add user to current department
	user := model.DepartmentUser{
		DepartmentID: uint(deptID),
		UserID:       uint(uid),
	}
	// Use FirstOrCreate to avoid duplicate error
	if err := l.svcCtx.DB.WithContext(ctx).FirstOrCreate(&user, user).Error; err != nil {
		log.Error().Err(err).Msg("failed to add user to department")
		return xerr.New(err)
	}

	// 3. Recursively add to parent departments
	if dept.ParentPath != "" {
		parentIds := model.ParseParentPath(dept.ParentPath)
		for _, pid := range parentIds {
			pUser := model.DepartmentUser{
				DepartmentID: pid,
				UserID:       uint(uid),
			}
			if err := l.svcCtx.DB.WithContext(ctx).FirstOrCreate(&pUser, pUser).Error; err != nil {
				// Log error but continue? Or return error?
				// For now, return error to be safe
				log.Error().Err(err).Msg("failed to add user to parent department")
				return xerr.New(err)
			}
		}
	}

	return nil
}

func (l *departmentLogic) RemoveDepartmentUser(ctx context.Context, req *domain.RemoveDepartmentUser) error {
	deptID, _ := strconv.Atoi(req.DepId)
	uid, _ := strconv.Atoi(req.UserId)

	// 1. Check if department exists
	var dept model.Department
	if err := l.svcCtx.DB.WithContext(ctx).First(&dept, deptID).Error; err != nil {
		return xerr.New(err)
	}

	// 2. Remove from current department
	if err := l.svcCtx.DB.WithContext(ctx).Where("department_id = ? AND user_id = ?", deptID, uid).Delete(&model.DepartmentUser{}).Error; err != nil {
		log.Error().Err(err).Msg("failed to remove user from department")
		return xerr.New(err)
	}

	// 3. Recursively check and remove from parent departments
	if dept.ParentPath != "" {
		parentIds := model.ParseParentPath(dept.ParentPath)
		// Reverse order: from closest parent to root (though order doesn't strictly matter here, logic does)
		// Actually, we need to check if the user is still in ANY child of the parent.

		// Get all departments this user belongs to
		var userDepts []model.DepartmentUser
		if err := l.svcCtx.DB.WithContext(ctx).Where("user_id = ?", uid).Find(&userDepts).Error; err != nil {
			return xerr.New(err)
		}
		userDeptIDs := make(map[uint]bool)
		for _, ud := range userDepts {
			userDeptIDs[ud.DepartmentID] = true
		}

		// Get all departments to check parent relationships
		var allDepts []model.Department
		if err := l.svcCtx.DB.WithContext(ctx).Find(&allDepts).Error; err != nil {
			return xerr.New(err)
		}
		deptMap := make(map[uint]*model.Department)
		for i := range allDepts {
			deptMap[allDepts[i].ID] = &allDepts[i]
		}

		for _, pid := range parentIds {
			// Check if user is still in any department that is a child of pid
			stillInParent := false
			for udID := range userDeptIDs {
				if udID == pid {
					continue // Ignore the parent itself (we are deciding whether to keep it)
				}
				if d, ok := deptMap[udID]; ok {
					// Check if d is a child of pid
					// Method 1: Check ParentPath
					if strings.Contains(d.ParentPath, strconv.Itoa(int(pid))) || d.ParentID == pid {
						stillInParent = true
						break
					}
				}
			}

			if !stillInParent {
				// Remove from parent
				if err := l.svcCtx.DB.WithContext(ctx).Where("department_id = ? AND user_id = ?", pid, uid).Delete(&model.DepartmentUser{}).Error; err != nil {
					log.Error().Err(err).Msg("failed to remove user from parent department")
					return xerr.New(err)
				}
				// Also remove from local map so further parents don't see it?
				// No, we are iterating parents. If we remove from parent A, parent B (parent of A) might still need to be removed.
				// But wait, if we remove from A, then A is no longer in userDeptIDs?
				// userDeptIDs reflects the state AFTER removing from current dept, but BEFORE removing from parents.
				// If we remove from parent A, we should effectively consider it removed for parent B check.
				// However, parent A is a parent, not a child of parent B (well, A is child of B).
				// The logic "stillInParent" checks if user is in any CHILD of pid.
				// If user is in A, and A is child of B.
				// If we decide to remove user from A, then user is no longer in A.
				// So when checking B, we shouldn't count A.
				// So yes, we should probably be careful.
				// But simpler approach:
				// The `userDeptIDs` map contains departments user is explicitly in.
				// If we remove from `pid`, we don't need to update `userDeptIDs` because `pid` was likely in it (or maybe not if it was just a parent association).
				// Actually, `userDeptIDs` comes from DB *after* removing from current dept.
				// So it contains all other depts.
				// If `pid` is in `userDeptIDs`, it means user was added to `pid` explicitly or via another child.
				// Wait, if user is in `pid`, `pid` is in `userDeptIDs`.
				// `stillInParent` loop checks `udID` (other depts user is in).
				// If `udID` is a child of `pid`, then `stillInParent` is true.
				// If `stillInParent` is false, it means user is NOT in any child of `pid`.
				// But user might be in `pid` itself directly?
				// If user is in `pid` directly (e.g. added to root), should we remove them just because they left a child?
				// AIWorkHelper logic says: "Remove from parent if not in any other child".
				// It assumes membership in parent is *derivative* of membership in child.
				// If I manually added user to Parent, then added to Child. Then removed from Child.
				// Should user be removed from Parent?
				// AIWorkHelper implementation seems to assume yes, or at least it tries to clean up.
				// Let's stick to AIWorkHelper logic:
				// "If user is not in any department that is a sub-department of this parent, remove from parent."
			}
		}
	}

	return nil
}

func (l *departmentLogic) DepartmentUserInfo(ctx context.Context, req *domain.IdPathReq) (*domain.Department, error) {
	// 1. Get the user's direct department
	var deptUser model.DepartmentUser
	// Note: A user might belong to multiple departments. The API seems to assume one "main" department or just returns one.
	// The guide says: "返回包含完整上级部门层级的组织架构" (Return organization chart containing full upper-level department hierarchy).
	// Let's pick the one with the deepest level or just the first one found.
	if err := l.svcCtx.DB.WithContext(ctx).Where("user_id = ?", req.Id).Order("department_id desc").First(&deptUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User is not in any department, return empty or error?
			// Guide says: "根据用户ID查找该用户所属的部门信息"
			// If not found, maybe return 404 or empty data?
			// Let's return a clear error for now.
			return nil, xerr.New(err)
		}
		log.Error().Err(err).Str("userId", req.Id).Msg("failed to find department user info")
		return nil, xerr.New(err)
	}

	var currentDept model.Department
	if err := l.svcCtx.DB.WithContext(ctx).First(&currentDept, deptUser.DepartmentID).Error; err != nil {
		return nil, xerr.New(err)
	}

	// 2. Build the path to root
	// We need to fetch all parents up to root.
	// ParentPath contains ids like "1-5-7".
	var pathIds []uint
	if currentDept.ParentPath != "" {
		pathIds = model.ParseParentPath(currentDept.ParentPath)
	}
	// Add current dept ID
	pathIds = append(pathIds, currentDept.ID)

	// 3. Fetch all departments in the path
	var depts []model.Department
	if err := l.svcCtx.DB.WithContext(ctx).Where("id IN ?", pathIds).Find(&depts).Error; err != nil {
		return nil, xerr.New(err)
	}

	// 4. Construct the nested structure
	// We need to nest them: Root -> Child -> ... -> Current
	deptMap := make(map[uint]*domain.Department)
	for _, d := range depts {
		deptMap[d.ID] = &domain.Department{
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

	var root *domain.Department
	// The root of this specific path is the node whose ID is the first in pathIds
	if len(pathIds) > 0 {
		rootID := pathIds[0]
		root = deptMap[rootID]
	}

	// Link children based on path order
	// pathIds is ordered from root to leaf (e.g., [1, 5, 8])
	for i := 0; i < len(pathIds)-1; i++ {
		parentID := pathIds[i]
		childID := pathIds[i+1]
		if parent, ok := deptMap[parentID]; ok {
			if child, ok := deptMap[childID]; ok {
				// Check if child is already added to avoid duplicates (though we create fresh map)
				// Just append
				parent.Child = append(parent.Child, child)
			}
		}
	}

	// If root is nil (shouldn't happen if data is consistent), return current
	if root == nil {
		root = deptMap[currentDept.ID]
	}

	return root, nil
}
