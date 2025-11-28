/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package logic

import (
	"AIWorkHelper/pkg/timeutil"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/token"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Approval 定义审批业务逻辑接口
type Approval interface {
	// Info 获取审批详情
	Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.ApprovalInfoResp, err error)
	// Create 创建审批申请
	Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error)
	// Dispose 处理审批申请
	Dispose(ctx context.Context, req *domain.DisposeReq) (err error)
	// List 获取审批列表
	List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error)
}

// approval 审批业务逻辑实现结构体
type approval struct {
	svcCtx *svc.ServiceContext // 服务上下文，包含数据库连接等
}

// NewApproval 创建审批业务逻辑实例
func NewApproval(svcCtx *svc.ServiceContext) Approval {
	return &approval{
		svcCtx: svcCtx,
	}
}

// Info 获取审批详情信息
func (l *approval) Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.ApprovalInfoResp, err error) {
	// 根据ID查找审批记录
	approval, err := l.svcCtx.ApprovalModel.FindOne(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	// 转换为响应格式
	resp = approval.ToDomainApprovalInfo()
	// 获取参与审批的所有用户信息
	users, err := l.svcCtx.UserModel.ListToMaps(ctx, &domain.UserListReq{
		Ids: approval.Participation,
	})
	if err != nil || len(users) == 0 {
		return resp, err
	}
	// 设置申请人信息
	resp.User = &domain.Approver{
		UserId:   users[approval.UserId].ID.Hex(),
		UserName: users[approval.UserId].Name,
	}
	// 设置当前审批人信息
	resp.Approver = &domain.Approver{
		UserId:   users[approval.ApprovalId].ID.Hex(),
		UserName: users[approval.ApprovalId].Name,
	}
	// 设置所有审批人信息列表
	for _, approver := range approval.Approvers {
		resp.Approvers = append(resp.Approvers, &domain.Approver{
			UserId:   users[approver.UserId].ID.Hex(),
			UserName: users[approver.UserId].Name,
			Status:   int(approver.Status),
			Reason:   approver.Reason,
		})
	}

	return
}

// Create 创建审批申请
func (l *approval) Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error) {
	// 从上下文中获取当前用户ID
	uid := token.GetUId(ctx)
	req.UserId = uid
	// 创建新的审批记录
	approval := l.newApproval(req)

	var abstract string
	// 根据审批类型处理不同的审批内容
	switch model.ApprovalType(req.Type) {
	case model.LeaveApproval: // 请假审批
		if req.Leave != nil {
			approval.Leave = &model.Leave{
				Type:      model.LeaveType(req.Leave.Type),
				StartTime: req.Leave.StartTime,
				EndTime:   req.Leave.EndTime,
				Reason:    req.Leave.Reason,
				TimeType:  model.TimeFormatType(req.Leave.TimeType),
			}
			// 生成请假审批摘要
			abstract = fmt.Sprintf("【%s】: 【%s】-【%s】", model.LeaveType(req.Leave.Type).ToString(),
				timeutil.Format(req.Leave.StartTime), timeutil.Format(req.Leave.EndTime))
			approval.Reason = req.Leave.Reason
		}
	case model.GoOutApproval: // 外出审批
		// 计算外出时长（小时）
		duration := float32(req.GoOut.EndTime-req.GoOut.StartTime) / 3600.0
		approval.GoOut = &model.GoOut{
			StartTime: req.GoOut.StartTime,
			EndTime:   req.GoOut.EndTime,
			Duration:  duration,
			Reason:    req.GoOut.Reason,
		}
		// 生成外出审批摘要
		abstract = fmt.Sprintf("【%s】-【%s】", timeutil.Format(req.GoOut.StartTime), timeutil.Format(req.GoOut.EndTime))
		approval.Reason = req.GoOut.Reason
	case model.MakeCardApproval: // 补卡审批
		approval.MakeCard = &model.MakeCard{
			Date:      req.MakeCard.Date,
			Reason:    req.MakeCard.Reason,
			Day:       req.MakeCard.Day,
			CheckType: model.WorkCheckType(req.MakeCard.CheckType),
		}
		// 生成补卡审批摘要
		abstract = fmt.Sprintf("【%s】【%s】", timeutil.Format(req.MakeCard.Date), req.MakeCard.Reason)
		approval.Reason = req.MakeCard.Reason
	default:
		// 其他类型审批处理
	}

	// 获取申请人信息
	user, err := l.svcCtx.UserModel.FindOne(ctx, uid)
	if err != nil {
		return
	}
	// 生成审批标题和摘要
	approval.Title = fmt.Sprintf("%s 提交的 %s", user.Name, model.ApprovalType(req.Type).ToString())
	approval.Abstract = abstract

	// 设置审批流程：根据部门层级确定审批人
	// 关键：需要找到用户实际直接所属的部门（层级最深的部门），而不是级联关联的父部门
	allDepUsers, err := l.svcCtx.DepartmentUserModel.List(ctx, &domain.DepartmentListReq{})
	if err != nil {
		return
	}

	// 获取用户所有部门关联
	var userDepIds []string
	for _, du := range allDepUsers {
		if du.UserId == user.ID.Hex() {
			userDepIds = append(userDepIds, du.DepId)
		}
	}

	if len(userDepIds) == 0 {
		return nil, errors.New("用户未关联任何部���")
	}

	// 查询所有这些部门的详细信息
	userDeps, err := l.svcCtx.DepartmentModel.List(ctx, &domain.DepartmentListReq{
		DepIds: userDepIds,
	})
	if err != nil {
		return
	}

	// 找出层级最深的部门（ParentPath最长的，即用户实际直接所属的部门）
	var dep *model.Department
	maxPathLen := -1
	for _, d := range userDeps {
		pathLen := len(d.ParentPath)
		if pathLen > maxPathLen {
			maxPathLen = pathLen
			dep = d
		}
	}

	if dep == nil {
		return nil, errors.New("未找到用户所属部门")
	}

	// 解析部门父级路径，构建审批层级
	parentIds := model.ParseParentPath(dep.ParentPath)
	pdeps, err := l.svcCtx.DepartmentModel.ListToMap(ctx, &domain.DepartmentListReq{
		DepId:  "",
		DepIds: parentIds,
	})
	var (
		approvals      []*model.Approver // 审批人列表
		participations []string          // 参与人员ID列表
	)
	// 添加直属部门负责人作为第一级审批人
	approvals = append(approvals, &model.Approver{
		UserId: dep.LeaderId,
		Status: model.Processed, // 设置为待处理状态
	})
	participations = append(participations, dep.LeaderId, uid)

	// 按部门层级从下到上添加审批人
	for i := len(parentIds) - 1; i > 0; i-- {
		if _, ok := pdeps[parentIds[i]]; !ok {
			continue
		}
		// 添加上级部门负责人作为审批人
		approvals = append(approvals, &model.Approver{
			UserId: pdeps[parentIds[i]].LeaderId,
		})
		// 将上级部门负责人加入参与人员列表
		participations = append(participations, pdeps[parentIds[i]].LeaderId)
	}

	// 设置审批流程相关信息
	approval.Approvers = approvals          // 审批人列表
	approval.Participation = participations // 参与人员列表
	approval.ApprovalId = dep.LeaderId      // 当前审批人ID（从直属部门负责人开始）
	approval.UserId = uid                   // 申请人ID

	// 保存审批记录到数据库
	if err = l.svcCtx.ApprovalModel.Insert(ctx, approval); err != nil {
		return
	}

	// 返回创建成功的审批ID
	return &domain.IdResp{
		Id: approval.ID.Hex(),
	}, nil
}

