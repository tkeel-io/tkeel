package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/tkeel-io/kit/log"
	transport_http "github.com/tkeel-io/kit/transport/http"
	"github.com/tkeel-io/security/authz/rbac"
	s_model "github.com/tkeel-io/security/model"
	pb "github.com/tkeel-io/tkeel/api/authentication/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	keel_v1 "github.com/tkeel-io/tkeel/pkg/service/keel/v1"
	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/util"
	"github.com/tkeel-io/tkeel/pkg/version"
	"gorm.io/gorm"
)

var (
	errNotFoundAddons    = errors.New("not found addons")
	errNotActiveUpstream = errors.New("not found active upstream")
)

type session struct {
	Src           *endpoint
	Dst           *endpoint
	User          *model.User
	RequestMethod string
}

func (s *session) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type endpoint struct {
	ID           string
	TKeelDepened string
}

func (e *endpoint) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type AuthenticationService struct {
	pb.UnimplementedAuthenticationServer

	conf           *TokenConf
	m              *manage.Manager
	userDB         *gorm.DB
	rbacOp         *casbin.SyncedEnforcer
	prOp           proute.Operator
	secretProvider token.Provider
	tenantPluginOp rbac.TenantPluginMgr

	regExpWhiteList []string
}

func NewAuthenticationService(m *manage.Manager, userDB *gorm.DB, conf *TokenConf,
	rbacOp *casbin.SyncedEnforcer, prOp proute.Operator, tpOp rbac.TenantPluginMgr) *AuthenticationService {
	secret := util.RandStringBytesMaskImpr(0, 10)
	tokenConf := manage.DefaultAuthorizeCodeTokenCfg
	if conf.AccessTokenExp != 0 && conf.RefreshTokenExp != 0 {
		tokenConf.AccessTokenExp = conf.AccessTokenExp
		tokenConf.RefreshTokenExp = conf.RefreshTokenExp
	}
	return &AuthenticationService{
		secretProvider: token.InitProvider(secret, "", ""),
		userDB:         userDB,
		conf:           conf,
		m:              m,
		rbacOp:         rbacOp,
		prOp:           prOp,
		tenantPluginOp: tpOp,
		regExpWhiteList: []string{
			"/apis/rudder/v1/oauth2*",
			"/apis/security/v1/oauth*",
		},
	}
}

func (s *AuthenticationService) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	sess := new(session)
	if err := s.setDstAndMethod(ctx, sess, req.Path); err != nil {
		if errors.Is(err, pb.ErrUpstreamNotFound()) {
			return nil, err
		}
		log.Errorf("error set dst and method session: %s", err)
		return nil, pb.ErrInternalError()
	}
	// white path.
	if !s.matchRegExpWhiteList(req.Path) {
		// set src plugin id.
		pluginID, err := s.checkXPluginJwt(ctx, header)
		if err != nil {
			log.Errorf("error get plugin ID from request: %s", err)
			return nil, pb.ErrInvalidXPluginJwtToken()
		}
		sess.Src = &endpoint{
			ID: pluginID,
		}
		// set user.
		if pluginID == "keel" {
			log.Debugf("external flow(%s)", req)
			sess.User, err = s.checkAuthorization(ctx, header)
			if err != nil {
				log.Errorf("error external get user: %s", err)
				return nil, pb.ErrUnauthenticated()
			}
			sess.Src.TKeelDepened = version.Version
			return convertSession2PB(sess), nil
		}
		log.Debugf("internal flow(%s)", req)
		sess.User, err = checkXtKeelAtuh(ctx, header)
		if err != nil {
			log.Errorf("error internal get user: %s", err)
			return nil, pb.ErrInvalidXTkeelAuthToken()
		}
		// set src tkeel depened.
		if pluginIsTkeelComponent(pluginID) {
			pr, err1 := s.prOp.Get(ctx, pluginID)
			if err1 != nil {
				if errors.Is(err1, proute.ErrPluginRouteNotExsist) {
					return nil, pb.ErrUpstreamNotFound()
				}
				log.Errorf("error get plugin(%s) route", pluginID)
				return nil, pb.ErrInternalError()
			}
			sess.Src.TKeelDepened = pr.TkeelVersion
		} else {
			sess.Src.TKeelDepened = version.Version
		}
		if err = s.checkSession(ctx, sess); err != nil {
			log.Errorf("error check session(%s): %s", sess, err)
			if errors.Is(err, pb.ErrInvalidArgument()) || errors.Is(err, pb.ErrUpstreamNotEnable()) {
				return nil, err
			}
			return nil, pb.ErrInternalError()
		}
	}
	log.Debugf("session: %s", sess)
	return convertSession2PB(sess), nil
}

