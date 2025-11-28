/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package approval

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/curl"
	"AIWorkHelper/pkg/langchain"
	"AIWorkHelper/pkg/langchain/outputparserx"
	"AIWorkHelper/pkg/token"
	"AIWorkHelper/pkg/xerr"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

type Approval interface {
	Create(ctx context.Context, input string) (string, error)
}

var Approvals map[model.ApprovalType]Approval

func NewApproval(svc *svc.ServiceContext, approvalType model.ApprovalType) (Approval, error) {
	if Approvals == nil {
		Approvals = map[model.ApprovalType]Approval{
			model.LeaveApproval:    NewLeave(svc),
			model.MakeCardApproval: NewMakeCard(svc),
			model.GoOutApproval:    NewGoOut(svc),
		}
	}

	a := Approvals[approvalType]
	if a == nil {
		return nil, errors.New("不存在该审批类型" + fmt.Sprintf("%v", approvalType))
	}

	return a, nil
}

type MakeCard struct {
	svc          *svc.ServiceContext
	c            chains.Chain
	outputparser outputparserx.Structured
}

func NewMakeCard(svc *svc.ServiceContext) *MakeCard {
	output := outputparserx.NewStructured([]outputparserx.ResponseSchema{
		{
			Name:        "date",
			Description: "filling time,data application time stamp,  such as 1720921573",
			Type:        "int64",
		}, {
			Name:        "reason",
			Description: "reason for replacement card",
		}, {
			Name:        "day",
			Description: "replacement date, such as 20221011",
			Type:        "int64",
		}, {
			Name:        "workCheckType",
			Description: "replacement card type; enum: 1=上班卡(On-work check), 2=下班卡(Off-work check)",
			Type:        "int",
		},
	})
	return &MakeCard{
		svc: svc,
		c: chains.NewLLMChain(svc.LLMs, prompts.NewPromptTemplate(
			_defaultCreateApprovalTemplate+output.GetFormatInstructions(), []string{"input"},
		)),
		outputparser: output,
	}
}

func (m *MakeCard) Create(ctx context.Context, input string) (string, error) {
	out, err := chains.Predict(ctx, m.c, map[string]any{
		langchain.Input: input,
	})
	if err != nil {
		return "", err
	}

	v, err := m.outputparser.Parse(out)
	if err != nil {
		return "", err
	}

	var data domain.MakeCard
	if err := mapstructure.Decode(v, &data); err != nil {
		return "", err
	}

	req := domain.Approval{
		Type:     int(model.MakeCardApproval),
		MakeCard: &data,
	}
	addRes, err := curl.PostRequest(token.GetTokenStr(ctx), m.svc.Config.Host+"/v1/approval", req)
	if err != nil {
		return "", err
	}

	var idResp domain.IdRespInfo
	if err := json.Unmarshal(addRes, &idResp); err != nil {
		return "", xerr.WithMessage(err, "")
	}
	fmt.Println("创建补卡审批 结果 ", string(addRes))
	return idResp.Data.Id, err

}

type Leave struct {
	svc          *svc.ServiceContext
	c            chains.Chain
	outPutParser outputparserx.Structured
}

func NewLeave(svc *svc.ServiceContext) *Leave {
	output := outputparserx.NewStructured([]outputparserx.ResponseSchema{
		{
			Name:        "type",
			Description: "type of leave; Extract from the input which should contain this information. enum: 1=事假, 2=调休, 3=病假, 4=年假, 5=产假, 6=陪产假, 7=婚假, 8=丧假, 9=哺乳假",
			Type:        "int",
		}, {
			Name:        "startTime",
			Description: "leave start time in Unix timestamp (seconds). The input should already contain the timestamp value obtained from time_parser tool.",
			Type:        "int64",
		}, {
			Name:        "endTime",
			Description: "leave end time in Unix timestamp (seconds). The input should already contain the timestamp value obtained from time_parser tool.",
			Type:        "int64",
		}, {
			Name:        "reason",
			Description: "Reason for leave. Extract from input.",
		}, {
			Name:        "timeType",
			Description: "Leave duration type; enum: 1=小时(Hours), 2=天(Days). Calculate from startTime and endTime difference.",
			Type:        "int64",
		},
	})
	return &Leave{
		svc:          svc,
		outPutParser: output,
		c: chains.NewLLMChain(svc.LLMs, prompts.NewPromptTemplate(
			_defaultCreateApprovalTemplate+output.GetFormatInstructions(), []string{"input"},
		)),
	}
}

func (m *Leave) Create(ctx context.Context, input string) (string, error) {
	out, err := chains.Predict(ctx, m.c, map[string]any{
		langchain.Input: input,
	}, chains.WithCallback(m.svc.Callbacks))
	if err != nil {
		return "", xerr.WithMessage(err, "chains.Predict : "+input)
	}

	v, err := m.outPutParser.Parse(out)
	if err != nil {
		return "", xerr.WithMessage(err, "m.outPutParser.Parse")
	}

	var data domain.Leave
	if err := mapstructure.Decode(v, &data); err != nil {
		return "", xerr.WithMessage(err, "domain.GoOut")
	}

	req := domain.Approval{
		Type:  int(model.LeaveApproval),
		Leave: &data,
	}

	fmt.Println("提交请假审批 ： ", req, " \n ", req.Leave)

	addRes, err := curl.PostRequest(token.GetTokenStr(ctx), m.svc.Config.Host+"/v1/approval", req)
	var idResp domain.IdRespInfo
	if err := json.Unmarshal(addRes, &idResp); err != nil {
		return "", xerr.WithMessage(err, "")
	}
	fmt.Println("创建请假审批 结果 ", string(addRes))
	return idResp.Data.Id, err
}

type GoOut struct {
	svc          *svc.ServiceContext
	c            chains.Chain
	outPutParser outputparserx.Structured
}

func NewGoOut(svc *svc.ServiceContext) *MakeCard {
	output := outputparserx.NewStructured([]outputparserx.ResponseSchema{
		{
			Name:        "startTime",
			Description: "go out start time,data application time stamp, such as 1720921573",
			Type:        "int64",
		}, {
			Name:        "startTime",
			Description: "go out end time,data application time stamp, such as 1720921573",
			Type:        "int64",
		}, {
			Name:        "reason",
			Description: "Reason for go out",
		},
	})
	return &MakeCard{
		svc: svc,
		c: chains.NewLLMChain(svc.LLMs, prompts.NewPromptTemplate(
			_defaultCreateApprovalTemplate+output.GetFormatInstructions(), []string{"input"},
		)),
	}
}

func (m *GoOut) Create(ctx context.Context, input string) (string, error) {

	out, err := chains.Predict(ctx, m.c, map[string]any{
		langchain.Input: input,
	})
	if err != nil {
		return "", xerr.WithMessage(err, "chains.Predict : "+input)
	}

	v, err := m.outPutParser.Parse(out)
	if err != nil {
		return "", xerr.WithMessage(err, " m.outPutParser.Parse")
	}

	var data domain.GoOut
	if err := mapstructure.Decode(v, &data); err != nil {
		return "", xerr.WithMessage(err, "domain.GoOut")
	}

	req := domain.Approval{
		Type:  int(model.GoOutApproval),
		GoOut: &data,
	}

	addRes, err := curl.PostRequest(token.GetTokenStr(ctx), m.svc.Config.Host+"/v1/approval", req)
	var idResp domain.IdRespInfo
	if err := json.Unmarshal(addRes, &idResp); err != nil {
		return "", xerr.WithMessage(err, "")
	}
	fmt.Println("创建外出审批 结果 ", string(addRes))
	return idResp.Data.Id, err
}
