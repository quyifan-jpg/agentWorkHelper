/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package ws 提供WebSocket聊天服务的核心实现
package ws

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/logic"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/token"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/dn-jinmin/tlog"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

// Ws WebSocket服务结构体，管理所有WebSocket连接和聊天功能
type Ws struct {
	websocket.Upgrader                     // WebSocket升级器，用于HTTP到WebSocket的协议升级
	svc                *svc.ServiceContext // 服务上下文，包含数据库连接等依赖

	uidToConn map[string]*websocket.Conn // 用户ID到WebSocket连接的映射
	ConnToUid map[*websocket.Conn]string // WebSocket连接到用户ID的映射

	sync.RWMutex              // 读写锁，保护连接映射的并发安全
	tokenParser  *token.Parse // JWT Token解析器，用于用户身份验证

	chat logic.Chat // 聊天业务逻辑处理器
}

// NewWs 创建WebSocket服务实例
func NewWs(svc *svc.ServiceContext) *Ws {
	// 初始化分布式链路追踪日志
	tlog.Init(
		tlog.WithLoggerWriter(tlog.NewLoggerWriter()),
		tlog.WithLabel(svc.Config.Tlog.Label),
		tlog.WithMode(svc.Config.Tlog.Mode),
	)

	return &Ws{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源的WebSocket连接（生产环境应该限制）
			},
		},
		svc:         svc,                                        // 注入服务上下文
		chat:        logic.NewChat(svc),                         // 初始化聊天业务逻辑
		tokenParser: token.NewTokenParse(svc.Config.Jwt.Secret), // 初始化JWT解析器

		uidToConn: make(map[string]*websocket.Conn), // 初始化用户ID到连接的映射
		ConnToUid: make(map[*websocket.Conn]string), // 初始化连接到用户ID的映射
	}
}

// Run 启动WebSocket服务
func (s *Ws) Run() {
	http.HandleFunc("/ws", s.ServerWs) // 注册WebSocket处理路由
	fmt.Println("启动websocket服务", s.svc.Config.Ws.Addr)
	http.ListenAndServe(s.svc.Config.Ws.Addr, nil) // 启动HTTP服务监听WebSocket连接
}

// ServerWs WebSocket连接处理入口
func (s *Ws) ServerWs(w http.ResponseWriter, r *http.Request) {
	// 异常恢复处理，防止panic导致服务崩溃
	defer func() {
		if e := recover(); e != nil {
			tlog.ErrorCtx(r.Context(), "serverWs", e)
		}
	}()

	// 用户身份验证，从请求头中解析JWT Token
	uid, token, err := s.auth(r)
	if err != nil {
		tlog.ErrorfCtx(r.Context(), "serverWs", "auth fail %v", err.Error())
		return
	}

	// 将HTTP连接升级为WebSocket连接
	respHeader := http.Header{
		"websocket": []string{token}, // 在响应头中返回Token
	}
	c, err := s.Upgrade(w, r, respHeader)
	if err != nil {
		tlog.ErrorfCtx(r.Context(), "serverWs", "Upgrade fail %v", err.Error())
		return
	}

	// 将新连接添加到连接管理器中
	s.addConn(c, uid)

	// 启动协程处理该连接的消息收发
	go s.handleConn(c, uid, token)
}

// addConn 添加WebSocket连接到管理器
func (s *Ws) addConn(conn *websocket.Conn, uid string) {
	s.RWMutex.Lock()         // 获取写锁，保证并发安全
	defer s.RWMutex.Unlock() // 函数结束时释放锁

	// 如果用户已有连接，先关闭旧连接（实现单点登录）
	if c := s.uidToConn[uid]; c != nil {
		c.Close()
	}

	// 建立双向映射关系
	s.uidToConn[uid] = conn // 用户ID -> 连接
	s.ConnToUid[conn] = uid // 连接 -> 用户ID
}

