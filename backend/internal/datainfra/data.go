// 数据层基础设施
// Data 封装所有持久化依赖，提供统一的数据库与 Casbin 访问入口
// 独立包，避免 data 子包引用 data 父包导致循环依赖
package datainfra

import (
	"context"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/ngq/admin-platform/internal/data/po"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"gorm.io/gorm"
)

// defaultCasbinRBACModel 内置默认 RBAC 模型定义，当外部配置文件不存在时作为高可用兜底。
const defaultCasbinRBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// gormDBWrapper 适配 *gorm.DB 到框架 datacontract.GormDB 契约
// gorm.DB.WithContext 返回 *gorm.DB，而契约要求返回 any，需桥接
type gormDBWrapper struct {
	db *gorm.DB
}

// WithContext 实现 datacontract.GormDB 接口
// 将 *gorm.DB.WithContext 的返回值转为 any，满足契约签名
func (w *gormDBWrapper) WithContext(ctx context.Context) any {
	return w.db.WithContext(ctx)
}

// Data 数据层基础设施，封装所有持久化依赖
// db 字段使用框架 GormDB 契约（强制 WithContext），而非裸 *gorm.DB
// DB(ctx) 方法通过契约的 WithContext(ctx) 返回绑定 context 的 *gorm.DB
type Data struct {
	db     datacontract.GormDB    // 框架 GormDB 契约（私有字段，通过 DB(ctx) 访问）
	rawDB  *gorm.DB               // 原始数据库连接（用于 Biz 层特殊查询）
	Casbin *casbin.SyncedEnforcer // Casbin 同步权限执行器（并发安全，适用于 HTTP 服务）
}

// NewData 创建数据层实例
// 接收框架提供的 *gorm.DB，包装为框架 GormDB 契约
// 初始化 GORM 适配器并创建 Casbin 同步执行器（支持指定 modelPath 或内置字符串兜底）
func NewData(db *gorm.DB, modelPath ...string) *Data {
	// 使用 NewAdapterByDBWithCustomTable 传入自定义 CasbinRulePO 结构体
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &po.CasbinRulePO{})
	if err != nil {
		panic("failed to create casbin adapter: " + err.Error())
	}

	// 优先检查外部配置文件，若文件不存在则无缝降级为内置模型字符串（避免相对路径依赖导致的 panic）
	var enforcer *casbin.SyncedEnforcer
	customPath := "config/casbin_model.conf"
	if len(modelPath) > 0 && modelPath[0] != "" {
		customPath = modelPath[0]
	}

	if _, statErr := os.Stat(customPath); statErr == nil {
		enforcer, err = casbin.NewSyncedEnforcer(customPath, adapter)
	} else {
		m, modelErr := model.NewModelFromString(defaultCasbinRBACModel)
		if modelErr != nil {
			panic("failed to parse default casbin model: " + modelErr.Error())
		}
		enforcer, err = casbin.NewSyncedEnforcer(m, adapter)
	}
	if err != nil {
		panic("failed to create casbin synced enforcer: " + err.Error())
	}

	// 加载策略到内存
	_ = enforcer.LoadPolicy()

	return &Data{
		db:     &gormDBWrapper{db: db},
		rawDB:  db,
		Casbin: enforcer,
	}
}

// DB 返回绑定 context 的数据库连接
// 强制要求传入 context，确保 trace/timeout/cancel 能力生效
// 通过框架 GormDB 契约的 WithContext(ctx) 方法获取带 context 的 *gorm.DB
func (d *Data) DB(ctx context.Context) *gorm.DB {
	// WithContext(ctx) 返回 any（框架契约定义），实际类型为 *gorm.DB
	// 类型断言安全：gorm.DB.WithContext 始终返回 *gorm.DB
	return d.db.WithContext(ctx).(*gorm.DB)
}

// RawDB 返回原始数据库连接（不绑定 context）
// 仅用于 Biz 层需要存储原始 DB 连接的场景
func (d *Data) RawDB() *gorm.DB {
	return d.rawDB
}
