package hans

import "time"

var (
	maxBadExits       = 5
	maxBadExitsWindow = "15s"
)

type BadExit struct {
	N    int
	Mark time.Time
	Ko   bool
}

func (b *BadExit) Init() {
	if b.Mark.IsZero() {
		b.Mark = time.Now()
	}
}

func (b *BadExit) Inc() {
	b.N++
}

func (b *BadExit) MaxReached() bool {
	return b.N > maxBadExits
}

func (b *BadExit) WithinWindow() bool {
	d, _ := time.ParseDuration(maxBadExitsWindow)
	return time.Now().Sub(b.Mark) < d
}

func (b *BadExit) Reset() {
	b.N = 0
	b.Mark = time.Now()
}
