package api

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/svc"
	"BackEnd/pkg/httpx"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

type Upload struct {
	svcCtx *svc.ServiceContext
	chat   logic.Chat
}

func NewUpload(svcCtx *svc.ServiceContext, chat logic.Chat) *Upload {
	return &Upload{
		svcCtx: svcCtx,
		chat:   chat,
	}
}

func (h *Upload) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/upload", h.svcCtx.Jwt.Handler)
	g.POST("/file", h.File)
	g.POST("/multiplefiles", h.Multiplefiles)
}

// File 上传单个文件
// @Summary 上传单个文件
// @Description 上传文件并保存文件信息，可选参数 chat 用于将文件信息写入 AI 记忆
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param chat formData string false "是否写入 AI 记忆（任意值）"
// @Success 200 {object} object{code=int,msg=string,data=domain.FileResp}
// @Router /v1/upload/file [post]
func (h *Upload) File(ctx *gin.Context) {
	// 从请求中获取上传的文件
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	defer file.Close()

	var (
		filename string
		buf      = bytes.NewBuffer(nil) // 用于缓存文件内容
	)

	// 读取文件内容到缓冲区
	if _, err := io.Copy(buf, file); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 生成唯一文件名：使用 ksuid + 原文件扩展名
	filename = ksuid.New().String() + filepath.Ext(header.Filename)

	// 确保上传目录存在
	savePath := h.svcCtx.Config.Upload.SavePath
	if err := os.MkdirAll(savePath, 0755); err != nil {
		httpx.FailWithErr(ctx, fmt.Errorf("创建上传目录失败: %w", err))
		return
	}

	// 创建目标文件
	newFile, err := os.Create(savePath + filename)
	if err != nil {
		httpx.FailWithErr(ctx, fmt.Errorf("创建文件失败: %w", err))
		return
	}
	defer newFile.Close()

	// 写入文件内容
	if _, err := newFile.Write(buf.Bytes()); err != nil {
		httpx.FailWithErr(ctx, fmt.Errorf("写入文件失败: %w", err))
		return
	}

	// 构建响应
	resp := domain.FileResp{
		Host:     h.svcCtx.Config.Upload.Host,
		File:     fmt.Sprintf("%s%s", savePath, filename),
		Filename: filename,
	}

	// 如果指定了 chat 参数，将文件信息写入 AI 记忆机制
	chat := ctx.Request.FormValue("chat")
	if len(chat) > 0 {
		if err := h.chat.File(ctx.Request.Context(), []*domain.FileResp{&resp}); err != nil {
			// 文件已保存，即使写入记忆失败也不影响上传成功
			// 可以记录日志，但不返回错误
			fmt.Printf("警告: 文件信息写入 AI 记忆失败: %v\n", err)
		}
	}

	httpx.Success(ctx, resp)
}

// Multiplefiles 批量上传文件
// @Summary 批量上传文件
// @Description 批量上传多个文件，可选参数 chat 用于将文件信息写入 AI 记忆
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData []file true "文件列表"
// @Param chat formData string false "是否写入 AI 记忆（任意值）"
// @Success 200 {object} object{code=int,msg=string,data=domain.FileListResp}
// @Router /v1/upload/multiplefiles [post]
func (h *Upload) Multiplefiles(ctx *gin.Context) {
	// 获取表单
	form, err := ctx.MultipartForm()
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		httpx.FailWithErr(ctx, fmt.Errorf("未找到上传的文件"))
		return
	}

	// 确保上传目录存在
	savePath := h.svcCtx.Config.Upload.SavePath
	if err := os.MkdirAll(savePath, 0755); err != nil {
		httpx.FailWithErr(ctx, fmt.Errorf("创建上传目录失败: %w", err))
		return
	}

	var fileList []*domain.FileResp

	// 处理每个文件
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue // 跳过无法打开的文件
		}

		var buf = bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			file.Close()
			continue // 跳过读取失败的文件
		}
		file.Close()

		// 生成唯一文件名
		filename := ksuid.New().String() + filepath.Ext(fileHeader.Filename)

		// 创建目标文件
		newFile, err := os.Create(savePath + filename)
		if err != nil {
			continue // 跳过创建失败的文件
		}

		// 写入文件内容
		if _, err := newFile.Write(buf.Bytes()); err != nil {
			newFile.Close()
			continue // 跳过写入失败的文件
		}
		newFile.Close()

		// 添加到响应列表
		fileList = append(fileList, &domain.FileResp{
			Host:     h.svcCtx.Config.Upload.Host,
			File:     fmt.Sprintf("%s%s", savePath, filename),
			Filename: filename,
		})
	}

	if len(fileList) == 0 {
		httpx.FailWithErr(ctx, fmt.Errorf("所有文件上传失败"))
		return
	}

	// 如果指定了 chat 参数，将文件信息写入 AI 记忆机制
	chat := ctx.Request.FormValue("chat")
	if len(chat) > 0 {
		if err := h.chat.File(ctx.Request.Context(), fileList); err != nil {
			// 文件已保存，即使写入记忆失败也不影响上传成功
			fmt.Printf("警告: 文件信息写入 AI 记忆失败: %v\n", err)
		}
	}

	resp := domain.FileListResp{
		List: fileList,
	}

	httpx.Success(ctx, resp)
}
