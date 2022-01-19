package service

import (
	"context"

	pb "github.com/tkeel-io/tkeel/api/tenant/v1"

	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/security/authz/rbac"
	"github.com/tkeel-io/security/model"
	"github.com/tkeel-io/security/utils"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type TenantService struct {
	pb.UnimplementedTenantServer
	DB             *gorm.DB
	TenantPluginOp rbac.TenantPluginMgr
}

func NewTenantService(db *gorm.DB, tenantPluginOp rbac.TenantPluginMgr) *TenantService {
	return &TenantService{DB: db, TenantPluginOp: tenantPluginOp}
}

func (s *TenantService) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantResponse, error) {
	var (
		err    error
		tenant = &model.Tenant{}
		resp   = &pb.CreateTenantResponse{}
	)
	tenant.ID = req.Body.GetTenantId()
	if tenant.ID == "" {
		tenant.ID, _ = utils.RandBase64String(6)
	}
	tenant.Title = req.Body.Title
	tenant.Remark = req.Body.Remark
	if tenant.Existed(s.DB) {
		return nil, pb.ErrTenantAlreadyExisted()
	}
	err = tenant.Create(s.DB)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	resp.TenantId = tenant.ID
	resp.TenantTitle = tenant.Title
	if req.Body.Admin != nil {
		user := model.User{TenantID: tenant.ID, UserName: req.Body.Admin.Username, Password: req.Body.Admin.Password}
		err = user.Create(s.DB)
		if err != nil {
			log.Error(err)
			return resp, pb.ErrUnknown()
		}
		resp.AdminUsername = user.UserName
	}
	s.TenantPluginOp.OnCreateTenant(tenant.ID)
	return resp, nil
}

func (s *TenantService) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantResponse, error) {
	var (
		err     error
		tenants []*model.Tenant
		tenant  = &model.Tenant{}
		resp    = &pb.GetTenantResponse{}
	)
	if req.GetTenantId() == "" {
		return nil, pb.ErrInvalidArgument()
	}
	tenant.ID = req.GetTenantId()
	tenants, err = tenant.List(s.DB, nil)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrListTenant()
	}
	if len(tenants) == 1 {
		resp.Title = tenants[0].Title
		resp.TenantId = tenants[0].ID
		resp.Remark = tenants[0].Remark
	}
	return resp, nil
}

func (s *TenantService) ListTenant(ctx context.Context, _ *emptypb.Empty) (*pb.ListTenantResponse, error) {
	var (
		err     error
		tenant  = &model.Tenant{}
		tenants = []*model.Tenant{}
		resp    = &pb.ListTenantResponse{}
	)
	tenants, err = tenant.List(s.DB, nil)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrListTenant()
	}

	resp.Tenants = make([]*pb.TenantDetail, len(tenants))
	for _, v := range tenants {
		userDao := &model.User{}
		detail := &pb.TenantDetail{TenantId: v.ID, Title: v.Title, Remark: v.Remark}
		num_user, err := userDao.CountInTenant(s.DB, v.ID)
		if err != nil {
			log.Error(err)
			return nil, pb.ErrListTenant()
		}

		detail.NumUser = num_user
		resp.Tenants = append(resp.Tenants, detail)
	}

	return resp, nil
}

func (s *TenantService) DeleteTenant(ctx context.Context, req *pb.DeleteTenantRequest) (*emptypb.Empty, error) {
	var (
		err    error
		tenant = &model.Tenant{}
		resp   = &emptypb.Empty{}
	)
	tenant.ID = req.TenantId
	if !tenant.Existed(s.DB) {
		return nil, pb.ErrInvalidArgument()
	}

	err = tenant.Delete(s.DB)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}

	user := &model.User{}
	err = user.DeleteAllInTenant(s.DB, req.TenantId)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	return resp, nil
}

