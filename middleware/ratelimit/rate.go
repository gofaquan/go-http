package ratelimit

import (
	"errors"
	"sync"
	"time"
)

const infinityDuration time.Duration = 0x7fffffffffffffff

var (
	fillIntervalErr = errors.New("token bucket fill interval is not > 0")
	capacityErr     = errors.New("token bucket capacity is not > 0")
	quantumErr      = errors.New("token bucket quantum is not > 0")
)

// tokenBucket 令牌桶
type tokenBucket struct {
	// startTime 保存存储桶首次创建并开始滴答的时刻。
	startTime time.Time

	// 容量
	capacity int64

	// 每次加多少
	quantum int64

	// 填充的间隔
	fillInterval time.Duration

	// mu guards the fields below it.
	// 保证下面两个字段的并发安全
	mu sync.Mutex

	// 可以拿到的 token 数
	availableTokens int64

	// 上次填充的时间戳
	latestTick int64
}
type tbOption func(bucket *tokenBucket)

func WithQuantum(quantum int64) tbOption {
	return func(bucket *tokenBucket) {
		bucket.quantum = quantum
	}
}

func NewBucket(fillInterval time.Duration, capacity int64, opts ...tbOption) (*tokenBucket, error) {
	tb := &tokenBucket{
		startTime:       time.Now(),
		capacity:        capacity,
		quantum:         1,
		fillInterval:    fillInterval,
		mu:              sync.Mutex{},
		availableTokens: capacity,
		latestTick:      0,
	}

	for _, opt := range opts {
		opt(tb)
	}

	if err := tb.check(); err != nil {
		return nil, err
	}

	return tb, nil
}

// 获取现在的间隔数
func (tb *tokenBucket) check() error {
	if tb.fillInterval <= 0 {
		return fillIntervalErr
	}
	if tb.capacity <= 0 {
		return capacityErr
	}
	if tb.quantum <= 0 {
		return quantumErr
	}
	return nil
}

// 获取现在的间隔数
func (tb *tokenBucket) currentTick(now time.Time) int64 {
	return int64(now.Sub(tb.startTime) / tb.fillInterval)
}

// 根据过去的间隔来更新令牌数
func (tb *tokenBucket) adjustAvailableTokens(tick int64) {
	latestTick := tb.latestTick
	tb.latestTick = tick
	if tb.availableTokens >= tb.capacity {
		return
	}

	tb.availableTokens += (tick - latestTick) * tb.quantum
	if tb.availableTokens > tb.capacity {
		tb.availableTokens = tb.capacity
	}

	return
}

// WaitMaxDuration 等待一段时间拿令牌，拿不到就返回 false
func (tb *tokenBucket) WaitMaxDuration(count int64, maxWait time.Duration) bool {
	d, ok := tb.TakeMaxDuration(count, maxWait)
	// 根据 take 的代码推出，其实只有在 这段时间内可以拿到对应数目的令牌才会 sleep
	if d > 0 {
		time.Sleep(d)
		//tb.clock.Sleep(d)
	}
	return ok
}

// TakeMaxDuration 返回 能拿到的话需要的时间 和 是否能拿到
func (tb *tokenBucket) TakeMaxDuration(count int64, maxWait time.Duration) (time.Duration, bool) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.take(time.Now(), count, maxWait)
	//return tb.take(tb.clock.Now(), count, maxWait)
	////return tb.take(tb.clock.Now(), count, maxWait)
}
func (tb *tokenBucket) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}
	// 拿到现在的间隔数
	tick := tb.currentTick(now)
	tb.adjustAvailableTokens(tick)

	avail := tb.availableTokens - count
	// 获取成功，直接返回
	if avail >= 0 {
		tb.availableTokens = avail
		return 0, true
	}

	// avail < 0
	// 这里开始算我要等多少个间隔才能拿到
	endTick := tick + (-avail+tb.quantum-1)/tb.quantum
	endTime := tb.startTime.Add(time.Duration(endTick) * tb.fillInterval)
	waitTime := endTime.Sub(now)
	// 等了还是拿不到
	if waitTime > maxWait {
		return 0, false
	}
	// 等了时间内可以拿到
	tb.availableTokens = avail
	return waitTime, true
}

// Wait 一直等待到可用
func (tb *tokenBucket) Wait(count int64) {
	if d := tb.Take(count); d > 0 {
		time.Sleep(d)
		//tb.clock.Sleep(d)
	}
}

// Take 返回拿 count 个需要的时间
func (tb *tokenBucket) Take(count int64) time.Duration {
	tb.mu.Lock()
	tb.mu.Unlock()
	d, _ := tb.take(time.Now(), count, infinityDuration)
	//d, _ := tb.take(tb.clock.Now(), count, infinityDuration)
	return d
}

// 有多少拿多少
func (tb *tokenBucket) takeAvailable(now time.Time, count int64) int64 {
	if count <= 0 {
		return 0
	}
	tb.adjustAvailableTokens(tb.currentTick(now))
	if tb.availableTokens <= 0 {
		return 0 // 拿不到就返回
	}
	// 如果要的令牌不够多
	if count > tb.availableTokens {
		count = tb.availableTokens // 拿到尽可能多的令牌
	}

	tb.availableTokens -= count
	return count
}

func (tb *tokenBucket) TakeAvailable(count int64) int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.takeAvailable(time.Now(), count)
	//return tb.takeAvailable(tb.clock.Now(), count)
}

// Available 查看有多少可以拿
func (tb *tokenBucket) Available() int64 {
	return tb.available(time.Now())
	//return tb.available(tb.clock.Now())
}
func (tb *tokenBucket) available(now time.Time) int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.adjustAvailableTokens(tb.currentTick(now))
	return tb.availableTokens
}
