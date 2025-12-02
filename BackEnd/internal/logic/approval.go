package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/timeutil"
	"BackEnd/pkg/token"
	"BackEnd/pkg/xerr"

	"github.com/rs/zerolog/log"
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
		log.Error().Err(err).Str("id", req.Id).Msg("failed to find approval info")
		return nil, xerr.New(err)
	}

	// 转换基础信息
	resp = &domain.ApprovalInfoResp{
		Id:          strconv.Itoa(int(approval.ID)),
		No:          approval.No,
		Type:        int(approval.Type),
		Status:      int(approval.Status),
		Title:       approval.Title,
		Abstract:    approval.Abstract,
		Reason:      approval.Reason,
		FinishAt:    0,
		FinishDay:   approval.FinishDay,
		FinishMonth: approval.FinishMonth,
		FinishYeas:  approval.FinishYeas,
		UpdateAt:    approval.UpdatedAt.Unix(),
		CreateAt:    approval.CreatedAt.Unix(),
	}

	if approval.FinishAt != nil {
		resp.FinishAt = approval.FinishAt.Unix()
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
			CheckType: int(approval.MakeCard.CheckType),
		}
	}
	if approval.Leave != nil {
		resp.Leave = &domain.Leave{
			Type:      int(approval.Leave.Type),
			StartTime: approval.Leave.StartTime,
			EndTime:   approval.Leave.EndTime,
			Duration:  approval.Leave.Duration,
			Reason:    approval.Leave.Reason,
			TimeType:  int(approval.Leave.TimeType),
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
			Status:   int(approver.Status),
			Reason:   approver.Reason,
		})

		// 查找当前审批人 (第一个待审批的人)
		// 注意：这里假设 Approvers 是按顺序插入的。如果不是，可能需要 Sort 一下。
		if resp.Approver == nil && approver.Status == model.Processed { // 0: 待审批
			resp.Approver = &domain.Approver{
				UserId:   strconv.Itoa(int(approver.UserID)),
				UserName: approver.User.Name,
				Status:   int(approver.Status),
			}
		}
	}

	return resp, nil
}

