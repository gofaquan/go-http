package ratelimit

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type rateLimitSuite struct {
	suite.Suite
	takeTests          []takeTest
	availTests         []availTest
	takeAvailableTests []takeAvailableTest
}

type takeTest struct {
	name         string
	fillInterval time.Duration
	capacity     int64

	reqs []struct {
		time       time.Duration
		count      int64
		expectWait time.Duration
	}
}

type availTest struct {
	name         string
	capacity     int64
	fillInterval time.Duration
	take         int64
	sleep        time.Duration

	expectCountAfterTake  int64
	expectCountAfterSleep int64
}
type takeAvailableTest struct {
	name         string
	fillInterval time.Duration
	capacity     int64
	reqs         []struct {
		time   time.Duration
		count  int64
		expect int64
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	suite.Run(t, new(rateLimitSuite))
}

func (r *rateLimitSuite) SetupSuite() {
	r.takeTests = []takeTest{{
		name:         "serial requests",
		fillInterval: 250 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time       time.Duration
			count      int64
			expectWait time.Duration
		}{{
			time:       0,
			count:      0,
			expectWait: 0,
		}, {
			time:       0,
			count:      10,
			expectWait: 0,
		}, {
			time:       0,
			count:      1,
			expectWait: 250 * time.Millisecond,
		}, {
			time:       250 * time.Millisecond,
			count:      1,
			expectWait: 250 * time.Millisecond,
		}},
	}, {
		name:         "concurrent requests",
		fillInterval: 250 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time       time.Duration
			count      int64
			expectWait time.Duration
		}{{
			time:       0,
			count:      10,
			expectWait: 0,
		}, {
			time:       0,
			count:      2,
			expectWait: 500 * time.Millisecond,
		}, {
			time:       0,
			count:      2,
			expectWait: 1000 * time.Millisecond,
		}, {
			time:       0,
			count:      1,
			expectWait: 1250 * time.Millisecond,
		}},
	}, {
		name:         "more than capacity",
		fillInterval: 1 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time       time.Duration
			count      int64
			expectWait time.Duration
		}{{
			time:       0,
			count:      10,
			expectWait: 0,
		}, {
			time:       20 * time.Millisecond,
			count:      15,
			expectWait: 5 * time.Millisecond,
		}},
	}, {
		name:         "sub-quantum time",
		fillInterval: 10 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time       time.Duration
			count      int64
			expectWait time.Duration
		}{{
			time:       0,
			count:      10,
			expectWait: 0,
		}, {
			time:       7 * time.Millisecond,
			count:      1,
			expectWait: 3 * time.Millisecond,
		}, {
			time:       8 * time.Millisecond,
			count:      1,
			expectWait: 12 * time.Millisecond,
		}},
	}, {
		name:         "within capacity",
		fillInterval: 10 * time.Millisecond,
		capacity:     5,
		reqs: []struct {
			time       time.Duration
			count      int64
			expectWait time.Duration
		}{{
			time:       0,
			count:      5,
			expectWait: 0,
		}, {
			time:       60 * time.Millisecond,
			count:      5,
			expectWait: 0,
		}, {
			time:       60 * time.Millisecond,
			count:      1,
			expectWait: 10 * time.Millisecond,
		}, {
			time:       80 * time.Millisecond,
			count:      2,
			expectWait: 10 * time.Millisecond,
		}},
	}}
	r.availTests = []availTest{{
		name:                  "should fill tokens after interval",
		capacity:              5,
		fillInterval:          time.Second,
		take:                  5,
		sleep:                 time.Second,
		expectCountAfterTake:  0,
		expectCountAfterSleep: 1,
	}, {
		name:                  "should fill tokens plus existing count",
		capacity:              2,
		fillInterval:          time.Second,
		take:                  1,
		sleep:                 time.Second,
		expectCountAfterTake:  1,
		expectCountAfterSleep: 2,
	}, {
		name:                  "shouldn't fill before interval",
		capacity:              2,
		fillInterval:          2 * time.Second,
		take:                  1,
		sleep:                 time.Second,
		expectCountAfterTake:  1,
		expectCountAfterSleep: 1,
	}, {
		name:                  "should fill only once after 1*interval before 2*interval",
		capacity:              2,
		fillInterval:          2 * time.Second,
		take:                  1,
		sleep:                 3 * time.Second,
		expectCountAfterTake:  1,
		expectCountAfterSleep: 2,
	}}
	r.takeAvailableTests = []takeAvailableTest{{
		name:         "serial requests",
		fillInterval: 250 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time   time.Duration
			count  int64
			expect int64
		}{{
			time:   0,
			count:  0,
			expect: 0,
		}, {
			time:   0,
			count:  10,
			expect: 10,
		}, {
			time:   0,
			count:  1,
			expect: 0,
		}, {
			time:   250 * time.Millisecond,
			count:  1,
			expect: 1,
		}},
	}, {
		name:         "concurrent requests",
		fillInterval: 250 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time   time.Duration
			count  int64
			expect int64
		}{{
			time:   0,
			count:  5,
			expect: 5,
		}, {
			time:   0,
			count:  2,
			expect: 2,
		}, {
			time:   0,
			count:  5,
			expect: 3,
		}, {
			time:   0,
			count:  1,
			expect: 0,
		}},
	}, {
		name:         "more than capacity",
		fillInterval: 1 * time.Millisecond,
		capacity:     10,
		reqs: []struct {
			time   time.Duration
			count  int64
			expect int64
		}{{
			time:   0,
			count:  10,
			expect: 10,
		}, {
			time:   20 * time.Millisecond,
			count:  15,
			expect: 10,
		}},
	}, {
		name:         "within capacity",
		fillInterval: 10 * time.Millisecond,
		capacity:     5,
		reqs: []struct {
			time   time.Duration
			count  int64
			expect int64
		}{{
			time:   0,
			count:  5,
			expect: 5,
		}, {
			time:   60 * time.Millisecond,
			count:  5,
			expect: 5,
		}, {
			time:   70 * time.Millisecond,
			count:  1,
			expect: 1,
		}},
	}}
}

