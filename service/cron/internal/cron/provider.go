package cron

import (
	"github.com/google/wire"
)

// ProviderSet 是 cron 模块的依赖注入集合
var ProviderSet = wire.NewSet(NewScheduler, NewTaskManager)
