package global

import "sync"

// BotState 存储每个聊天中的分页状态
type DepositState struct {
	CurrentPage int64
	TotalPages  int64
}
type CostState struct {
	CurrentPage int64
	TotalPages  int64
}

var (
	DepositStates = make(map[int64]*DepositState) // 按ChatID存储状态
	CostStates    = make(map[int64]*CostState)    // 按ChatID存储状态
)

// 定义全局变量
var (
	Translations = make(map[string]map[string]string) // 存储所有翻译

	TranslationsDir = "translations"                                                       // 翻译文件目录
	SupportedLangs  = []string{"en", "zh", "ar", "es", "pt", "ko", "th", "ja", "vi", "ch"} // 支持的语言
	DefaultLang     = "zh"                                                                 // 默认语言
	Mutex           = &sync.RWMutex{}                                                      // 读写锁
)