func (r *rateLimitSuite) TestTake() {
	t := r.T()
	for i, test := range r.takeTests {
		tb, err := NewBucket(test.fillInterval, test.capacity)
		assert.Nil(t, err)

		for j, req := range test.reqs {
			d, ok := tb.take(tb.startTime.Add(req.time), req.count, infinityDuration)
			assert.Equal(t, ok, true)
			if d != req.expectWait {
				t.Fatalf("test %d.%d, %s, got %v want %v", i, j, test.name, d, req.expectWait)
			}
		}
	}
}

func (r *rateLimitSuite) TestTakeMaxDuration() {
	t := r.T()
	for i, test := range r.takeTests {
		tb, err := NewBucket(test.fillInterval, test.capacity)
		assert.Nil(t, err)
		for j, req := range test.reqs {
			if req.expectWait > 0 {
				d, ok := tb.take(tb.startTime.Add(req.time), req.count, req.expectWait-1)
				assert.Equal(t, ok, false)
				assert.Equal(t, d, time.Duration(0))
			}
			d, ok := tb.take(tb.startTime.Add(req.time), req.count, req.expectWait)
			assert.Equal(t, ok, true)
			if d != req.expectWait {
				t.Fatalf("test %d.%d, %s, got %v want %v", i, j, test.name, d, req.expectWait)
			}
		}
	}
}

func (r *rateLimitSuite) TestTakeAvailable() {
	t := r.T()
	for i, test := range r.takeAvailableTests {
		tb, err := NewBucket(test.fillInterval, test.capacity)
		assert.Nil(t, err)
		for j, req := range test.reqs {
			d := tb.takeAvailable(tb.startTime.Add(req.time), req.count)
			if d != req.expect {
				t.Fatalf("test %d.%d, %s, got %v want %v", i, j, test.name, d, req.expect)
			}
		}
	}
}

func (r *rateLimitSuite) TestUnexpectedErr() {
	t := r.T()
	assert.Equal(t, func() error {
		_, err := NewBucket(0, 1)
		return err
	}(), fillIntervalErr)
	assert.Equal(t, func() error {
		_, err := NewBucket(-2, 1)
		return err
	}(), fillIntervalErr)
	assert.Equal(t, func() error {
		_, err := NewBucket(1, 0)
		return err
	}(), capacityErr)
	assert.Equal(t, func() error {
		_, err := NewBucket(1, -2)
		return err
	}(), capacityErr)
}

func (r *rateLimitSuite) TestAvailable() {
	t := r.T()
	for i, test := range r.availTests {
		tb, err := NewBucket(test.fillInterval, test.capacity)
		assert.Nil(t, err)
		if c := tb.takeAvailable(tb.startTime, test.take); c != test.take {
			t.Fatalf("#%d: %s, take = %d, want = %d", i, test.name, c, test.take)
		}
		if c := tb.available(tb.startTime); c != test.expectCountAfterTake {
			t.Fatalf("#%d: %s, after take, available = %d, want = %d", i, test.name, c, test.expectCountAfterTake)
		}
		if c := tb.available(tb.startTime.Add(test.sleep)); c != test.expectCountAfterSleep {
			t.Fatalf("#%d: %s, after some time it should fill in new tokens, available = %d, want = %d",
				i, test.name, c, test.expectCountAfterSleep)
		}
	}

}

func (r *rateLimitSuite) TestNoBonusTokenAfterBucketIsFull() {
	t := r.T()
	tb, err := NewBucket(time.Second*1, 100, WithQuantum(20))
	assert.Nil(t, err)

	curAvail := tb.Available()
	if curAvail != 100 {
		t.Fatalf("initially: actual available = %d, expected = %d", curAvail, 100)
	}

	time.Sleep(time.Second * 5)

	curAvail = tb.Available()
	if curAvail != 100 {
		t.Fatalf("after pause: actual available = %d, expected = %d", curAvail, 100)
	}

	cnt := tb.TakeAvailable(100)
	if cnt != 100 {
		t.Fatalf("taking: actual taken count = %d, expected = %d", cnt, 100)
	}

	curAvail = tb.Available()
	if curAvail != 0 {
		t.Fatalf("after taken: actual available = %d, expected = %d", curAvail, 0)
	}
}

func BenchmarkWait(b *testing.B) {
	tb, _ := NewBucket(1, 16*1024)

	for i := b.N - 1; i >= 0; i-- {
		tb.Wait(1)
	}
}