func (s *TenantService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	var (
		err     error
		existed bool
		resp    *pb.CreateUserResponse
		user    = &model.User{}
	)

	user.TenantID = req.GetTenantId()
	user.UserName = req.GetBody().GetUsername()
	user.Password = req.GetBody().GetPassword()
	existed, err = user.Existed(s.DB)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	if existed {
		return nil, pb.ErrAlreadyExistedUser()
	}

	err = user.Create(s.DB)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	resp = &pb.CreateUserResponse{TenantId: user.TenantID, Username: user.UserName, UserId: user.ID}
	return resp, nil
}

func (s *TenantService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var (
		err       error
		resp      *pb.GetUserResponse
		user      = &model.User{}
		condition = make(map[string]interface{})
	)
	condition["id"] = req.GetUserId()
	condition["tenant_id"] = req.GetTenantId()
	_, users, err := user.QueryByCondition(s.DB, condition, nil)
	if err != nil || len(users) != 1 {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	resp = &pb.GetUserResponse{
		TenantId:   users[0].TenantID,
		UserId:     users[0].ID,
		Username:   users[0].UserName,
		Email:      users[0].Email,
		ExternalId: users[0].ExternalID,
		Avatar:     users[0].Avatar,
		NickName:   users[0].NickName,
	}
	return resp, nil
}

func (s *TenantService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	var (
		err       error
		resp      *pb.ListUserResponse
		user      = &model.User{}
		condition = make(map[string]interface{})
		page      = &model.Page{}
	)
	condition["tenant_id"] = req.GetTenantId()
	page.PageNum = int(req.PageNum)
	page.PageSize = int(req.PageSize)
	page.OrderBy = req.OrderBy
	page.IsDescending = req.IsDescending
	total, users, err := user.QueryByCondition(s.DB, condition, page)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	userList := make([]*pb.UserListData, len(users))
	for i, v := range users {
		detail := &pb.UserListData{TenantId: v.TenantID, UserId: v.ID, Username: v.UserName,
			Email: v.Email, ExternalId: v.ExternalID, Avatar: v.Avatar, NickName: v.NickName}
		userList[i] = detail
	}
	resp = &pb.ListUserResponse{Total: total, PageSize: int32(page.PageSize), PageNum: int32(page.PageNum), Users: userList}
	return resp, nil
}

func (s *TenantService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	var (
		err     error
		existed bool
		user    = &model.User{}
	)
	user.ID = req.GetUserId()
	user.TenantID = req.GetTenantId()
	existed, err = user.Existed(s.DB)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	if !existed {
		return nil, pb.ErrResourceNotFound()
	}
	err = user.Delete(s.DB)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalStore()
	}
	return &emptypb.Empty{}, nil
}

func (s *TenantService) AddTenantPlugin(ctx context.Context, req *pb.AddTenantPluginRequest) (*pb.AddTenantPluginResponse, error) {
	ok, err := s.TenantPluginOp.AddTenantPlugin(req.GetTenantId(), req.GetBody().GetPluginId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.AddTenantPluginResponse{Ok: ok}, nil
}

func (s *TenantService) ListTenantPlugin(ctx context.Context, req *pb.ListTenantPluginRequest) (*pb.ListTenantPluginResponse, error) {
	plugins := s.TenantPluginOp.ListTenantPlugins(req.GetTenantId())
	return &pb.ListTenantPluginResponse{Plugins: plugins}, nil
}

func (s *TenantService) DeleteTenantPlugin(ctx context.Context, req *pb.DeleteTenantPluginRequest) (*pb.DeleteTenantPluginResponse, error) {
	ok, err := s.TenantPluginOp.DeleteTenantPlugin(req.GetTenantId(), req.GetPluginId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.DeleteTenantPluginResponse{Ok: ok}, nil
}

func (s *TenantService) TenantPluginPermissible(ctx context.Context, req *pb.PluginPermissibleRequest) (*pb.PluginPermissibleResponse, error) {
	allowed, err := s.TenantPluginOp.TenantPluginPermissible(req.GetTenantId(), req.GetPluginId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.PluginPermissibleResponse{Allowed: allowed}, nil
}
