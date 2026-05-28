package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrInvalidToken = errors.New("invalid token")
var ErrDictItemAlreadyExists = errors.New("dict item already exists")
var ErrDictItemNotFound = errors.New("dict item not found")
var ErrRegistrationClosed = errors.New("registration closed")
var ErrUserAlreadyExists = errors.New("user already exists")

type AuthService struct {
	users         dao.UserDAO
	dictItems     dao.DictItemDAO
	operationLogs dao.OperationLogDAO
	jwtSecret     []byte
	tokenDuration time.Duration
	revokedTokens map[string]time.Time
	revokedMu     sync.Mutex
}

func NewAuthService(
	users dao.UserDAO,
	dictItems dao.DictItemDAO,
	operationLogs dao.OperationLogDAO,
	jwtSecret string,
	tokenDuration time.Duration,
) *AuthService {
	return &AuthService{
		users:         users,
		dictItems:     dictItems,
		operationLogs: operationLogs,
		jwtSecret:     []byte(jwtSecret),
		tokenDuration: tokenDuration,
		revokedTokens: make(map[string]time.Time),
	}
}

func (svc *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AdminUser, error) {
	count, err := svc.users.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrRegistrationClosed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := svc.users.Create(ctx, model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
		Email:        req.Email,
		Status:       "active",
	})
	if err != nil {
		if errors.Is(err, dao.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return toAdminUser(user), nil
}

func (svc *AuthService) Login(ctx context.Context, req dto.LoginRequest, meta dto.OperationMeta) (*dto.LoginResponse, error) {
	user, err := svc.users.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive() {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := svc.issueToken(user)
	if err != nil {
		return nil, err
	}

	if err := svc.users.UpdateLastLoginAt(ctx, user.ID); err != nil {
		return nil, err
	}
	svc.safeRecordOperation(ctx, user, 1, meta)

	return &dto.LoginResponse{
		AccessToken: token,
		ExpiresIn:   int64(svc.tokenDuration.Seconds()),
		User:        *toAdminUser(user),
	}, nil
}

func (svc *AuthService) Me(ctx context.Context, authorization string) (*dto.AdminUser, error) {
	user, err := svc.authenticate(ctx, authorization)
	if err != nil {
		return nil, err
	}

	return toAdminUser(user), nil
}

func (svc *AuthService) Refresh(ctx context.Context, authorization string) (*dto.LoginResponse, error) {
	claims, err := svc.parseAuthorization(authorization)
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID, err := parseSubject(claims["sub"])
	if err != nil {
		return nil, ErrInvalidToken
	}
	user, err := svc.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if !user.IsActive() {
		return nil, ErrInvalidToken
	}

	token, err := svc.issueToken(user)
	if err != nil {
		return nil, err
	}
	svc.revokeClaims(claims)

	return &dto.LoginResponse{
		AccessToken: token,
		ExpiresIn:   int64(svc.tokenDuration.Seconds()),
		User:        *toAdminUser(user),
	}, nil
}

func (svc *AuthService) Logout(ctx context.Context, authorization string, meta dto.OperationMeta) error {
	claims, err := svc.parseAuthorization(authorization)
	if err != nil {
		return ErrInvalidToken
	}

	userID, err := parseSubject(claims["sub"])
	if err != nil {
		return ErrInvalidToken
	}
	user, err := svc.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			return ErrInvalidToken
		}
		return err
	}
	svc.revokeClaims(claims)
	svc.safeRecordOperation(ctx, user, 2, meta)
	return nil
}

func (svc *AuthService) CreateDictItem(ctx context.Context, req dto.DictItemRequest) (*dto.DictItem, error) {
	dict, err := svc.dictItems.Create(ctx, model.DictItem{
		DictType: req.DictType,
		Value:    req.Value,
		Label:    req.Label,
		Enabled:  req.Enabled,
		Sort:     req.Sort,
	})
	if err != nil {
		if errors.Is(err, dao.ErrDictItemAlreadyExists) {
			return nil, ErrDictItemAlreadyExists
		}
		return nil, err
	}
	return toDictItemDTO(dict), nil
}

func (svc *AuthService) EnsureAuthenticated(ctx context.Context, authorization string) error {
	_, err := svc.authenticate(ctx, authorization)
	return err
}

func (svc *AuthService) CurrentUser(ctx context.Context, authorization string) (*model.User, error) {
	return svc.authenticate(ctx, authorization)
}

func (svc *AuthService) DeleteDictItem(ctx context.Context, id uint64) error {
	err := svc.dictItems.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrDictItemNotFound) {
			return ErrDictItemNotFound
		}
		return err
	}
	return nil
}

func (svc *AuthService) ListDictItems(ctx context.Context, dictType string) ([]dto.DictItem, error) {
	dicts, err := svc.dictItems.List(ctx, dictType)
	if err != nil {
		return nil, err
	}

	items := make([]dto.DictItem, 0, len(dicts))
	for _, dict := range dicts {
		items = append(items, *toDictItemDTO(&dict))
	}
	return items, nil
}

func (svc *AuthService) ListOperationLogs(ctx context.Context, authorization string, page int, pageSize int) (*dto.PageResult[dto.OperationLog], error) {
	if _, err := svc.parseAuthorization(authorization); err != nil {
		return nil, ErrInvalidToken
	}

	logs, total, err := svc.operationLogs.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]dto.OperationLog, 0, len(logs))
	for _, log := range logs {
		items = append(items, dto.OperationLog{
			ID:          log.ID,
			UserID:      log.UserID,
			Username:    log.Username,
			ActionValue: log.ActionValue,
			ActionLabel: log.ActionLabel,
			IP:          log.IP,
			Region:      log.Region,
			UserAgent:   log.UserAgent,
			CreatedAt:   log.CreatedAt,
		})
	}
	return &dto.PageResult[dto.OperationLog]{
		Items: items,
		Total: total,
	}, nil
}

