package ws

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/logic"
	"BackEnd/internal/model"
	"BackEnd/internal/svc"
	"BackEnd/pkg/jwt"
	"BackEnd/pkg/token"
	"BackEnd/pkg/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Ws WebSocket服务结构体，管理所有WebSocket连接和聊天功能
type Ws struct {
	websocket.Upgrader                     // WebSocket升级器，用于HTTP到WebSocket的协议升级
	svcCtx             *svc.ServiceContext // 服务上下文，包含数据库连接等依赖

	uidToConn map[string]*websocket.Conn // 用户ID到WebSocket连接的映射
	connToUid map[*websocket.Conn]string // WebSocket连接到用户ID的映射

	sync.RWMutex            // 读写锁，保护连接映射的并发安全
	chat         logic.Chat // 聊天业务逻辑处理器
	jwtSecret    string     // JWT密钥
}

// NewWs 创建WebSocket服务实例
func NewWs(svcCtx *svc.ServiceContext) *Ws {
	return &Ws{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源的WebSocket连接（生产环境应该限制）
			},
		},
		svcCtx:    svcCtx,
		chat:      logic.NewChat(svcCtx),
		jwtSecret: svcCtx.Config.Auth.Secret,
		uidToConn: make(map[string]*websocket.Conn),
		connToUid: make(map[*websocket.Conn]string),
	}
}

// Run 启动WebSocket服务
func (s *Ws) Run() {
	http.HandleFunc("/ws", s.ServerWs) // 注册WebSocket处理路由
	addr := fmt.Sprintf("%s:%d", s.svcCtx.Config.WS.Host, s.svcCtx.Config.WS.Port)
	log.Info().Str("addr", addr).Msg("启动 WebSocket 服务")
	fmt.Printf("启动 WebSocket 服务: ws://%s/ws\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("WebSocket 服务启动失败")
	}
}

// ServerWs WebSocket连接处理入口
func (s *Ws) ServerWs(w http.ResponseWriter, r *http.Request) {
	// 异常恢复处理，防止panic导致服务崩溃
	defer func() {
		if e := recover(); e != nil {
			log.Error().Interface("error", e).Msg("WebSocket 连接处理异常")
		}
	}()

	// 用户身份验证，从请求头或URL参数中解析JWT Token
	userID, tokenStr, err := s.auth(r)
	if err != nil {
		log.Error().Err(err).
			Str("remote_addr", r.RemoteAddr).
			Str("token_from", s.getTokenSource(r)).
			Msg("WebSocket 认证失败")
		// 返回 HTTP 401 错误，前端会收到连接失败
		http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
		return
	}

	// 将HTTP连接升级为WebSocket连接
	respHeader := http.Header{
		"websocket": []string{tokenStr}, // 在响应头中返回Token
	}
	conn, err := s.Upgrade(w, r, respHeader)
	if err != nil {
		log.Error().Err(err).
			Uint("userID", userID).
			Msg("WebSocket 升级失败")
		return
	}

	// log.Info().
	// 	Uint("userID", userID).
	// 	Msg("WebSocket 连接升级成功")

	// 将新连接添加到连接管理器中
	s.addConn(conn, util.UintToString(userID))

	// 启动协程处理该连接的消息收发
	go s.handleConn(conn, userID, tokenStr)
}

// getTokenSource 获取 Token 来源（用于日志）
func (s *Ws) getTokenSource(r *http.Request) string {
	if r.Header.Get("websocket") != "" {
		return "header"
	}
	if r.URL.Query().Get("token") != "" {
		return "query"
	}
	return "none"
}

// addConn 添加WebSocket连接到管理器
func (s *Ws) addConn(conn *websocket.Conn, uid string) {
	s.RWMutex.Lock()         // 获取写锁，保证并发安全
	defer s.RWMutex.Unlock() // 函数结束时释放锁

	// 如果用户已有连接，先关闭旧连接（实现单点登录）
	if oldConn := s.uidToConn[uid]; oldConn != nil {
		oldConn.Close()
		delete(s.connToUid, oldConn)
	}

	// 建立双向映射关系
	s.uidToConn[uid] = conn // 用户ID -> 连接
	s.connToUid[conn] = uid // 连接 -> 用户ID

	// log.Info().Str("uid", uid).Msg("WebSocket 连接已建立")
}

