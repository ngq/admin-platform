package audit

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewBiz,
	NewOperationLogRepo,
	NewLoginLogRepo,
	NewErrorLogRepo,
)
