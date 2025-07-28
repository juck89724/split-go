package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"split-go/internal/config"
	"split-go/internal/middleware"
	"split-go/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JWTService JWT 服務
type JWTService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewJWTService 創建新的 JWT 服務
func NewJWTService(db *gorm.DB, cfg *config.Config) *JWTService {
	return &JWTService{
		db:  db,
		cfg: cfg,
	}
}

// DeviceFingerprint 設備指紋結構
type DeviceFingerprint struct {
	UserAgent string         `json:"user_agent"`
	Language  string         `json:"language"`
	TimeZone  string         `json:"timezone"`
	Screen    map[string]int `json:"screen"`
	Platform  string         `json:"platform"`
	IPAddress string         `json:"ip_address"`
}

// EnterpriseTokens 企業級 Token 響應
type EnterpriseTokens struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	DeviceToken  string             `json:"device_token"`
	ExpiresIn    int64              `json:"expires_in"`
	User         models.User        `json:"user"`
	Session      models.UserSession `json:"session"`
}

// HashToken 計算 token 的 SHA256 哈希值
func (s *JWTService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateDeviceID 生成設備唯一識別碼
func (s *JWTService) GenerateDeviceID(fingerprint DeviceFingerprint) string {
	data := fmt.Sprintf("%s_%s_%s_%s",
		fingerprint.UserAgent,
		fingerprint.Platform,
		fingerprint.Language,
		fingerprint.IPAddress)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // 取前16字節
}

// CreateSession 創建用戶會話
func (s *JWTService) CreateSession(user models.User, deviceID string, fingerprint DeviceFingerprint, deviceName, deviceType string) (*models.UserSession, error) {
	fingerprintJSON, _ := json.Marshal(fingerprint)

	session := models.UserSession{
		ID:                 uuid.New().String(),
		UserID:             user.ID,
		DeviceID:           deviceID,
		DeviceFingerprint:  fingerprintJSON,
		DeviceName:         deviceName,
		DeviceType:         deviceType,
		UserAgent:          fingerprint.UserAgent,
		IPAddress:          fingerprint.IPAddress,
		TrustLevel:         0, // 新設備默認為不可信
		LastActivity:       time.Now(),
		ExpiresAt:          time.Now().Add(s.cfg.DeviceTokenDuration), // 使用配置的 Device Token 過期時間
		AccessTokenVersion: 1,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// SignToken 簽名 JWT token
func (s *JWTService) SignToken(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateEnterpriseTokens 生成企業級三層級 tokens
func (s *JWTService) GenerateEnterpriseTokens(user models.User, session *models.UserSession) (*EnterpriseTokens, error) {
	now := time.Now()

	// Access Token - 使用配置的過期時間
	accessClaims := &middleware.AccessTokenClaims{
		UserID:       user.ID,
		Email:        user.Email,
		SessionID:    session.ID,
		TokenVersion: session.AccessTokenVersion,
		DeviceID:     session.DeviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "split-go-enterprise",
		},
	}

	// Refresh Token - 使用配置的過期時間
	refreshClaims := &middleware.RefreshTokenClaims{
		UserID:    user.ID,
		SessionID: session.ID,
		DeviceID:  session.DeviceID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "split-go-enterprise",
		},
	}

	// Device Token - 使用配置的過期時間
	deviceClaims := &middleware.DeviceTokenClaims{
		UserID:    user.ID,
		DeviceID:  session.DeviceID,
		TokenType: "device_auth",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.DeviceTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "split-go-enterprise",
		},
	}

	// 生成 tokens
	accessToken, err := s.SignToken(accessClaims, s.cfg.AccessTokenSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.SignToken(refreshClaims, s.cfg.RefreshTokenSecret)
	if err != nil {
		return nil, err
	}

	deviceToken, err := s.SignToken(deviceClaims, s.cfg.DeviceTokenSecret)
	if err != nil {
		return nil, err
	}

	// 更新會話的 refresh token hash
	refreshTokenHash := s.HashToken(refreshToken)
	session.RefreshTokenHash = refreshTokenHash
	if err := s.db.Save(session).Error; err != nil {
		return nil, err
	}

	// 將 user 資料填入 session 物件中
	session.User = user

	return &EnterpriseTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceToken:  deviceToken,
		ExpiresIn:    int64(s.cfg.AccessTokenDuration.Seconds()), // 使用配置的過期時間
		User:         user,
		Session:      *session,
	}, nil
}

// ValidateRefreshToken 驗證 refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*middleware.RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &middleware.RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的簽名方法")
		}
		return []byte(s.cfg.RefreshTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*middleware.RefreshTokenClaims)
	if !ok || !token.Valid || claims.TokenType != "refresh" {
		return nil, fmt.Errorf("無效的 refresh token")
	}

	return claims, nil
}