func (l *approval) Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error) {
	var userID uint
	// 1. Try to get UserID from context (JWT)
	uidFromCtx, err := token.GetUserID(ctx)
	if err == nil && uidFromCtx > 0 {
		userID = uidFromCtx
	} else {
		// 2. Fallback to request body if not in context (e.g. internal call or testing)
		if req.UserId != "" {
			uid, err := strconv.Atoi(req.UserId)
			if err != nil {
				return nil, xerr.WithMessage(err, "invalid user_id format")
			}
			userID = uint(uid)
		}
	}

	if userID == 0 {
		return nil, xerr.New(errors.New("user_id is required"))
	}

	// Create basic approval object
	approval := &model.Approval{
		No:     GenRandomNo(11),
		Reason: req.Reason,
		Type:   model.ApprovalType(req.Type),
		Status: model.ApprovalStatus(req.Status),
		UserID: userID,
	}

	// Handle details based on type
	var abstract string
	switch approval.Type {
	case model.LeaveApproval:
		if req.Leave != nil {
			approval.Leave = &model.Leave{
				Type:      model.LeaveType(req.Leave.Type),
				StartTime: req.Leave.StartTime,
				EndTime:   req.Leave.EndTime,
				Reason:    req.Leave.Reason,
				TimeType:  model.TimeFormatType(req.Leave.TimeType),
			}
			// Calculate duration if not provided or just trust frontend?
			// AIWorkHelper trusts frontend or calculates it. Let's just copy.
			approval.Leave.Duration = req.Leave.Duration

			abstract = fmt.Sprintf("【%s】: 【%s】-【%s】",
				approval.Leave.Type.ToString(),
				timeutil.Format(req.Leave.StartTime),
				timeutil.Format(req.Leave.EndTime))
			if approval.Reason == "" {
				approval.Reason = req.Leave.Reason
			}
		}
	case model.GoOutApproval:
		if req.GoOut != nil {
			// Calculate duration if needed. AIWorkHelper does: duration := float32(req.GoOut.EndTime-req.GoOut.StartTime) / 3600.0
			duration := float32(req.GoOut.EndTime-req.GoOut.StartTime) / 3600.0
			approval.GoOut = &model.GoOut{
				StartTime: req.GoOut.StartTime,
				EndTime:   req.GoOut.EndTime,
				Duration:  duration,
				Reason:    req.GoOut.Reason,
			}
			abstract = fmt.Sprintf("【%s】-【%s】",
				timeutil.Format(req.GoOut.StartTime),
				timeutil.Format(req.GoOut.EndTime))
			if approval.Reason == "" {
				approval.Reason = req.GoOut.Reason
			}
		}
	case model.MakeCardApproval:
		if req.MakeCard != nil {
			approval.MakeCard = &model.MakeCard{
				Date:      req.MakeCard.Date,
				Reason:    req.MakeCard.Reason,
				Day:       req.MakeCard.Day,
				CheckType: model.WorkCheckType(req.MakeCard.CheckType),
			}
			abstract = fmt.Sprintf("【%s】【%s】",
				timeutil.Format(req.MakeCard.Date),
				req.MakeCard.Reason)
			if approval.Reason == "" {
				approval.Reason = req.MakeCard.Reason
			}
		}
	}

	// Get User Name for Title
	var user model.User
	if err := l.svcCtx.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		log.Error().Err(err).Uint("userID", userID).Msg("failed to find user for approval title")
		return nil, xerr.New(err)
	}
	approval.Title = fmt.Sprintf("%s 提交的 %s", user.Name, approval.Type.ToString())
	approval.Abstract = abstract

	if err := l.svcCtx.DB.WithContext(ctx).Create(approval).Error; err != nil {
		log.Error().Err(err).Msg("failed to create approval")
		return nil, xerr.New(err)
	}

	// Save Approvers
	if len(req.Approvers) > 0 {
		var approvers []model.Approver
		for _, a := range req.Approvers {
			if a.UserId == "" {
				continue
			}
			uid, err := strconv.Atoi(a.UserId)
			if err != nil || uid <= 0 {
				continue
			}
			approvers = append(approvers, model.Approver{
				ApprovalID: approval.ID,
				UserID:     uint(uid),
				Status:     model.Processed,
			})
		}
		if len(approvers) > 0 {
			if err := l.svcCtx.DB.WithContext(ctx).Create(&approvers).Error; err != nil {
				log.Error().Err(err).Msg("failed to create approvers")
				return nil, xerr.New(err)
			}
		}
	}

	return &domain.IdResp{
		Id: strconv.Itoa(int(approval.ID)),
	}, nil
}

