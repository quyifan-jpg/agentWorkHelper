/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Department interface {
	Soa(ctx context.Context) (resp *domain.DepartmentSoaResp, err error)
	Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.Department, err error)
	Create(ctx context.Context, req *domain.Department) (err error)
	Edit(ctx context.Context, req *domain.Department) (err error)
	Delete(ctx context.Context, req *domain.IdPathReq) (err error)
	SetDepartmentUsers(ctx context.Context, req *domain.SetDepartmentUser) (err error)
	AddDepartmentUser(ctx context.Context, req *domain.AddDepartmentUser) (err error)
	RemoveDepartmentUser(ctx context.Context, req *domain.RemoveDepartmentUser) (err error)
	DepartmentUserInfo(ctx context.Context, req *domain.IdPathReq) (resp *domain.Department, err error)
}

type department struct {
	svcCtx *svc.ServiceContext
}

func NewDepartment(svcCtx *svc.ServiceContext) Department {
	return &department{
		svcCtx: svcCtx,
	}
}

// Soa 获取部门SOA信息业务逻辑
func (l *department) Soa(ctx context.Context) (resp *domain.DepartmentSoaResp, err error) {
	deps, err := l.svcCtx.DepartmentModel.All(ctx) // 获取所有部门数据
	if err != nil {
		return nil, err
	}

	// 收集所有负责人ID
	leaderIds := make([]string, 0, len(deps))
	for _, dep := range deps {
		if len(dep.LeaderId) > 0 {
			leaderIds = append(leaderIds, dep.LeaderId)
		}
	}

	// 批量查询所有负责人信息
	leaderMap := make(map[string]string)
	if len(leaderIds) > 0 {
		users, _, err := l.svcCtx.UserModel.List(ctx, &domain.UserListReq{Ids: leaderIds})
		if err == nil {
			for _, user := range users {
				leaderMap[user.ID.Hex()] = user.Name
			}
		}
	}

	// 查询所有部门的成员数量和用户列表
	depUserCountMap := make(map[string]int64)
	depUsersMap := make(map[string][]*domain.DepartmentUser) // 部门ID -> 用户列表映射
	allDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{})
	if err == nil {
		// 收集所有部门用户的用户ID
		allUserIds := make([]string, 0, len(allDepUsers))
		for _, depUser := range allDepUsers {
			allUserIds = append(allUserIds, depUser.UserId)
		}

		// 批量查询所有用户信息
		userMap := make(map[string]string)
		if len(allUserIds) > 0 {
			users, _, err := l.svcCtx.UserModel.List(ctx, &domain.UserListReq{Ids: allUserIds})
			if err == nil {
				for _, user := range users {
					userMap[user.ID.Hex()] = user.Name
				}
			}
		}

		// 填充部门用户信息
		for _, depUser := range allDepUsers {
			depUserCountMap[depUser.DepId]++
			userName := userMap[depUser.UserId]
			depUsersMap[depUser.DepId] = append(depUsersMap[depUser.DepId], &domain.DepartmentUser{
				UserId:   depUser.UserId,
				UserName: userName, // 填充用户名称
			})
		}
	}

	groupDep := make(map[string][]*domain.Department, len(deps)) // 按父路径分组的部门映射
	rootDep := make([]*domain.Department, 0)                     // 根部门列表

	for i, _ := range deps {
		depDomain := deps[i].ToDepartment()
		// 填充负责人名称
		if name, ok := leaderMap[depDomain.LeaderId]; ok {
			depDomain.Leader = name
		}
		// 填充部门成员数量
		depDomain.Count = depUserCountMap[depDomain.Id]
		// 填充部门用户列表
		if users, ok := depUsersMap[depDomain.Id]; ok {
			depDomain.Users = users
		}

		if len(deps[i].ParentPath) == 0 { // 根部门（没有父路径）
			rootDep = append(rootDep, depDomain)
			continue
		}
		// 按父路径分组子部门
		groupDep[deps[i].ParentPath] = append(groupDep[deps[i].ParentPath], depDomain)
	}

	l.buildTree(rootDep, groupDep) // 构建部门树结构

	return &domain.DepartmentSoaResp{
		Child: rootDep, // 返回根部门列表
	}, nil
}

