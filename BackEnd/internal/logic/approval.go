package logic

import (
	"context"
	"strconv"

	"BackEnd/internal/domain"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
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
	return
}

func (l *approval) Dispose(ctx context.Context, req *domain.DisposeReq) (err error) {
	return
}

func (l *approval) List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error) {
	return
}