func (svc *AuthService) UpdateDictItem(ctx context.Context, id uint64, req dto.DictItemRequest) (*dto.DictItem, error) {
	dict, err := svc.dictItems.Update(ctx, id, model.DictItem{
		DictType: req.DictType,
		Value:    req.Value,
		Label:    req.Label,
		Enabled:  req.Enabled,
		Sort:     req.Sort,
	})
	if err != nil {
		if errors.Is(err, dao.ErrDictItemNotFound) {
			return nil, ErrDictItemNotFound
		}
		if errors.Is(err, dao.ErrDictItemAlreadyExists) {
			return nil, ErrDictItemAlreadyExists
		}
		return nil, err
	}
	return toDictItemDTO(dict), nil
}

func (svc *AuthService) issueToken(user *model.User) (string, error) {
	now := time.Now()
	jti, err := randomTokenID()
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"sub":      strconv.FormatUint(user.ID, 10),
		"username": user.Username,
		"jti":      jti,
		"iat":      now.Unix(),
		"exp":      now.Add(svc.tokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(svc.jwtSecret)
}

func (svc *AuthService) authenticate(ctx context.Context, authorization string) (*model.User, error) {
	claims, err := svc.parseAuthorization(authorization)
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID, err := parseSubject(claims["sub"])
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := svc.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if !user.IsActive() {
		return nil, ErrInvalidToken
	}
	return user, nil
}

func (svc *AuthService) parseAuthorization(authorization string) (jwt.MapClaims, error) {
	tokenString := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if tokenString == "" || tokenString == authorization {
		return nil, ErrInvalidToken
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return svc.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if svc.isRevoked(claims) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (svc *AuthService) isRevoked(claims jwt.MapClaims) bool {
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return true
	}

	now := time.Now()
	svc.revokedMu.Lock()
	defer svc.revokedMu.Unlock()
	for tokenID, expiresAt := range svc.revokedTokens {
		if now.After(expiresAt) {
			delete(svc.revokedTokens, tokenID)
		}
	}
	_, ok := svc.revokedTokens[jti]
	return ok
}

func (svc *AuthService) revokeClaims(claims jwt.MapClaims) {
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return
	}

	expiresAt := time.Now().Add(svc.tokenDuration)
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	}

	svc.revokedMu.Lock()
	defer svc.revokedMu.Unlock()
	svc.revokedTokens[jti] = expiresAt
}

func parseSubject(value any) (uint64, error) {
	subject, ok := value.(string)
	if !ok {
		return 0, ErrInvalidToken
	}
	return strconv.ParseUint(subject, 10, 64)
}

func toAdminUser(user *model.User) *dto.AdminUser {
	return &dto.AdminUser{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Email:    user.Email,
	}
}

func (svc *AuthService) recordOperation(ctx context.Context, user *model.User, actionValue int, meta dto.OperationMeta) error {
	action := defaultDictItem(actionValue)
	dict, err := svc.dictItems.FindByTypeAndValue(ctx, "operation", actionValue)
	if err != nil {
		if errors.Is(err, dao.ErrDictItemNotFound) {
			dict, err = svc.dictItems.Create(ctx, defaultDictItem(actionValue))
			if err != nil && errors.Is(err, dao.ErrDictItemAlreadyExists) {
				dict, err = svc.dictItems.FindByTypeAndValue(ctx, "operation", actionValue)
			}
		}
		if err != nil {
			log.Printf("load operation dict failed: actionValue=%d err=%v", actionValue, err)
			dict = &action
		}
	}
	if !dict.Enabled {
		return nil
	}
	if !isSystemOperation(actionValue) {
		action = *dict
	}

	return svc.operationLogs.Create(ctx, model.OperationLog{
		UserID:      user.ID,
		Username:    user.Username,
		ActionValue: action.Value,
		ActionLabel: action.Label,
		IP:          meta.IP,
		Region:      meta.Region,
		UserAgent:   meta.UserAgent,
	})
}

func (svc *AuthService) safeRecordOperation(ctx context.Context, user *model.User, actionValue int, meta dto.OperationMeta) {
	if err := svc.recordOperation(ctx, user, actionValue, meta); err != nil {
		log.Printf("record operation failed: userId=%d actionValue=%d err=%v", user.ID, actionValue, err)
	}
}

func defaultDictItem(actionValue int) model.DictItem {
	switch actionValue {
	case 1:
		return model.DictItem{DictType: "operation", Value: 1, Label: "登录", Enabled: true, Sort: 1}
	case 2:
		return model.DictItem{DictType: "operation", Value: 2, Label: "退出", Enabled: true, Sort: 2}
	default:
		return model.DictItem{DictType: "operation", Value: actionValue, Label: "未知操作", Enabled: true, Sort: 100}
	}
}

func isSystemOperation(actionValue int) bool {
	return actionValue == 1 || actionValue == 2
}

func toDictItemDTO(dict *model.DictItem) *dto.DictItem {
	return &dto.DictItem{
		ID:        dict.ID,
		DictType:  dict.DictType,
		Value:     dict.Value,
		Label:     dict.Label,
		Enabled:   dict.Enabled,
		Sort:      dict.Sort,
		CreatedAt: dict.CreatedAt,
		UpdatedAt: dict.UpdatedAt,
	}
}

func ResolveIPRegion(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "未知区域"
	}
	if parsed.IsLoopback() {
		return "本机"
	}
	if parsed.IsPrivate() {
		return "内网"
	}
	return "未知区域"
}

func randomTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
