package ratelimit

import (
	"sync"
	"time"
)

const infinityDuration time.Duration = 0x7fffffffffffffff

type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

type c struct{}

// Now 返回当前时间
func (c c) Now() time.Time {
	return time.Now()
}

// Sleep 睡眠 d 时间
func (c c) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Bucket 令牌桶
type Bucket struct {
	// 用来计时
	clock Clock

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

func NewBucket(fillInterval time.Duration, capacity int64) *Bucket {
	return NewBucketWithClock(fillInterval, capacity, nil)
}
func NewBucketWithClock(fillInterval time.Duration, capacity int64, clock Clock) *Bucket {
	return NewBucketWithQuantumAndClock(fillInterval, capacity, 1, clock)
}
func NewBucketWithQuantum(fillInterval time.Duration, capacity, quantum int64) *Bucket {
	return NewBucketWithQuantumAndClock(fillInterval, capacity, quantum, nil)
}
func NewBucketWithQuantumAndClock(fillInterval time.Duration, quantum, capacity int64, clock Clock) *Bucket {
	if clock == nil {
		clock = c{}
	}

	if fillInterval <= 0 {
		panic("token bucket fill interval is not > 0")
	}
	if capacity <= 0 {
		panic("token bucket capacity is not > 0")
	}
	if quantum <= 0 {
		panic("token bucket quantum is not > 0")
	}

	return &Bucket{
		clock:           clock,
		startTime:       clock.Now(),
		capacity:        capacity,
		quantum:         quantum,
		fillInterval:    fillInterval,
		mu:              sync.Mutex{},
		availableTokens: capacity,
		latestTick:      0,
	}
}

// 获取现在的间隔数
func (b *Bucket) currentTick(now time.Time) int64 {
	return int64(now.Sub(b.startTime) / b.fillInterval)
}

// 根据过去的间隔来更新令牌数
func (b *Bucket) adjustAvailableTokens(tick int64) {
	latestTick := b.latestTick
	b.latestTick = tick
	if b.availableTokens >= b.capacity {
		return
	}

	b.availableTokens += (tick - latestTick) * b.quantum
	if b.availableTokens > b.capacity {
		b.availableTokens = b.capacity
	}

	return
}

// WaitMaxDuration 等待一段时间拿令牌，拿不到就返回 false
func (b *Bucket) WaitMaxDuration(count int64, maxWait time.Duration) bool {
	d, ok := b.TakeMaxDuration(count, maxWait)
	// 根据 take 的代码推出，其实只有在 这段时间内可以拿到对应数目的令牌才会 sleep
	if d > 0 {
		b.clock.Sleep(d)
	}
	return ok
}

// TakeMaxDuration 返回 能拿到的话需要的时间 和 是否能拿到
func (b *Bucket) TakeMaxDuration(count int64, maxWait time.Duration) (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.take(b.clock.Now(), count, maxWait)
}
func (b *Bucket) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}
	// 拿到现在的间隔数
	tick := b.currentTick(now)
	b.adjustAvailableTokens(tick)

	avail := b.availableTokens - count
	// 获取成功，直接返回
	if avail >= 0 {
		return 0, true
	}

	// avail < 0
	// 这里开始算我要等多少个间隔才能拿到
	endTick := tick + (-avail+b.quantum-1)/b.quantum
	endTime := b.startTime.Add(time.Duration(endTick) * b.fillInterval)
	waitTime := endTime.Sub(now)
	// 等了还是拿不到
	if waitTime > maxWait {
		return 0, false
	}
	// 等了时间内可以拿到
	b.availableTokens = avail
	return waitTime, true
}

// Wait 一直等待到可用
func (b *Bucket) Wait(count int64) {
	if d := b.Take(count); d > 0 {
		b.clock.Sleep(d)
	}
}

// Take 返回 count 个 拿到需要的时间
func (b *Bucket) Take(count int64) time.Duration {
	b.mu.Lock()
	b.mu.Unlock()
	d, _ := b.take(b.clock.Now(), count, infinityDuration)
	return d
}

// 有多少拿多少
func (b *Bucket) takeAvailable(now time.Time, count int64) int64 {
	if count <= 0 {
		return 0
	}
	b.adjustAvailableTokens(b.currentTick(now))
	if b.availableTokens <= 0 {
		return 0 // 拿不到就返回
	}
	// 如果要的令牌不够多
	if count > b.availableTokens {
		count = b.availableTokens // 拿到尽可能多的令牌
	}

	b.availableTokens -= count
	return count
}
func (b *Bucket) TakeAvailable(count int64) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.takeAvailable(b.clock.Now(), count)
}

// Available 查看有多少可以拿
func (b *Bucket) Available() int64 {
	return b.available(b.clock.Now())
}
func (b *Bucket) available(now time.Time) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.adjustAvailableTokens(b.currentTick(now))
	return b.availableTokens
}
