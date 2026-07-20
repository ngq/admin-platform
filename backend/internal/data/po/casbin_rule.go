// Domain: IAM
// Casbin 策略规则持久化对象
// CasbinRulePO 对应 casbin_rule 表，存储 Casbin RBAC 策略
// 此表由 gorm-adapter 通过 NewAdapterByDBWithCustomTable 管理建表
// 列宽使用 varchar(100)，与 gorm-adapter 默认的 size:100 一致
// 注意：列名必须用 gorm:"column:xxx" 显式指定，因为 GORM 默认将 PType 转为 p_type
package po

// CasbinRulePO Casbin策略规则持久化对象
type CasbinRulePO struct {
	ID    int64  `gorm:"primaryKey;autoIncrement;column:id"`
	PType string `gorm:"type:varchar(12);column:ptype"`
	V0    string `gorm:"type:varchar(100);column:v0"`
	V1    string `gorm:"type:varchar(100);column:v1"`
	V2    string `gorm:"type:varchar(100);column:v2"`
	V3    string `gorm:"type:varchar(100);column:v3"`
	V4    string `gorm:"type:varchar(100);column:v4"`
	V5    string `gorm:"type:varchar(100);column:v5"`
}

// TableName 指定表名
func (CasbinRulePO) TableName() string { return "casbin_rule" }