// closeConn 关闭WebSocket连接并清理相关资源
func (s *Ws) closeConn(conn *websocket.Conn) {
	s.RWMutex.Lock()         // 获取写锁，保证并发安全
	defer s.RWMutex.Unlock() // 函数结束时释放锁

	// 根据连接获取对应的用户ID
	uid := s.connToUid[conn]
	if uid == "" {
		return // 连接不存在，直接返回
	}

	// log.Info().Str("uid", uid).Msg("WebSocket 连接已关闭")

	// 清理双向映射关系
	delete(s.connToUid, conn) // 删除连接到用户ID的映射
	delete(s.uidToConn, uid)  // 删除用户ID到连接的映射

	conn.Close() // 关闭WebSocket连接
}

// send 向指定WebSocket连接发送消息
func (s *Ws) send(ctx context.Context, conn *websocket.Conn, v interface{}) error {
	// 将消息对象序列化为JSON格式
	b, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("序列化消息失败")
		return err
	}

	// 通过WebSocket连接发送文本消息
	return conn.WriteMessage(websocket.TextMessage, b)
}

// sendByUids 根据用户ID列表发送消息
// 如果uids为空，则广播给所有在线用户
func (s *Ws) sendByUids(ctx context.Context, msg interface{}, uids ...string) error {
	s.RWMutex.RLock()         // 获取读锁，允许并发读取
	defer s.RWMutex.RUnlock() // 函数结束时释放锁

	// 如果没有指定用户ID，则广播给所有在线用户
	if len(uids) == 0 {
		for uid, conn := range s.uidToConn {
			if err := s.send(ctx, conn, msg); err != nil {
				log.Error().Err(err).Str("uid", uid).Msg("广播消息失败")
				// 继续发送给其他用户，不中断
			}
		}
		return nil
	}

	// 向指定的用户ID列表发送消息
	for _, uid := range uids {
		conn, ok := s.uidToConn[uid]
		if !ok {
			continue // 用户不在线，跳过
		}
		if err := s.send(ctx, conn, msg); err != nil {
			log.Error().Err(err).Str("uid", uid).Msg("发送消息失败")
			// 继续发送给其他用户，不中断
		}
	}
	return nil
}

// auth 用户身份验证，从HTTP请求头或URL参数中解析JWT Token
func (s *Ws) auth(r *http.Request) (userID uint, tokenStr string, err error) {
	// 优先从请求头中获取WebSocket认证Token
	tok := r.Header.Get("websocket")

	// 如果请求头中没有token，尝试从URL参数中获取（兼容浏览器WebSocket API）
	if tok == "" {
		tok = r.URL.Query().Get("token")
	}

	// 如果两种方式都没有获取到token，返回错误
	if tok == "" {
		return 0, "", errors.New("没有登入，不存在访问权限：未提供Token")
	}

	// 解析JWT Token，获取用户身份信息
	userID, err = jwt.ParseToken(tok, s.jwtSecret)
	if err != nil {
		return 0, "", fmt.Errorf("token解析失败: %w", err)
	}

	return userID, tok, nil
}