// buildTree 递归构建部门树结构
func (l *department) buildTree(rootDep []*domain.Department, groupDep map[string][]*domain.Department) {
	for i, _ := range rootDep {
		path := model.DepartmentParentPath(rootDep[i].ParentPath, rootDep[i].Id) // 构建当前部门的路径

		data, ok := groupDep[path] // 查找当前部门的子部门
		if !ok || len(data) == 0 {
			continue // 没有子部门，跳过
		}

		l.buildTree(data, groupDep) // 递归构建子部门树

		rootDep[i].Child = data // 设置子部门
	}
}

// Info 获取部门详情业务逻辑
func (l *department) Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.Department, err error) {
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.Id) // 根据ID查找部门
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.UserModel.FindOne(ctx, dep.LeaderId) // 查找部门负责人信息
	if err != nil {
		return nil, err
	}

	var res domain.Department
	copier.Copy(&res, dep) // 复制部门数据

	res.Leader = user.Name // 设置负责人姓名

	return &res, nil
}

// Create 创建部门业务逻辑
func (l *department) Create(ctx context.Context, req *domain.Department) (err error) {
	dep, err := l.svcCtx.DepartmentModel.FindByName(ctx, req.Name) // 检查部门名称是否已存在
	if err != nil && err != model.ErrDepartmentNotFound {
		return err
	}
	if dep != nil {
		return errors.New("已存在该部门") // 部门名称重复
	}

	var parentPath string
	if len(req.ParentId) > 0 { // 如果有父部门
		pdep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.ParentId) // 查找父部门
		if err != nil && err != model.ErrDepartmentNotFound {
			return err
		}
		parentPath = model.DepartmentParentPath(pdep.ParentPath, req.ParentId) // 构建父路径
	}

	depId := primitive.NewObjectID() // 生成新的部门ID

	err = l.svcCtx.DepartmentModel.Insert(ctx, &model.Department{ // 插入新部门到数据库
		ID:         depId,
		Name:       req.Name,
		ParentId:   req.ParentId,
		ParentPath: parentPath,
		Level:      req.Level,
		LeaderId:   req.LeaderId,
		Count:      1, // 创建时默认有1个成员（负责人）
		CreateAt:   time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	// 将部门负责人添加到当前部门，并级联添加到所有父部门
	// 使用 AddDepartmentUser 方法，它会自动处理级联逻辑
	return l.AddDepartmentUser(ctx, &domain.AddDepartmentUser{
		DepId:  depId.Hex(),
		UserId: req.LeaderId,
	})
}

// Edit 更新部门业务逻辑
func (l *department) Edit(ctx context.Context, req *domain.Department) (err error) {
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.Id) // 查找要更新的部门
	if err != nil {
		return err
	}

	dep2, err := l.svcCtx.DepartmentModel.FindByName(ctx, req.Name) // 检查新名称是否重复
	if err != nil && err != model.ErrDepartmentNotFound {
		return err
	}
	if dep2 != nil && dep2.ID.Hex() != dep.ID.Hex() {
		return errors.New("已存在该部门") // 名称重复且不是当前部门
	}

	return l.svcCtx.DepartmentModel.Update(ctx, &model.Department{ // 更新部门信息
		ID:       dep.ID,
		Name:     req.Name,
		ParentId: req.ParentId,
		Level:    req.Level,
		LeaderId: req.LeaderId,
	})
}

