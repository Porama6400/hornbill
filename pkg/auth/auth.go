package auth

import (
	"context"
	"fmt"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"net/http"
	"os"
	"strings"
	"time"
)
import "github.com/gin-gonic/gin"

type User struct {
	CreatedTime time.Time      `json:"createdTime"`
	ExpiryTime  time.Time      `json:"expiryTime"`
	Info        *oidc.UserInfo `json:"info"`
}

func (u *User) GetId() string {
	return u.Info.Name
}

func (u *User) IsAdmin() bool {
	//TODO perm ch
	return false
}

type ServiceConfig struct {
	ClientId                 string
	ClientSecret             string
	Issuer                   string
	CodeExchangeRedirectUri  string
	AuthenticatedRedirectUri string
	Scopes                   []string
}

type Service struct {
	ServiceConfig
	RelyingParty   rp.RelyingParty
	ResourceServer rs.ResourceServer
	SessionMap     map[string]User
}

const SessionTTL = 15 * time.Minute
const CookieSessionKey = "session"

func LoadAuthServiceConfigEnv() ServiceConfig {
	return ServiceConfig{
		ClientId:                 os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret:             os.Getenv("OIDC_CLIENT_SECRET"),
		Issuer:                   os.Getenv("OIDC_ISSUER"),
		CodeExchangeRedirectUri:  os.Getenv("OIDC_REDIRECT"),
		AuthenticatedRedirectUri: os.Getenv("AUTHENTICATED_REDIRECT"),
		Scopes:                   []string{"email", "openid", "profile"},
	}
}

func NewAuthService(ctx context.Context, config ServiceConfig) (*Service, error) {
	relyingParty, err := rp.NewRelyingPartyOIDC(ctx, config.Issuer, config.ClientId, config.ClientSecret, config.CodeExchangeRedirectUri, config.Scopes)
	if err != nil {
		return nil, err
	}

	resourceServer, err := rs.NewResourceServerClientCredentials(ctx, config.Issuer, config.ClientId, config.ClientSecret)
	if err != nil {
		return nil, err
	}

	return &Service{
		ServiceConfig:  config,
		RelyingParty:   relyingParty,
		ResourceServer: resourceServer,
		SessionMap:     make(map[string]User),
	}, nil
}

func (s *Service) codeExchangeCallback(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
	sessionId, err := GenerateSessionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieSessionKey,
		Path:     "/",
		Value:    sessionId,
		Expires:  time.Now().Add(SessionTTL),
		Secure:   false,
		HttpOnly: false,
	})

	s.SessionMap[sessionId] = User{
		CreatedTime: time.Now(),
		ExpiryTime:  time.Now().Add(SessionTTL),
		Info:        info,
	}

	http.Redirect(w, r, s.AuthenticatedRedirectUri, http.StatusFound)
}

func (s *Service) GetUser(c *gin.Context) (*User, error) {
	token, err := c.Cookie(CookieSessionKey)
	if err != nil || token == "" {
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
	}

	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	value, ok := s.SessionMap[token]
	if ok {
		return &value, nil
	} else {
		return nil, fmt.Errorf("unauthorized")
	}
}

func (s *Service) HandlePathInfo(c *gin.Context) {
	user, err := s.GetUser(c)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *Service) BindPaths(group *gin.RouterGroup) {
	group.GET("/login", gin.WrapF(rp.AuthURLHandler(GenerateState, s.RelyingParty)))
	group.GET("/callback", gin.WrapF(rp.CodeExchangeHandler(rp.UserinfoCallback(s.codeExchangeCallback), s.RelyingParty)))
	group.GET("/info", s.HandlePathInfo)
}