// handleConn 处理WebSocket连接的消息收发循环
func (s *Ws) handleConn(conn *websocket.Conn, userID uint, tok string) {
	defer func() {
		s.closeConn(conn) // 确保连接关闭时清理资源
		// log.Info().Uint("userID", userID).Msg("WebSocket 连接处理结束")
	}()

	// log.Info().Uint("userID", userID).Msg("开始处理 WebSocket 连接")

	for {
		// 从WebSocket连接中读取客户端发送的消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// 连接异常或客户端断开连接
			// 检查是否是正常关闭
			return
		}

		// 创建带有用户身份信息的上下文
		ctx := s.context(userID, tok)

		// 解析客户端发送的JSON消息
		var req domain.Message
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Error().Err(err).
				Uint("userID", userID).
				Str("raw_message", string(msg)).
				Msg("解析WebSocket消息失败")
			// 发送错误消息给客户端
			errorMsg := map[string]interface{}{
				"error":  "消息格式错误",
				"detail": err.Error(),
			}
			if sendErr := s.send(ctx, conn, errorMsg); sendErr != nil {
				log.Error().Err(sendErr).Msg("发送错误消息失败")
			}
			continue // 继续处理下一条消息
		}
		req.SendId = util.UintToString(userID) // 设置发送者ID（从Token中获取，防止伪造）

		log.Debug().
			Uint("userID", userID).
			Int("chatType", req.ChatType).
			Str("type", req.Type).
			Str("recvId", req.RecvId).
			Msg("收到 WebSocket 消息")

		// 处理心跳消息
		if req.Type == "ping" {
			log.Debug().Uint("userID", userID).Msg("收到心跳包")
			continue
		}

		// 根据聊天类型分发消息处理
		switch model.ChatType(req.ChatType) {
		case model.SingleChatType: // 私聊消息 (chatType = 2)
			err = s.privateChat(ctx, conn, &req)
		case model.GroupChatType: // 群聊消息 (chatType = 1)
			err = s.groupChat(ctx, conn, &req)
		default:
			log.Warn().Int("chatType", req.ChatType).Msg("未知的聊天类型")
			// 发送错误消息给客户端
			errorMsg := map[string]interface{}{
				"error":    "不支持的聊天类型",
				"chatType": req.ChatType,
			}
			if sendErr := s.send(ctx, conn, errorMsg); sendErr != nil {
				log.Error().Err(sendErr).Msg("发送错误消息失败")
			}
			continue
		}

		// 处理消息发送过程中的错误
		if err != nil {
			log.Error().Err(err).
				Uint("userID", userID).
				Interface("msg", req).
				Msg("处理消息失败")
			// 发送错误消息给客户端
			errorMsg := map[string]interface{}{
				"error":  "消息处理失败",
				"detail": err.Error(),
			}
			if sendErr := s.send(ctx, conn, errorMsg); sendErr != nil {
				log.Error().Err(sendErr).Msg("发送错误消息失败")
			}
			// 不返回，继续处理下一条消息
		}
	}
}

// context 创建带有用户身份信息的上下文
func (s *Ws) context(userID uint, tok string) context.Context {
	// 将用户ID注入到上下文中
	ctx := token.SetUserID(context.Background(), userID)
	return ctx
}

// privateChat 处理私聊消息
// 将消息保存到数据库并发送给指定的接收者
func (s *Ws) privateChat(ctx context.Context, conn *websocket.Conn, req *domain.Message) error {
	// 调用聊天业务逻辑，保存私聊消息到数据库
	if err := s.chat.PrivateChat(ctx, req); err != nil {
		return err
	}

	// 将消息发送给接收者（注意：当前逻辑发送者看不到自己发送的消息）
	if req.RecvId != "" {
		return s.sendByUids(ctx, req, req.RecvId)
	}
	return nil
}

// groupChat 处理群聊消息
// 将消息保存到数据库并广播给所有在线用户（或群成员）
func (s *Ws) groupChat(ctx context.Context, conn *websocket.Conn, req *domain.Message) error {
	// 调用聊天业务逻辑，保存群聊消息到数据库
	_, err := s.chat.GroupChat(ctx, req)
	if err != nil {
		return err
	}

	// 群聊：广播给所有在线用户（排除发送者）
	// TODO: 后续可以实现根据群ID查询群成员，只发送给群成员
	// 当前实现：广播给所有在线用户
	return s.sendByUids(ctx, req) // 广播给所有在线用户
}