// Dispose 处理审批申请（通过/拒绝/撤销）
func (l *approval) Dispose(ctx context.Context, req *domain.DisposeReq) (err error) {
	// 根据审批ID查找审批记录
	approval, err := l.svcCtx.ApprovalModel.FindOne(ctx, req.ApprovalId)
	if err != nil {
		return err
	}
	uid := token.GetUId(ctx)

	// 处理撤销操作
	if model.ApprovalStatus(req.Status) == model.Cancel {
		// 只有申请人才能撤销自己的申请
		if req.ApprovalId != approval.UserId {
			return errors.New("审核用户错误")
		}
		// 设置审批状态为撤销
		approval.Status = model.Cancel
		return l.svcCtx.ApprovalModel.Update(ctx, approval)
	}

	// 处理审批通过或拒绝操作
	// 验证当前用户是否为当前审批人
	if approval.ApprovalId != uid {
		return errors.New("审核用户错误")
	}
	// 检查审批状态，避免重复操作
	switch approval.Status {
	case model.Cancel:
		return errors.New("该审核已撤销")
	case model.Pass:
		return errors.New("该审核已通过")
	case model.Refuse:
		return errors.New("该审核已拒绝")
	}

	// 处理拒绝操作
	if model.ApprovalStatus(req.Status) == model.Refuse {
		// 记录当前审批人的拒绝状态和原因
		approval.Approvers[approval.ApprovalIdx].Status = model.Refuse
		approval.Approvers[approval.ApprovalIdx].Reason = req.Reason
		// 设置整体审批状态为拒绝
		approval.Status = model.Refuse
	} else if model.ApprovalStatus(req.Status) == model.Pass {
		// 处理通过操作
		// 记录当前审批人的通过状态和原因
		approval.Approvers[approval.ApprovalIdx].Status = model.Pass
		approval.Approvers[approval.ApprovalIdx].Reason = req.Reason

		// 如果还有下一级审批人，则流转到下一级
		if len(approval.Approvers)-1 > approval.ApprovalIdx {
			approval.ApprovalIdx++                                                // 审批索引递增
			approval.ApprovalId = approval.Approvers[approval.ApprovalIdx].UserId // 设置下一个审批人
			// 整体状态保持为处理中
		} else {
			// 这是最后一个审批人，检查是否所有审批人都已通过
			isPass := true
			for _, approver := range approval.Approvers {
				if approver.Status != model.Pass {
					isPass = false
					break
				}
			}
			// 如果所有审批人都已通过，则设置整体状态为通过
			if isPass {
				approval.Status = model.Pass
			}
		}
	}

	// 更新审批记录到数据库
	return l.svcCtx.ApprovalModel.Update(ctx, approval)
}

