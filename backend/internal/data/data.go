// 数据层根包
// 重新导出 datainfra.Data，保持对外的兼容性
package data

import (
	"gorm.io/gorm"

	"github.com/ngq/admin-platform/internal/datainfra"
)

// Data 重新导出 datainfra.Data，保持兼容
// 推荐直接使用 datainfra.Data（子包应引用 datainfra 而非 data，避免循环依赖）
type Data = datainfra.Data

// NewData 重新导出 datainfra.NewData
func NewData(db *gorm.DB) *Data {
	return datainfra.NewData(db)
}