// ValidateDeviceToken 驗證 device token
func (s *JWTService) ValidateDeviceToken(tokenString string) (*middleware.DeviceTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &middleware.DeviceTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的簽名方法")
		}
		return []byte(s.cfg.DeviceTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*middleware.DeviceTokenClaims)
	if !ok || !token.Valid || claims.TokenType != "device_auth" {
		return nil, fmt.Errorf("無效的 device token")
	}

	return claims, nil
}

// GetSession 獲取會話信息
func (s *JWTService) GetSession(sessionID string) (*models.UserSession, error) {
	var session models.UserSession
	if err := s.db.Where("id = ? AND revoked_at IS NULL", sessionID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSessionActivity 更新會話活動時間
func (s *JWTService) UpdateSessionActivity(session *models.UserSession, ipAddress string) error {
	session.LastActivity = time.Now()
	session.IPAddress = ipAddress
	session.AccessTokenVersion++ // 讓舊的 access token 失效
	return s.db.Save(session).Error
}

// RevokeSession 撤銷會話
func (s *JWTService) RevokeSession(sessionID string, userID uint) error {
	now := time.Now()
	result := s.db.Model(&models.UserSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("revoked_at", &now)

	if result.RowsAffected == 0 {
		return fmt.Errorf("會話不存在")
	}

	return nil
}

// LogSecurityEvent 記錄安全事件
func (s *JWTService) LogSecurityEvent(userID uint, sessionID, eventType, ipAddress string, eventData map[string]interface{}) error {
	eventDataJSON, _ := json.Marshal(eventData)

	event := models.SecurityEvent{
		UserID:    userID,
		SessionID: sessionID,
		EventType: eventType,
		EventData: eventDataJSON,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}

	return s.db.Create(&event).Error
}

// DeviceInfo 設備信息結構
type DeviceInfo struct {
	ID           string    `json:"id"`
	DeviceID     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	LastActivity time.Time `json:"last_activity"`
	IPAddress    string    `json:"ip_address"`
	Location     string    `json:"location"`
	IsCurrent    bool      `json:"is_current"`
	TrustLevel   int       `json:"trust_level"`
	CreatedAt    time.Time `json:"created_at"`
}

// HandleLogin 處理用戶登入邏輯
func (s *JWTService) HandleLogin(user models.User, deviceFingerprint DeviceFingerprint, deviceName, deviceType, ipAddress string) (*EnterpriseTokens, error) {
	// 生成設備ID
	deviceID := s.GenerateDeviceID(deviceFingerprint)

	// 檢查是否是現有設備
	var existingSession models.UserSession
	isNewDevice := s.db.Where("user_id = ? AND device_id = ? AND revoked_at IS NULL",
		user.ID, deviceID).First(&existingSession).Error != nil

	// 創建或更新會話
	var session *models.UserSession
	var err error

	if isNewDevice {
		session, err = s.CreateSession(user, deviceID, deviceFingerprint, deviceName, deviceType)
		if err != nil {
			return nil, err
		}

		s.LogSecurityEvent(user.ID, session.ID, "new_device_login", ipAddress, map[string]interface{}{
			"device_id":   deviceID,
			"device_name": deviceName,
		})
	} else {
		// 更新現有會話
		existingSession.LastActivity = time.Now()
		existingSession.IPAddress = ipAddress
		if err := s.db.Save(&existingSession).Error; err != nil {
			return nil, err
		}
		session = &existingSession

		s.LogSecurityEvent(user.ID, session.ID, "login", ipAddress, map[string]interface{}{
			"device_id": deviceID,
		})
	}

	// 生成 tokens
	return s.GenerateEnterpriseTokens(user, session)
}

// HandleRefresh 處理 token 刷新邏輯
func (s *JWTService) HandleRefresh(refreshTokenString, ipAddress string) (*EnterpriseTokens, error) {
	// 驗證 refresh token
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// 獲取會話
	session, err := s.GetSession(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("會話不存在或已失效")
	}

	// 驗證 refresh token hash
	currentTokenHash := s.HashToken(refreshTokenString)
	if session.RefreshTokenHash != currentTokenHash {
		s.LogSecurityEvent(session.UserID, session.ID, "suspicious_refresh", ipAddress, map[string]interface{}{
			"reason": "token_hash_mismatch",
		})
		return nil, fmt.Errorf("refresh token 已被撤銷")
	}

	// 更新會話活動時間
	if err := s.UpdateSessionActivity(session, ipAddress); err != nil {
		return nil, err
	}

	// 獲取用戶信息
	var user models.User
	if err := s.db.First(&user, session.UserID).Error; err != nil {
		return nil, fmt.Errorf("用戶不存在")
	}

	// 生成新的 tokens
	tokens, err := s.GenerateEnterpriseTokens(user, session)
	if err != nil {
		return nil, err
	}

	s.LogSecurityEvent(session.UserID, session.ID, "token_refresh", ipAddress, nil)
	return tokens, nil
}

// HandleDeviceRefresh 處理設備認證刷新邏輯
func (s *JWTService) HandleDeviceRefresh(deviceTokenString string, deviceFingerprint DeviceFingerprint, ipAddress string) (*EnterpriseTokens, error) {
	// 驗證 device token
	claims, err := s.ValidateDeviceToken(deviceTokenString)
	if err != nil {
		return nil, err
	}

	// 驗證設備指紋
	expectedDeviceID := s.GenerateDeviceID(deviceFingerprint)
	if expectedDeviceID != claims.DeviceID {
		s.LogSecurityEvent(claims.UserID, "", "device_fingerprint_mismatch", ipAddress, map[string]interface{}{
			"expected_device_id": expectedDeviceID,
			"claimed_device_id":  claims.DeviceID,
		})
		return nil, fmt.Errorf("設備指紋不匹配")
	}

	// 查找會話
	var session models.UserSession
	err = s.db.Where("user_id = ? AND device_id = ? AND revoked_at IS NULL",
		claims.UserID, claims.DeviceID).First(&session).Error

	if err != nil {
		return nil, fmt.Errorf("會話已失效，請重新登入")
	}

	// 更新會話
	session.LastActivity = time.Now()
	session.IPAddress = ipAddress
	session.AccessTokenVersion++
	if err := s.db.Save(&session).Error; err != nil {
		return nil, err
	}

	// 獲取用戶信息
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, fmt.Errorf("用戶不存在")
	}

	// 生成新的 tokens
	tokens, err := s.GenerateEnterpriseTokens(user, &session)
	if err != nil {
		return nil, err
	}

	s.LogSecurityEvent(session.UserID, session.ID, "device_refresh", ipAddress, nil)
	return tokens, nil
}

// HandleLogout 處理用戶登出邏輯
func (s *JWTService) HandleLogout(sessionID string, userID uint, ipAddress string) error {
	// 撤銷會話
	if err := s.RevokeSession(sessionID, userID); err != nil {
		return err
	}

	s.LogSecurityEvent(userID, sessionID, "logout", ipAddress, nil)
	return nil
}

// GetUserDevices 獲取用戶所有設備
func (s *JWTService) GetUserDevices(userID uint, currentSessionID string) ([]DeviceInfo, error) {
	var sessions []models.UserSession
	if err := s.db.Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("last_activity DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	devices := make([]DeviceInfo, len(sessions))
	for i, session := range sessions {
		location := fmt.Sprintf("%s, %s", session.City, session.Country)
		if location == ", " {
			location = "未知位置"
		}

		devices[i] = DeviceInfo{
			ID:           session.ID,
			DeviceID:     session.DeviceID,
			DeviceName:   session.DeviceName,
			DeviceType:   session.DeviceType,
			LastActivity: session.LastActivity,
			IPAddress:    session.IPAddress,
			Location:     location,
			IsCurrent:    session.ID == currentSessionID,
			TrustLevel:   session.TrustLevel,
			CreatedAt:    session.CreatedAt,
		}
	}

	return devices, nil
}

// RevokeDevice 撤銷指定設備
func (s *JWTService) RevokeDevice(deviceID string, userID uint, currentSessionID, ipAddress string) error {
	// 檢查設備是否存在
	var targetSession models.UserSession
	if err := s.db.Where("id = ? AND user_id = ?", deviceID, userID).First(&targetSession).Error; err != nil {
		return fmt.Errorf("設備不存在")
	}

	// 撤銷設備
	now := time.Now()
	if err := s.db.Model(&targetSession).Update("revoked_at", &now).Error; err != nil {
		return err
	}

	// 記錄安全事件
	s.LogSecurityEvent(userID, targetSession.ID, "device_revoked", ipAddress, map[string]interface{}{
		"revoked_device_id":   deviceID,
		"revoked_by_session":  currentSessionID,
		"revoked_device_name": targetSession.DeviceName,
	})

	return nil
}

// GetSecurityEvents 獲取安全事件記錄
func (s *JWTService) GetSecurityEvents(userID uint) ([]models.SecurityEvent, error) {
	var events []models.SecurityEvent
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}
