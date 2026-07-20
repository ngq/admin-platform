// 数据层基础设施
// Data 封装所有持久化依赖，提供统一的数据库与 Casbin 访问入口
// 独立包，避免 data 子包引用 data 父包导致循环依赖
package datainfra

import (
	"context"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/ngq/admin-platform/internal/data/po"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"gorm.io/gorm"
)

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
// 初始化 GORM 适配器并创建 Casbin 同步执行器
func NewData(db *gorm.DB) *Data {
	// 使用 NewAdapterByDBWithCustomTable 传入自定义 CasbinRulePO 结构体
	// 根因：gorm-adapter 默认的 CasbinRule 使用 gorm:"size:100"（varchar(100)），
	// 但如果 migrate 中 AutoMigrate 了列宽更大的 PO（如 varchar(128)），
	// GORM 会检测到列宽差异并执行 ALTER TABLE MODIFY COLUMN，
	// 在 MySQL utf8mb4 字符集下 varchar(128) * 4 = 512 > 3072/7（组合索引限制），导致索引长度报错。
	// 解决方案：统一由 gorm-adapter 通过自定义结构体管理 casbin_rule 表，
	// 不在 migrate 中 AutoMigrate 此表，避免列宽冲突。
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &po.CasbinRulePO{})
	if err != nil {
		panic("failed to create casbin adapter: " + err.Error())
	}

	// 创建 Casbin 同步执行器，使用 RBAC 模型配置文件
	// SyncedEnforcer 内部加锁，适合并发 HTTP 场景
	enforcer, err := casbin.NewSyncedEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		panic("failed to create casbin synced enforcer: " + err.Error())
	}

	// 加载策略到内存
	_ = enforcer.LoadPolicy()

	return &Data{
		// 将 *gorm.DB 包装为 datacontract.GormDB 接口
		// 确保所有 DB 访问都经过 WithContext，强制 context 传递
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
