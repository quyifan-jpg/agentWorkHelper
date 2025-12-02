package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/token"
)

type Approval interface {
	Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.ApprovalInfoResp, err error)
	Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error)
	Dispose(ctx context.Context, req *domain.DisposeReq) (err error)
	List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error)
}

type approval struct {
	svcCtx *svc.ServiceContext
}

func NewApproval(svcCtx *svc.ServiceContext) Approval {
	return &approval{
		svcCtx: svcCtx,
	}
}

func (l *approval) Info(ctx context.Context, req *domain.IdPathReq) (resp *domain.ApprovalInfoResp, err error) {
	var approval model.Approval
	// 关联查询：查出审批单，同时查出申请人信息，以及审批流程记录（包含审批人信息）
	if err := l.svcCtx.DB.WithContext(ctx).
		Preload("User").
		Preload("Approvers").
		Preload("Approvers.User").
		First(&approval, req.Id).Error; err != nil {
		return nil, err
	}

	// 转换基础信息
	resp = &domain.ApprovalInfoResp{
		Id:          strconv.Itoa(int(approval.ID)),
		No:          approval.No,
		Type:        approval.Type,
		Status:      approval.Status,
		Title:       approval.Title,
		Abstract:    approval.Abstract,
		Reason:      approval.Reason,
		FinishAt:    approval.FinishAt.Unix(),
		FinishDay:   approval.FinishDay,
		FinishMonth: approval.FinishMonth,
		FinishYeas:  approval.FinishYeas,
		UpdateAt:    approval.UpdatedAt.Unix(),
		CreateAt:    approval.CreatedAt.Unix(),
	}

	// 转换申请人信息
	resp.User = &domain.Approver{
		UserId:   strconv.Itoa(int(approval.UserID)),
		UserName: approval.User.Name,
	}

	// 转换 JSON 详情数据
	// GORM serializer 会自动处理，我们只需要把 model 里的 struct 转为 domain 里的 struct
	// 这里可以使用 copier，或者手动赋值。为了简单直接手动赋值。
	if approval.MakeCard != nil {
		resp.MakeCard = &domain.MakeCard{
			Date:      approval.MakeCard.Date,
			Reason:    approval.MakeCard.Reason,
			Day:       approval.MakeCard.Day,
			CheckType: approval.MakeCard.CheckType,
		}
	}
	if approval.Leave != nil {
		resp.Leave = &domain.Leave{
			Type:      approval.Leave.Type,
			StartTime: approval.Leave.StartTime,
			EndTime:   approval.Leave.EndTime,
			Duration:  approval.Leave.Duration,
			Reason:    approval.Leave.Reason,
			TimeType:  approval.Leave.TimeType,
		}
	}
	if approval.GoOut != nil {
		resp.GoOut = &domain.GoOut{
			StartTime: approval.GoOut.StartTime,
			EndTime:   approval.GoOut.EndTime,
			Duration:  approval.GoOut.Duration,
			Reason:    approval.GoOut.Reason,
		}
	}

	// 处理审批流程
	for _, approver := range approval.Approvers {
		// 添加到审批人列表
		resp.Approvers = append(resp.Approvers, &domain.Approver{
			UserId:   strconv.Itoa(int(approver.UserID)),
			UserName: approver.User.Name,
			Status:   approver.Status,
			Reason:   approver.Reason,
		})

		// 查找当前审批人 (第一个待审批的人)
		// 注意：这里假设 Approvers 是按顺序插入的。如果不是，可能需要 Sort 一下。
		if resp.Approver == nil && approver.Status == 0 { // 0: 待审批
			resp.Approver = &domain.Approver{
				UserId:   strconv.Itoa(int(approver.UserID)),
				UserName: approver.User.Name,
				Status:   approver.Status,
			}
		}
	}

	return resp, nil
}

