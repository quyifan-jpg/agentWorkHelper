/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package api

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/logic"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/httpx"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

// Upload 处理文件上传相关的HTTP请求
type Upload struct {
	svcCtx *svc.ServiceContext
	chat   logic.Chat // 聊天逻辑，用于将文件信息写入记忆机制
}

// NewUpload 创建文件上传处理器实例
func NewUpload(svcCtx *svc.ServiceContext, chat logic.Chat) *Upload {
	return &Upload{
		svcCtx: svcCtx,
		chat:   chat,
	}
}

// InitRegister 注册文件上传相关的路由
func (h *Upload) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/upload", h.svcCtx.Jwt.Handler) // 使用JWT中间件保护上传接口
	g.POST("/file", h.File)                              // 单文件上传接口
	g.POST("/multiplefiles", h.Multiplefiles)            // 多文件上传接口
}

// File 处理单个文件上传请求
func (h *Upload) File(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file") // 从请求中获取上传的文件
	var (
		filename string
		buf      = bytes.NewBuffer(nil) // 用于缓存文件内容
	)
	defer file.Close()

	if _, err := io.Copy(buf, file); err != nil { // 读取文件内容到缓冲区
		httpx.FailWithErr(ctx, err)
		return
	}

	filename = ksuid.New().String() + filepath.Ext(header.Filename) // 生成唯一文件名

	newFile, err := os.Create(h.svcCtx.Config.Upload.SavePath + filename) // 创建目标文件
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	defer newFile.Close()

	if _, err := newFile.Write(buf.Bytes()); err != nil { // 写入文件内容
		httpx.FailWithErr(ctx, err)
		return
	}

	resp := domain.FileResp{
		Host:     h.svcCtx.Config.Host,
		File:     fmt.Sprintf("%s%s", h.svcCtx.Config.Upload.SavePath, filename),
		Filename: filename,
	}

	chat := ctx.Request.FormValue("chat")
	if len(chat) > 0 { // 如果指定了chat参数，将文件信息写入记忆机制
		h.chat.File(ctx.Request.Context(), []*domain.FileResp{
			&resp,
		})
	}

	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, resp) // 返回文件上传成功响应
	}
}

// Multiplefiles 处理多文件上传请求（待实现）
func (h *Upload) Multiplefiles(ctx *gin.Context) {
}
