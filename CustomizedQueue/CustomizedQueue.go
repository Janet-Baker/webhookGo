package CustomizedQueue

// CustomizedQueue 自定义队列数据结构
type CustomizedQueue struct {
	data    []string
	maxSize int
	front   int
	rear    int
}

// NewQueue 初始化队列
func NewQueue(size int) *CustomizedQueue {
	var result = &CustomizedQueue{}

	result.maxSize = size
	result.data = make([]string, size)
	result.front = 0
	result.rear = 0

	return result
}

/*// IsEmpty 判断是否是空队列
func (s *CustomizedQueue) IsEmpty() bool {
	return s.data != nil && s.rear == s.front
}*/

/*// GetQueueLength 获取队列元素个数
func (s *CustomizedQueue) GetQueueLength() int {
	return (s.rear - s.front + s.maxSize) % s.maxSize
}*/

/*// DeQueue 出队
func (s *CustomizedQueue) DeQueue() (string, error) {
	if s.IsEmpty() {
		return "EmptyQueueError", errors.New("队列为空")
	}
	result := s.data[s.front]
	s.front = (s.front + 1) % s.maxSize
	return result, nil
}*/

// EnQueue 入队
func (s *CustomizedQueue) EnQueue(item string) {
	s.data[s.rear] = item
	s.rear = (s.rear + 1) % s.maxSize
	if s.front == s.rear {
		s.front = (s.front + 1) % s.maxSize
	}
}

// IsContain 判断队列中是否包含某个元素
func (s *CustomizedQueue) IsContain(item string) bool {
	for i := 0; i < s.maxSize; i++ {
		if s.data[i] == item {
			return true
		}
	}
	/*	for _, eachItem := range s.data {
		if eachItem == item {
			return true
		}
	}*/
	return false
}
