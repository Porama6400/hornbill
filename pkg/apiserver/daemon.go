package apiserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"hornbill/pkg/auth"
	"hornbill/pkg/pb"
	"log"
	"net/http"
	"time"
)

type DaemonUserEntry struct {
	Daemon   `json:"daemon"`
	UserList []*pb.User `json:"users,omitempty"`
}

type DaemonUserList []DaemonUserEntry

// HandleDaemonLogin godoc
// @Accept       json
// @Produce      json
// @Param		 req	  body		pb.Identity	false	"Body"
// @Param        server   path      string		true	"Server ID"
// @Success      200  {object} pb.ResultAdd
// @Failure      401  {object} auth.ResultErrorMessage
// @Failure      404  {object} auth.ResultErrorMessage
// @Router       /api/login/{server} [post]
func (s *Server) HandleDaemonLogin(c *gin.Context) {

	paramServer := c.Param("server")
	daemon := s.GetDaemon(paramServer)
	if daemon == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var identity pb.Identity
	_ = c.ShouldBindJSON(&identity)

	ctx := context.TODO()

	user, err := s.AuthService.GetUser(c)
	if err != nil || user == nil {
		// middleware already handled this check, this should not be reached
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if identity.Id != "" && user.IsAdmin() {
		log.Println("Admin user", user.GetId(), "bypassed identity lock")
	} else {
		identity.Id = user.GetId()
	}
	if identity.PublicKey == "" {
		key, err := wgtypes.GenerateKey()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, auth.NewResultErrorMessage(err))
		}

		identity.PrivateKey = key.String()
		identity.PublicKey = key.PublicKey().String()
	}

	expiryMaxUnixMillis := time.Now().Add(time.Minute * time.Duration(s.UserMaxTTLMinutes)).UnixMilli()
	if identity.Expiry == nil || *identity.Expiry > expiryMaxUnixMillis {
		if user.IsAdmin() {
			log.Println("Admin user", user.GetId(), "bypassed time limit")
		} else {
			identity.Expiry = &expiryMaxUnixMillis
		}
	}

	result, err := daemon.DaemonClient.Add(ctx, &identity)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, auth.NewResultErrorMessage(err))
		return
	}
	result.User.Identity = &identity
	c.JSON(http.StatusOK, result)
}

// HandleDaemonLogout godoc
// @Accept       json
// @Produce      json
// @Param        server   path      string		true	"Server ID"
// @Success      200  {object} pb.Result
// @Failure      401  {object} auth.ResultErrorMessage
// @Failure      404  {object} auth.ResultErrorMessage
// @Router       /api/logout/{server} [post]
func (s *Server) HandleDaemonLogout(c *gin.Context) {

	paramServer := c.Param("server")
	daemon := s.GetDaemon(paramServer)
	if daemon == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	ctx := context.TODO()

	var identity pb.Identity
	_ = c.ShouldBindJSON(&identity)

	user, err := s.AuthService.GetUser(c)
	if err != nil || user == nil {
		// middleware already handled this check, this should not be reached
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if identity.Id != "" && user.IsAdmin() {
		log.Println("Admin user", user.GetId(), "bypassed identity lock")
	} else {
		identity.Id = user.GetId()
	}

	result, err := daemon.DaemonClient.Remove(ctx, &identity)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, auth.NewResultErrorMessage(err))
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleDaemonList godoc
// @Accept       json
// @Produce      json
// @Success      200  {object} DaemonUserList
// @Failure      401  {object} auth.ResultErrorMessage
// @Failure      404  {object} auth.ResultErrorMessage
// @Router       /api/list [get]
func (s *Server) HandleDaemonList(c *gin.Context) {
	ctx := context.TODO()
	result := make(DaemonUserList, 0)
	for _, daemon := range s.DaemonList {
		list, err := daemon.DaemonClient.List(ctx, &pb.Empty{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, auth.NewResultErrorMessage(err))
			return
		}

		result = append(result, DaemonUserEntry{
			Daemon:   daemon,
			UserList: list.Users,
		})
	}
	c.JSON(http.StatusOK, result)
}