// Delete 删除部门业务逻辑
func (l *department) Delete(ctx context.Context, req *domain.IdPathReq) (err error) {
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.Id) // 查找要删除的部门
	if err != nil {
		if err == model.ErrDepartmentNotFound {
			return nil // 部门不存在，视为删除成功
		}
		return err
	}

	depUser, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: req.Id}) // 查找部门下的用户
	if err != nil {
		return err
	}

	if len(depUser) == 0 { // 部门下没有用户，可以直接删除
		return l.svcCtx.DepartmentModel.Delete(ctx, req.Id)
	}

	if len(depUser) > 1 || depUser[0].UserId != dep.LeaderId { // 部门下有其他用户（非负责人）
		return errors.New("该部门下还存在用户，不能删除该部门")
	}

	// 只有负责人的情况下，删除部门前需要：
	// 1. 使用 RemoveDepartmentUser 智能删��负责人（会检查是否在其他子部门）
	// 2. 然后删除部门本身

	leaderId := dep.LeaderId

	// 使用 RemoveDepartmentUser 方法删除负责人
	// 这个方法会智能处理：只有当负责人不在该父部门的其他子部门时，才从父部门删除
	// 但是这个方法会检查是否是负责人并拒绝删除，所以我们需要直接操作

	// 删除负责人在当前部门的关联
	for _, du := range depUser {
		if du.UserId == leaderId {
			err = l.svcCtx.DepartmentUserModel.Delete(ctx, du.ID.Hex())
			if err != nil {
				return err
			}
			break
		}
	}

	// 如果有父部门，智能地从父部门删除该负责人
	// 关键：只有当负责人不在该父部门管辖的任何其他部门中时，才从父部门删除
	if len(dep.ParentPath) > 0 {
		parentIds := model.ParseParentPath(dep.ParentPath)

		// 查询该负责人在所有部门中的关联
		allUserDeps, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{})
		if err != nil {
			return err
		}

		// 获取该负责人当前所在的所有部门ID（排除刚删除的部门）
		leaderDepIds := make(map[string]bool)
		for _, ud := range allUserDeps {
			if ud.UserId == leaderId && ud.DepId != req.Id {
				leaderDepIds[ud.DepId] = true
			}
		}

		// 如果负责人已经不在任何部门了，直接从所有父部门删除
		if len(leaderDepIds) == 0 {
			for _, parentId := range parentIds {
				parentDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: parentId})
				if err != nil {
					continue
				}

				for _, pdu := range parentDepUsers {
					if pdu.UserId == leaderId {
						l.svcCtx.DepartmentUserModel.Delete(ctx, pdu.ID.Hex())
						break
					}
				}
			}
		} else {
			// 获取所有部门信息
			allDeps, err := l.svcCtx.DepartmentModel.All(ctx)
			if err != nil {
				return err
			}

			// 构建部门ID到部门的映射
			depMap := make(map[string]*model.Department)
			for _, d := range allDeps {
				depMap[d.ID.Hex()] = d
			}

			// 关键修复：反转parentIds顺序，从近到远逐级处理
			// ParseParentPath返回的是从远到近的顺序，需要反转
			for i := len(parentIds) - 1; i >= 0; i-- {
				parentId := parentIds[i]

				// 检查负责人是否还在该父部门管辖的其他部门中
				// 关键：排除当前要检查的父部门自己
				stillUnderThisParent := false

				for leaderDepId := range leaderDepIds {
					// 跳过当前检查的父部门自己
					if leaderDepId == parentId {
						continue
					}

					leaderDep, exists := depMap[leaderDepId]
					if !exists {
						continue
					}

					// 检查这个负责人所在的部门是否在当前父部门的管辖下
					if strings.Contains(leaderDep.ParentPath, parentId) || leaderDep.ParentId == parentId {
						stillUnderThisParent = true
						break
					}
				}

				// 只有当负责人不在该父部门管辖的任何部门中时，才从父部门删除
				if !stillUnderThisParent {
					parentDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: parentId})
					if err != nil {
						continue
					}

					for _, pdu := range parentDepUsers {
						if pdu.UserId == leaderId {
							err = l.svcCtx.DepartmentUserModel.Delete(ctx, pdu.ID.Hex())
							if err != nil {
								continue
							}
							// 删除成功后，从leaderDepIds中移除这个父部门
							delete(leaderDepIds, parentId)
							break
						}
					}
				}
			}
		}
	}

	// 最后删除部门本身
	return l.svcCtx.DepartmentModel.Delete(ctx, req.Id)
}

// SetDepartmentUsers 设置部门用户业务逻辑
func (l *department) SetDepartmentUsers(ctx context.Context, req *domain.SetDepartmentUser) (err error) {
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.DepId) // 验证部门是否存在
	if err != nil {
		return err
	}

	// 1. 获取当前部门的所有用户
	currentDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: req.DepId})
	if err != nil {
		return err
	}

	// 将当前用户ID转为map方便查找
	currentUserMap := make(map[string]bool)
	for _, du := range currentDepUsers {
		currentUserMap[du.UserId] = true
	}

	// 将新用户ID转为map方便查找
	newUserMap := make(map[string]bool)
	for _, uid := range req.UserIds {
		newUserMap[uid] = true
	}

	// 2. 找出需要删除的用户(在当前列表中但不在新列表中)
	for _, du := range currentDepUsers {
		if !newUserMap[du.UserId] {
			// 不能删除部门负责人
			if du.UserId == dep.LeaderId {
				continue
			}
			// 使用级联删除方法
			err = l.RemoveDepartmentUser(ctx, &domain.RemoveDepartmentUser{
				DepId:  req.DepId,
				UserId: du.UserId,
			})
			if err != nil {
				// 记录错误但继续处理
				continue
			}
		}
	}

	// 3. 找出需要添加的用户(在新列表中但不在当前列表中)
	for _, uid := range req.UserIds {
		if !currentUserMap[uid] {
			// 使用级联添加方法
			err = l.AddDepartmentUser(ctx, &domain.AddDepartmentUser{
				DepId:  req.DepId,
				UserId: uid,
			})
			if err != nil {
				// 如果是"已在部门中"的错误,忽略并继续
				if strings.Contains(err.Error(), "已在此部门中") {
					continue
				}
				// 其他错误也继续处理
				continue
			}
		}
	}

	return nil
}

