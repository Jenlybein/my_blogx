// Package user_service 用户认证与会话管理核心服务包
// 负责：登录令牌生成、访问令牌校验、刷新令牌、会话吊销、用户状态校验等核心鉴权逻辑
package user_service

import (
	"time"

	"myblogx/core"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/redis_service/redis_jwt"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

// refreshTokenCookieName 刷新令牌在Cookie中的键名
const refreshTokenCookieName = "refresh_token"

type SessionMeta struct {
	IP   string
	Addr string
	UA   string
}

type AuthResult struct {
	Token   string                   // 原始访问令牌
	Claims  *jwts.MyClaims           // 解析后的JWT自定义声明
	User    *models.UserModel        // 对应用户信息
	Session *models.UserSessionModel // 对应用户会话信息
}

// BuildSessionMetaFromGin 从Gin上下文构建会话元数据
// 用于登录、刷新令牌时记录用户设备、IP信息
func BuildSessionMetaFromGin(c *gin.Context) SessionMeta {
	if c == nil || c.Request == nil {
		return SessionMeta{}
	}
	// 获取客户端真实IP
	ip := c.ClientIP()
	return SessionMeta{
		IP:   ip,
		Addr: core.GetIpAddr(ip),
		UA:   c.Request.UserAgent(),
	}
}

// AuthenticateAccessTokenByGin 从Gin上下文提取Token并完成认证
func AuthenticateAccessTokenByGin(c *gin.Context) (*AuthResult, error) {
	token := jwts.GetTokenByGin(c)
	return AuthenticateAccessToken(token)
}

// MustAuthenticateAccessTokenByGin 尝试认证AccessToken
// 不抛出错误，失败直接返回nil，用于可选鉴权接口
func MustAuthenticateAccessTokenByGin(c *gin.Context) *AuthResult {
	token := jwts.GetTokenByGin(c)
	if token == "" {
		return nil
	}
	result, err := AuthenticateAccessToken(token)
	if err != nil {
		return nil
	}
	return result
}

// AuthenticateAccessToken 核心方法：校验访问令牌（AccessToken）合法性
// 流程：解析Token → 校验黑名单 → 查用户 → 校验状态 → 校验令牌版本 → 校验会话
func AuthenticateAccessToken(token string) (*AuthResult, error) {
	if token == "" {
		return nil, ErrAuthRequired
	}

	// 解析令牌
	claims, err := jwts.ParseToken(token)
	if err != nil || claims.SessionID == 0 {
		return nil, ErrAuthInvalid
	}

	// 校验令牌是否在Redis黑名单中
	if _, ok := redis_jwt.HasTokenBlack(token); !ok {
		return nil, ErrAuthInvalid
	}

	// 查询用户是否存在，并校验令牌版本
	var user models.UserModel
	if err = global.DB.Take(&user, claims.UserID).Error; err != nil || !user.CheckTokenVersion(claims.TokenVersion) {
		return nil, ErrAuthInvalid
	}

	// 校验用户状态
	if err = user.ValidateUserStatus(); err != nil {
		return nil, err
	}

	// 校验会话是否有效（未吊销、未过期、归属正确）
	session, err := getSession(claims.SessionID, claims.UserID)
	if err != nil {
		return nil, ErrAuthInvalid
	}

	// 补充声明信息
	claims.Role = user.Role
	claims.Username = user.Username
	claims.TokenVersion = user.TokenVersion

	// 认证成功，返回完整结果
	return &AuthResult{
		Token:   token,
		Claims:  claims,
		User:    &user,
		Session: session,
	}, nil
}

// AuthenticateSession 根据userID和sessionID直接校验会话合法性
// 用于内部服务、跨服务鉴权
func AuthenticateSession(userID, sessionID ctype.ID) (*AuthResult, error) {
	if userID == 0 || sessionID == 0 {
		return nil, ErrAuthInvalid
	}

	// 查询用户是否存在
	var user models.UserModel
	if err := global.DB.Take(&user, userID).Error; err != nil {
		return nil, ErrAuthInvalid
	}
	// 校验用户状态（正常/禁用/封禁）
	if err := user.ValidateUserStatus(); err != nil {
		return nil, err
	}

	// 查询会话：必须有效、未吊销、未过期
	session, err := getSession(sessionID, userID)
	if err != nil {
		return nil, ErrAuthInvalid
	}

	// 构造认证结果并返回
	return &AuthResult{
		Claims: &jwts.MyClaims{
			Claims: jwts.Claims{
				UserID:       user.ID,
				SessionID:    session.ID,
				TokenVersion: user.TokenVersion,
				Username:     user.Username,
				Role:         user.Role,
			},
		},
		User:    &user,
		Session: session,
	}, nil
}

// 校验会话是否有效（未吊销、未过期、归属正确）
func getSession(sessionID, userID ctype.ID) (*models.UserSessionModel, error) {
	var session models.UserSessionModel
	if err := global.DB.
		Where("id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", sessionID, userID, time.Now()).
		Take(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}