func convertSession2PB(s *session) *pb.AuthenticateResponse {
	ret := &pb.AuthenticateResponse{}
	if s.User != nil {
		ret.UserId = s.User.User
		ret.TenantId = s.User.Tenant
		ret.Role = s.User.Role
	}
	if s.Dst != nil {
		ret.Destination = s.Dst.ID
	}
	ret.Method = s.RequestMethod
	return ret
}

func checkXtKeelAtuh(ctx context.Context, header http.Header) (*model.User, error) {
	xtKeelAuthString := header.Get(model.XtKeelAuthHeader)
	if xtKeelAuthString == "" {
		return nil, errors.New("invalid x-tKeel-auth")
	}
	user := new(model.User)
	if err := user.Base64Decode(xtKeelAuthString); err != nil {
		return nil, fmt.Errorf("error decode x-tKeel-auth(%s): %w", xtKeelAuthString, err)
	}
	return user, nil
}

func pluginIsTkeelComponent(pluginID string) bool {
	for _, v := range model.TKeelComponents {
		if v == pluginID {
			return true
		}
	}
	return false
}

func (s *AuthenticationService) setDstAndMethod(ctx context.Context, sess *session, path string) error {
	sess.Dst = &endpoint{}
	if isAddons(path) {
		if err := s.getAddonsDstAndMethod(ctx, sess, path); err != nil {
			if errors.Is(err, errNotFoundAddons) || errors.Is(err, errNotActiveUpstream) {
				return pb.ErrUpstreamNotFound()
			}
			return fmt.Errorf("error get(%s) addons destination and method: %w", path, err)
		}
	} else {
		if err := s.getPluginDstAndMethod(ctx, sess, path); err != nil {
			if errors.Is(err, errNotActiveUpstream) {
				return errNotActiveUpstream
			}
			return fmt.Errorf("error get(%s) plugin destination and method: %w", path, err)
		}
	}
	return nil
}

func (s *AuthenticationService) getPluginDstAndMethod(ctx context.Context, sess *session, path string) error {
	if err := checkPath(path); err != nil {
		return fmt.Errorf("error checkpath(%s): %w", path, err)
	}
	pluginID := getPluginIDFromPath(path)
	pluginMethod := getMethodApisPath(path)
	if pluginID == "" {
		return fmt.Errorf("get(%s) invalid plugin id", path)
	}
	if pluginIsTkeelComponent(pluginID) {
		// convert security ==> rudder.
		if pluginID == "security" {
			pluginID = "rudder"
		}
		sess.Dst.TKeelDepened = version.Version
	} else {
		upstreamRoute, err := s.prOp.Get(ctx, pluginID)
		if err != nil {
			if errors.Is(err, proute.ErrPluginRouteNotExsist) {
				return errNotActiveUpstream
			}
			return fmt.Errorf("errors get plugin(%s) route: %w", sess.Src.ID, err)
		}
		sess.Dst.TKeelDepened = upstreamRoute.TkeelVersion
	}
	sess.Dst.ID = pluginID
	sess.RequestMethod = pluginMethod
	return nil
}

func (s *AuthenticationService) getAddonsDstAndMethod(ctx context.Context, sess *session, path string) error {
	addonsMethod := getMethodApisPath(path)
	if addonsMethod == "" {
		return errors.New("invalid addons method")
	}
	srcRoute, err := s.prOp.Get(ctx, sess.Src.ID)
	if err != nil {
		return fmt.Errorf("errors get plugin(%s) route: %w", sess.Src.ID, err)
	}
	dstStr, ok := srcRoute.RegisterAddons[addonsMethod]
	if !ok {
		return errNotFoundAddons
	}
	dstID, dstMethod := util.DecodePluginRoute(dstStr)
	dstRoute, err := s.prOp.Get(ctx, dstID)
	if err != nil {
		if errors.Is(err, proute.ErrPluginRouteNotExsist) {
			return errNotFoundAddons
		}
		return fmt.Errorf("errors get plugin(%s) route: %w", dstID, err)
	}
	sess.Dst.ID = dstID
	sess.RequestMethod = dstMethod
	sess.Dst.TKeelDepened = dstRoute.TkeelVersion
	return nil
}

func (s *AuthenticationService) checkSession(ctx context.Context, sess *session) error {
	ok, err := util.CheckRegisterPluginTkeelVersion(sess.Dst.TKeelDepened, sess.Src.TKeelDepened)
	if err != nil {
		return fmt.Errorf("error check upstream tkeel depened(%s/%s): %w", sess.Dst.TKeelDepened, sess.Src.TKeelDepened, err)
	}
	if !ok {
		return pb.ErrInvalidArgument()
	}
	ok, err = s.tenantPluginOp.TenantPluginPermissible(sess.User.Tenant, sess.Dst.ID)
	if err != nil {
		return fmt.Errorf("error check tenant permissible(%s/%s): %w", sess.User.Tenant, sess.Dst.ID, err)
	}
	if !ok {
		return pb.ErrUpstreamNotEnable()
	}
	// TODO: check role permissible.
	return nil
}