// DepartmentUserInfo 获取用户部门信息业务逻辑（包含完整的上级部门层级）
func (l *department) DepartmentUserInfo(ctx context.Context, req *domain.IdPathReq) (resp *domain.Department, err error) {
	// 根据用户ID查找用户所属的部门关联
	depUser, err := l.svcCtx.DepartmentUserModel.FindByUserId(ctx, req.Id)
	if err != nil {
		return nil, errors.New("用户未关联任何部门")
	}

	// 根据部门ID查找部门信息
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, depUser.DepId)
	if err != nil {
		return nil, errors.New("用户关联的部门不存在")
	}

	// 如果是根部门，直接返回
	if len(dep.ParentPath) == 0 {
		return dep.ToDepartment(), nil
	}

	// 解析父路径，获取所有上级部门ID
	parentIds := model.ParseParentPath(dep.ParentPath)
	pdeps, err := l.svcCtx.DepartmentModel.ListToMap(ctx, &domain.DepartmentListReq{
		DepIds: parentIds,
	})
	if err != nil {
		return nil, err
	}

	// 构建完整的部门层级结构
	var root *domain.Department
	var node *domain.Department
	for _, id := range parentIds {
		if _, ok := pdeps[id]; !ok {
			continue
		}

		if root == nil {
			// 第一个父部门作为根节点
			root = pdeps[id].ToDepartment()
			node = root
			continue
		}
		// 构建层级关系
		tmp := pdeps[id].ToDepartment()
		node.Child = append(node.Child, tmp)
		node = tmp
	}

	// 将用户直接关联的部门添加为最后一级
	if node != nil {
		node.Child = append(node.Child, dep.ToDepartment())
	}

	return root, nil
}

// AddDepartmentUser 添加部门员工业务逻辑
// 当员工加入子部门时,会自动加入该部门的所有上级部门
func (l *department) AddDepartmentUser(ctx context.Context, req *domain.AddDepartmentUser) (err error) {
	// 验证部门是否存在
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.DepId)
	if err != nil {
		return err
	}

	// 验证用户是否存在
	_, err = l.svcCtx.UserModel.FindOne(ctx, req.UserId)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查用户是否已在该部门
	depUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: req.DepId})
	if err != nil {
		return err
	}
	for _, du := range depUsers {
		if du.UserId == req.UserId {
			return errors.New("该用户已在此部门中")
		}
	}

	// 1. 添加用户到当前部门
	err = l.svcCtx.DepartmentUserModel.Insert(ctx, &model.DepartmentUser{
		DepId:  req.DepId,
		UserId: req.UserId,
	})
	if err != nil {
		return err
	}

	// 2. 如果有上级部门,将用户也添加到所有上级部门
	if len(dep.ParentPath) > 0 {
		// 解析父路径获取所有上级部门ID
		parentIds := model.ParseParentPath(dep.ParentPath)

		// 为每个上级部门添加该用户
		for _, parentId := range parentIds {
			// 检查用户是否已在上级部门中
			parentDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: parentId})
			if err != nil {
				continue // 查询失败则跳过该上级部门
			}

			// 检查是否已存在
			exists := false
			for _, pdu := range parentDepUsers {
				if pdu.UserId == req.UserId {
					exists = true
					break
				}
			}

			// 如果不存在则添加
			if !exists {
				err = l.svcCtx.DepartmentUserModel.Insert(ctx, &model.DepartmentUser{
					DepId:  parentId,
					UserId: req.UserId,
				})
				if err != nil {
					// 添加失败,记录错误但继续处理其他部门
					continue
				}
			}
		}
	}

	return nil
}

