package webhookHandler

import "webhookTemplate/CustomizedQueue"

// webhookMessageIdList 用于存储已经处理过的webhook请求的id
var webhookMessageIdList = CustomizedQueue.NewQueue(15)