func (l *approval) Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error) {
	userID, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}

	approval := &model.Approval{
		No:       req.No,
		Title:    req.Title,
		Reason:   req.Reason,
		Type:     req.Type,
		Status:   req.Status,
		Abstract: req.Abstract,
		UserID:   uint(userID),
	}

	if err := l.svcCtx.DB.WithContext(ctx).Create(approval).Error; err != nil {
		return nil, err
	}

	// Save Approvers
	if len(req.Approvers) > 0 {
		var approvers []model.Approver
		for _, a := range req.Approvers {
			uid, _ := strconv.Atoi(a.UserId)
			approvers = append(approvers, model.Approver{
				ApprovalID: approval.ID,
				UserID:     uint(uid),
				Status:     0,
			})
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&approvers).Error; err != nil {
			return nil, err
		}
	}

	return &domain.IdResp{
		Id: strconv.Itoa(int(approval.ID)),
	}, nil
}

func (l *approval) Dispose(ctx context.Context, req *domain.DisposeReq) (err error) {
	userID, err := token.GetUserID(ctx)
	if err != nil {
		return err
	}

	// 1. Check if approval exists
	var approval model.Approval
	if err := l.svcCtx.DB.WithContext(ctx).Preload("Approvers").First(&approval, req.ApprovalId).Error; err != nil {
		return err
	}

	// 2. Find the approver record for current user
	var currentApprover *model.Approver
	for i := range approval.Approvers {
		if approval.Approvers[i].UserID == userID {
			currentApprover = &approval.Approvers[i]
			break
		}
	}

	if currentApprover == nil {
		return errors.New("you are not an approver for this request")
	}

	if currentApprover.Status != 0 {
		return errors.New("you have already processed this request")
	}

	// 3. Update approver status
	currentApprover.Status = req.Status
	currentApprover.Reason = req.Reason
	if err := l.svcCtx.DB.WithContext(ctx).Save(currentApprover).Error; err != nil {
		return err
	}

	// 4. Update approval status
	// Logic:
	// - If Reject (2) -> Approval Rejected (2)
	// - If Pass (1) -> Check if all passed -> Approval Passed (1)

	if req.Status == 2 {
		approval.Status = 2
	} else if req.Status == 1 {
		allPassed := true
		for _, a := range approval.Approvers {
			if a.Status != 1 {
				allPassed = false
				break
			}
		}
		if allPassed {
			approval.Status = 1
			approval.FinishAt = time.Now()
		}
	}

	if err := l.svcCtx.DB.WithContext(ctx).Save(&approval).Error; err != nil {
		return err
	}

	return nil
}

func (l *approval) List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error) {
	// 1. 处理分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	count := req.Count
	if count < 1 {
		count = 10
	}
	offset := (page - 1) * count

	// 2. 构建查询
	db := l.svcCtx.DB.WithContext(ctx).Model(&model.Approval{})

	// 过滤条件
	if req.UserId != "" {
		userId, _ := strconv.Atoi(req.UserId)
		if userId > 0 {
			db = db.Where("user_id = ?", userId)
		}
	}
	if req.Type > 0 {
		db = db.Where("type = ?", req.Type)
	}

	// 3. 查询总数
	var total int64
	if err = db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 4. 查询列表数据
	var approvals []*model.Approval
	if err = db.Order("id desc").Offset(offset).Limit(count).Find(&approvals).Error; err != nil {
		return nil, err
	}

	// 5. 组装响应
	list := make([]*domain.ApprovalList, 0, len(approvals))
	for _, v := range approvals {
		list = append(list, &domain.ApprovalList{
			Id:              strconv.Itoa(int(v.ID)),
			Type:            v.Type,
			Status:          v.Status,
			Title:           v.Title,
			Abstract:        v.Abstract,
			CreateId:        strconv.Itoa(int(v.UserID)),
			ParticipatingId: "", // 暂不处理
		})
	}

	resp = &domain.ApprovalListResp{
		Count: total,
		List:  list,
	}

	return resp, nil
}