// RemoveDepartmentUser 删除部门员工业务逻辑
// 当员工从子部门删除时,会自动从该部门的所有上级部门中删除
func (l *department) RemoveDepartmentUser(ctx context.Context, req *domain.RemoveDepartmentUser) (err error) {
	// 验证部门是否存在
	dep, err := l.svcCtx.DepartmentModel.FindOne(ctx, req.DepId)
	if err != nil {
		return err
	}

	// 不能删除部门负责人
	if req.UserId == dep.LeaderId {
		return errors.New("不能删除部门负责人")
	}

	// 检查用户是否在该部门
	depUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: req.DepId})
	if err != nil {
		return err
	}

	found := false
	var targetDepUser *model.DepartmentUser
	for _, du := range depUsers {
		if du.UserId == req.UserId {
			found = true
			targetDepUser = du
			break
		}
	}

	if !found {
		return errors.New("该用户不在此部门中")
	}

	// 1. 删除用户与当前部门的关联
	err = l.svcCtx.DepartmentUserModel.Delete(ctx, targetDepUser.ID.Hex())
	if err != nil {
		return err
	}

	// 2. 如果有上级部门,需要智能地从上级部门删除该用户
	// 关键：只有当用户不在该父部门管辖的任何其他部门中时，才从父部门删除
	if len(dep.ParentPath) > 0 {
		// 解析父路径获取所有上级部门ID
		parentIds := model.ParseParentPath(dep.ParentPath)

		// 查询该用户在所有部门中的关联（用于判断用户还在哪些部门）
		allUserDeps, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{})
		if err != nil {
			return err
		}

		// 获取该用户当前所在的所有部门ID（排除刚删除的部门）
		userDepIds := make(map[string]bool)
		for _, ud := range allUserDeps {
			if ud.UserId == req.UserId && ud.DepId != req.DepId {
				userDepIds[ud.DepId] = true
			}
		}

		// 如果用户已经不在任何部门了，直接从所有父部门删除
		if len(userDepIds) == 0 {
			for _, parentId := range parentIds {
				parentDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: parentId})
				if err != nil {
					continue
				}

				for _, pdu := range parentDepUsers {
					if pdu.UserId == req.UserId {
						l.svcCtx.DepartmentUserModel.Delete(ctx, pdu.ID.Hex())
						break
					}
				}
			}
			return nil
		}

		// 获取所有部门信息（用于检查部门的ParentPath）
		allDeps, err := l.svcCtx.DepartmentModel.All(ctx)
		if err != nil {
			return err
		}

		// 构建部门ID到部门的映射
		depMap := make(map[string]*model.Department)
		for _, d := range allDeps {
			depMap[d.ID.Hex()] = d
		}

		// 关键修复：反转parentIds顺序，从近到远逐级处理（技术研发部 -> 公司总部）
		// ParseParentPath返回的是从远到近的顺序，需要反转
		for i := len(parentIds) - 1; i >= 0; i-- {
			parentId := parentIds[i]

			// 检查用户是否还在该父部门管辖的其他部门中
			// 关键：排除当前要检查的父部门自己
			stillUnderThisParent := false

			for userDepId := range userDepIds {
				// 跳过当前检查的父部门自己
				if userDepId == parentId {
					continue
				}

				userDep, exists := depMap[userDepId]
				if !exists {
					continue
				}

				// 检查这个用户所在的部门是否在当前父部门的管辖下
				// 方法1：检查ParentPath是否包含parentId
				if strings.Contains(userDep.ParentPath, parentId) {
					stillUnderThisParent = true
					break
				}
				// 方法2：检查ParentId是否等于parentId
				if userDep.ParentId == parentId {
					stillUnderThisParent = true
					break
				}
			}

			// 只有当用户不在该父部门管辖的任何部门中时，才从父部门删除
			if !stillUnderThisParent {
				parentDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{DepId: parentId})
				if err != nil {
					continue
				}

				for _, pdu := range parentDepUsers {
					if pdu.UserId == req.UserId {
						err = l.svcCtx.DepartmentUserModel.Delete(ctx, pdu.ID.Hex())
						if err != nil {
							continue
						}
						// 删除成功后，从userDepIds中移除这个父部门
						delete(userDepIds, parentId)
						break
					}
				}
			}
		}
	}

	return nil
}
