package webhookHandler

import (
	"sync"
	"webhookTemplate/CustomizedQueue"
)

// webhookMessageIdList 用于存储已经处理过的webhook请求的id
var webhookMessageIdList = CustomizedQueue.NewQueue(31)

var webhookMessageIdListLock sync.Mutex
