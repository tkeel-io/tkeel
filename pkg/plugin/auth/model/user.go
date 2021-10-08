package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/openapi"
)

// user.tenantID
const UserPrefix = "user.%s"

var gUserStore = make(UserStoreOnTenant)

// UserName：User
type UserStoreOnTenant map[string]*User

// 用户
type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Password   string `json:"password"`
	TenantID   string `json:"tenant_id"`
	Email      string `json:"email"`
	CreateTime int64  `json:"create_time"`
}

func UserInit() {
	sysAdmin := User{
		ID:       "SysAdmin",
		Name:     "SysAdmin",
		Password: "SysAdmin",
		Email:    "sysadmin@tkeel.io",
	}
	c := keel.GetClient()

	data, _ := json.Marshal(sysAdmin)
	err := c.SaveState(context.Background(), keel.PrivateStore, sysAdmin.Name, data)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *User) Create(ctx context.Context) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	items, err := getDB().Select(ctx, genUserStateKey(r.TenantID))
	if err != nil {
		dblog.Error("[PluginAuth] User Create ", err)
		return err
	}
	if items != nil {
		json.Unmarshal(items, &gUserStore)
	}
	_, ok := gUserStore[r.Name]
	if ok {
		return errors.New(openapi.ErrResourceExisted)
	}
	r.CreateTime = time.Now().Unix()
	gUserStore[r.Name] = r
	saveData, _ := json.Marshal(gUserStore)
	user, _ := json.Marshal(r)
	getDB().Insert(ctx, r.Name, user)
	return getDB().Insert(ctx, genUserStateKey(r.TenantID), saveData)
}

// search by userName
func (r *User) List(ctx context.Context) []*User {
	if gUserStore == nil {
		return nil
	}
	users := make([]*User, 0)
	if r.Name != "" {
		user, ok := gUserStore[r.Name]
		if !ok {
			return nil
		}
		users = append(users, user)
		return users
	}
	items, err := getDB().Select(ctx, genUserStateKey(r.TenantID))
	if err != nil {
		dblog.Error("user list empty,tenantID:", r.TenantID)
		return nil
	}
	userMap := make(UserStoreOnTenant)
	if err := json.Unmarshal(items, &userMap); err != nil {
		dblog.Error("user list unmarshal err", err)
		return nil
	}
	for _, v := range userMap {
		users = append(users, v)
	}

	return users
}
func QueryUserByName(ctx context.Context, name string) *User {
	if name == "" {
		return nil
	}
	data, err := getDB().Select(ctx, name)
	if err != nil {
		dblog.Error(err)
		return nil
	}
	user := &User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		dblog.Error(err)
		return nil
	}
	return user
}

func genUserStateKey(tenantID string) string {
	return fmt.Sprintf(UserPrefix, tenantID)
}