func (s *AuthenticationService) checkXPluginJwt(ctx context.Context, header http.Header) (string, error) {
	pluginToken := header.Get(model.XPluginJwtHeader)
	if pluginToken == "" {
		return "", nil
	}
	payload, ok, err := s.secretProvider.Parse(strings.TrimPrefix(pluginToken, "Bearer "))
	if err != nil {
		return "", fmt.Errorf("error parse plugin token(%s): %w", pluginToken, err)
	}
	if !ok {
		return "", fmt.Errorf("plugin invalid token(%s)", pluginToken)
	}
	pluginIDInterface, ok := payload["plugin_id"]
	if !ok {
		return "", fmt.Errorf("error plugin token(%s) not found plugin_id", pluginToken)
	}
	pluginID, ok := pluginIDInterface.(string)
	if !ok {
		return "", fmt.Errorf("error plugin token(%s) payload plugin_id(%v) type invalid",
			pluginToken, pluginIDInterface)
	}
	return pluginID, nil
}

func (s *AuthenticationService) checkAuthorization(ctx context.Context, header http.Header) (*model.User, error) {
	authorization := header.Get(model.AuthorizationHeader)
	if authorization == "" {
		return nil, errors.New("invalid authorization")
	}
	isManager, err := s.isManagerToken(authorization)
	if err != nil {
		return nil, fmt.Errorf("error manager authorization(%s) check: %w", authorization, err)
	}
	user := new(model.User)
	tokenStr := strings.TrimPrefix(authorization, "Bearer ")
	if !isManager {
		token, err := s.m.LoadAccessToken(ctx, tokenStr)
		if err != nil {
			return nil, fmt.Errorf("error load access token(%s): %w", token, err)
		}
		userID := token.GetUserID()
		u := &s_model.User{}
		_, users, err := u.QueryByCondition(s.userDB, map[string]interface{}{"id": userID}, nil)
		if err != nil || len(users) != 1 {
			return nil, fmt.Errorf("error query user(%s)", userID)
		}
		roles := s.rbacOp.GetRolesForUserInDomain(u.TenantID, u.ID)
		if len(roles) > 0 {
			// only one role bind user.
			user.Role = roles[0]
		}
		user.User = u.ID
		user.Tenant = u.TenantID
	} else {
		_, valid, err := s.secretProvider.Parse(tokenStr)
		if err != nil {
			return nil, fmt.Errorf("error parse token(%s): %w", authorization, err)
		}
		if !valid {
			return nil, errors.New("invalid authorization")
		}
		user.User = model.TKeelUser
		user.Tenant = model.TKeelTenant
		user.Role = model.AdminRole
	}
	return user, nil
}

func (s *AuthenticationService) isManagerToken(token string) (bool, error) {
	payload, err := s.secretProvider.ParseUnverified(strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		return false, fmt.Errorf("error parse token(%s): %w", token, err)
	}
	iss, ok := payload["iss"]
	if !ok {
		return false, nil
	}
	if iss == "rudder" {
		return true, nil
	}
	return false, nil
}

func (s *AuthenticationService) matchRegExpWhiteList(path string) bool {
	for _, v := range s.regExpWhiteList {
		match, err := regexp.MatchString(v, path)
		if err != nil {
			log.Errorf("error regular expression(%s/%s) run: %s", v, path, err)
			return false
		}
		if match {
			return true
		}
	}
	return false
}

func isAddons(path string) bool {
	return strings.HasPrefix(path, keel_v1.ApisRootPath+keel_v1.AddonsRootPath)
}

func getPluginIDFromPath(pluginPath string) string {
	ss := strings.SplitN(pluginPath, "/", 4)
	if len(ss) != 4 {
		return ""
	}
	return ss[2]
}

func getMethodApisPath(apisPath string) string {
	ss := strings.SplitN(apisPath, "/", 4)
	if len(ss) != 4 {
		return ""
	}
	return strings.Split(ss[3], "?")[0]
}

func checkPath(path string) error {
	ok := false
	for _, v := range []string{"/apis", "/ws", "/static"} {
		if strings.HasPrefix(path, v) {
			ok = true
			break
		}
	}
	if !ok {
		return errors.New("invalid path: " + path)
	}
	return nil
}