// List 获取审批列表
func (l *approval) List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error) {
	// 从Token中获取当前用户ID，确保查询的是当前登录用户的数据
	uid := token.GetUId(ctx)
	req.UserId = uid

	// 从数据库查询审批列表数据
	data, count, err := l.svcCtx.ApprovalModel.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	var list []*domain.ApprovalList
	for i, _ := range data {
		list = append(list, data[i].ToDomainApprovalList())
	}

	// 返回列表数据和总数
	return &domain.ApprovalListResp{
		List:  list,
		Count: count,
	}, nil
}

// newApproval 创建新的审批记录实例
func (l *approval) newApproval(req *domain.Approval) *model.Approval {
	now := time.Now().Unix()
	return &model.Approval{
		ID:       primitive.NewObjectID(),      // 生成新的ObjectID
		UserId:   req.UserId,                   // 申请人ID
		No:       GenRandomNo(11),              // 生成随机审批编号
		Type:     model.ApprovalType(req.Type), // 审批类型
		Status:   model.Processed,              // 初始状态为处理中
		Reason:   req.Reason,                   // 申请理由
		CreateAt: now,                          // 创建时间
		UpdateAt: now,                          // 更新时间
	}
}

// GenRandomNo 生成指定位数的随机数字字符串
func GenRandomNo(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano()) // 设置随机种子

	var sb bytes.Buffer
	// 生成指定位数的随机数字
	for i := 0; i < width; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}