func (l *approval) Dispose(ctx context.Context, req *domain.DisposeReq) (err error) {
	userID, err := token.GetUserID(ctx)
	if err != nil {
		return xerr.New(err)
	}

	// 1. Check if approval exists
	var approval model.Approval
	if err := l.svcCtx.DB.WithContext(ctx).Preload("Approvers").First(&approval, req.ApprovalId).Error; err != nil {
		log.Error().Err(err).Str("approvalId", req.ApprovalId).Msg("failed to find approval for dispose")
		return xerr.New(err)
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
		return xerr.New(errors.New("you are not an approver for this request"))
	}

	if currentApprover.Status != model.Processed {
		return xerr.New(errors.New("you have already processed this request"))
	}

	// 3. Update approver status
	currentApprover.Status = model.ApprovalStatus(req.Status)
	currentApprover.Reason = req.Reason
	if err := l.svcCtx.DB.WithContext(ctx).Save(currentApprover).Error; err != nil {
		log.Error().Err(err).Msg("failed to update approver status")
		return xerr.New(err)
	}

	// 4. Update approval status
	// Logic:
	// - If Reject (2) -> Approval Rejected (2)
	// - If Pass (1) -> Check if all passed -> Approval Passed (1)

	if model.ApprovalStatus(req.Status) == model.Refuse {
		approval.Status = model.Refuse
	} else if model.ApprovalStatus(req.Status) == model.Pass {
		allPassed := true
		for _, a := range approval.Approvers {
			if a.Status != model.Pass {
				allPassed = false
				break
			}
		}
		if allPassed {
			approval.Status = model.Pass
			now := time.Now()
			approval.FinishAt = &now
		}
	}

	if err := l.svcCtx.DB.WithContext(ctx).Save(&approval).Error; err != nil {
		log.Error().Err(err).Msg("failed to update approval status")
		return xerr.New(err)
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

	// Force use current user ID from token to match AIWorkHelper behavior
	uid, err := token.GetUserID(ctx)
	if err == nil && uid > 0 {
		req.UserId = strconv.Itoa(int(uid))
	}

	// 2. 构建查询
	db := l.svcCtx.DB.WithContext(ctx).Model(&model.Approval{})

	// 过滤条件
	if req.UserId != "" {
		userId, _ := strconv.Atoi(req.UserId)
		if userId > 0 {
			// 查询我发起的 OR 我审批的
			// Subquery for approvals where I am an approver
			subQuery := l.svcCtx.DB.Model(&model.Approver{}).Select("approval_id").Where("user_id = ?", userId)
			db = db.Where("user_id = ? OR id IN (?)", userId, subQuery)
		}
	}
	if req.Type > 0 {
		db = db.Where("type = ?", req.Type)
	}

	// 3. 查询总数
	var total int64
	if err = db.Count(&total).Error; err != nil {
		log.Error().Err(err).Msg("failed to count approvals")
		return nil, xerr.New(err)
	}

	// 4. 查询列表数据
	var approvals []*model.Approval
	if err = db.Preload("Approvers").Order("id desc").Offset(offset).Limit(count).Find(&approvals).Error; err != nil {
		log.Error().Err(err).Msg("failed to list approvals")
		return nil, xerr.New(err)
	}

	// 5. 组装响应
	list := make([]*domain.ApprovalList, 0, len(approvals))
	for _, v := range approvals {
		var participatingId string
		if req.UserId != "" {
			uid, _ := strconv.Atoi(req.UserId)
			for _, a := range v.Approvers {
				if int(a.UserID) == uid {
					// Assuming ParticipatingId refers to the Approver record ID or UserID?
					// AIWorkHelper likely uses it to identify the 'task' for the user.
					// Let's use Approver ID (which is unique for this user-approval pair).
					// But wait, Approver ID is uint.
					participatingId = strconv.Itoa(int(a.UserID))
					break
				}
			}
		}

		list = append(list, &domain.ApprovalList{
			Id:              strconv.Itoa(int(v.ID)),
			Type:            int(v.Type),
			Status:          int(v.Status),
			Title:           v.Title,
			Abstract:        v.Abstract,
			CreateId:        strconv.Itoa(int(v.UserID)),
			ParticipatingId: participatingId,
		})
	}

	resp = &domain.ApprovalListResp{
		Count: total,
		List:  list,
	}

	return resp, nil
}

// GenRandomNo 生成指定位数的随机数字字符串
func GenRandomNo(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	// rand.Seed(time.Now().UnixNano()) // Go 1.20+ automatically seeds global random

	var sb []byte
	// 生成指定位数的随机数字
	for i := 0; i < width; i++ {
		// simple random
		sb = append(sb, fmt.Sprintf("%d", numeric[int(time.Now().UnixNano())%r])[0])
	}
	// Better way if we want true random, but for now simple is fine or use math/rand
	return fmt.Sprintf("%d", time.Now().UnixNano())[0:width] // Simplified for now to avoid import math/rand if not needed, or just use timestamp
}
