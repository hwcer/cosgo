package zset

// future 基础 Future 实现
type future struct {
	done chan struct{}
	err  error
}

// Wait 等待异步操作完成
func (f *future) Wait() error {
	<-f.done
	return f.err
}

// futureInt64 int64 类型的 Future 实现
type futureInt64 struct {
	future
	value int64
}

// Value 获取 int64 结果
func (f *futureInt64) Value() int64 {
	f.Wait()
	return f.value
}

// futureBool bool 类型的 Future 实现
type futureBool struct {
	future
	value bool
}

// Value 获取 bool 结果
func (f *futureBool) Value() bool {
	f.Wait()
	return f.value
}

// futureRank rank 和 score 类型的 Future 实现
type futureRank struct {
	future
	rank  int64
	score int64
}

// Value 获取 rank 和 score 结果
func (f *futureRank) Value() (int64, int64) {
	f.Wait()
	return f.rank, f.score
}

// futureScore score 和 ok 类型的 Future 实现
type futureScore struct {
	future
	score int64
	ok    bool
}

// Value 获取 score 和 ok 结果
func (f *futureScore) Value() (int64, bool) {
	f.Wait()
	return f.score, f.ok
}

// futureData key 和 score 类型的 Future 实现
type futureData struct {
	future
	key   string
	score int64
}

// Value 获取 key 和 score 结果
func (f *futureData) Value() (string, int64) {
	f.Wait()
	return f.key, f.score
}