// closeConn 关闭WebSocket连接并清理相关资源
func (s *Ws) closeConn(conn *websocket.Conn) {
	s.RWMutex.Lock()         // 获取写锁，保证并发安全
	defer s.RWMutex.Unlock() // 函数结束时释放锁

	// 根据连接获取对应的用户ID
	uid := s.ConnToUid[conn]
	if uid == "" {
		return // 连接不存在，直接返回
	}

	fmt.Printf("关闭 %s 连接\n", uid)

	// 清理双向映射关系
	delete(s.ConnToUid, conn) // 删除连接到用户ID的映射
	delete(s.uidToConn, uid)  // 删除用户ID到连接的映射

	conn.Close() // 关闭WebSocket连接
}

// send 向指定WebSocket连接发送消息
func (s *Ws) send(ctx context.Context, conn *websocket.Conn, v interface{}) error {
	// 将消息对象序列化为JSON格式
	b, err := json.Marshal(v)
	if err != nil {
		tlog.ErrorCtx(ctx, "conn.send", err.Error())
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
		for i, _ := range s.uidToConn {
			if err := s.send(ctx, s.uidToConn[i], msg); err != nil {
				tlog.ErrorCtx(ctx, "sendByUids.all.send", err.Error())
				return err
			}
		}
		return nil
	}

	// 向指定的用户ID列表发送消息
	for _, uid := range uids {
		c, ok := s.uidToConn[uid]
		if !ok {
			continue // 用户不在线，跳过
		}
		if err := s.send(ctx, c, msg); err != nil {
			tlog.ErrorfCtx(ctx, "sendByUids.one.send err %v, uid %v", err.Error(), uid)
			return err
		}
	}
	return nil
}

// auth 用户身份验证，从HTTP请求头或URL参数中解析JWT Token
func (s *Ws) auth(r *http.Request) (uid string, tokenStr string, err error) {
	// 优先从请求头中获取WebSocket认证Token（保持原有功能）
	tok := r.Header.Get("websocket")

	// 如果请求头中没有token，尝试从URL参数中获取（新增兼容）
	if tok == "" {
		tok = r.URL.Query().Get("token")
	}

	// 如果两种方式都没有获取到token，返回错误
	if tok == "" {
		return "", "", errors.New("没有登入，不存在访问权限")
	}

	// 解析JWT Token，获取用户身份信息
	claims, tokenStr, err := s.tokenParser.ParseToken(tok)
	if err != nil {
		return "", "", err
	}

	// 从Token声明中提取用户ID
	return claims[token.Identify].(string), tokenStr, nil
}

// handleConn 处理WebSocket连接的消息收发循环
func (s *Ws) handleConn(conn *websocket.Conn, uid, tok string) {
	for {
		// 从WebSocket连接中读取客户端发送的消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// 连接异常或客户端断开连接
			tlog.Errorf("serverWs", "conn.ReadMessage fail %v, uid %v", err.Error(), uid)
			s.closeConn(conn) // 清理连接资源
			return
		}

		// 创建带有用户身份信息的上下文
		ctx := s.context(uid, tok)

		// 解析客户端发送的JSON消息
		var req domain.Message
		if err := json.Unmarshal(msg, &req); err != nil {
			tlog.ErrorfCtx(ctx, "handlerConn", "json.Unmarshal fail %v", err.Error())
			return
		}
		req.SendId = uid // 设置发送者ID（从Token中获取，防止伪造）

		// 根据聊天类型分发消息处理
		switch model.ChatType(req.ChatType) {
		case model.SingleChatType: // 私聊消息 (chatType = 2)
			err = s.privateChat(ctx, conn, &req)
		case model.GroupChatType: // 群聊消息 (chatType = 1)
			err = s.groupChat(ctx, conn, &req)
		}

		// 处理消息发送过程中的错误
		if err != nil {
			tlog.ErrorfCtx(ctx, "handlerConn", "message handle fail %v, msg %v", err.Error(), req)
			return
		}
	}
}

// context 创建带有用户身份信息的上下文
func (s *Ws) context(uid, tok string) context.Context {
	// 将用户ID注入到上下文中
	ctx := context.WithValue(context.Background(), token.Identify, uid)
	// 将JWT Token注入到上下文中
	ctx = context.WithValue(ctx, token.Authorization, tok)

	return tlog.TraceStart(ctx) // 开启分布式链路追踪日志
}
