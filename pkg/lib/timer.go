package lib

import "time"

type TimerFunc = func() error

type Timer struct {
	f          TimerFunc
	quit       chan struct{}
	updateRate int
}

func (t *Timer) Run() error {
	ticker := time.NewTicker(time.Duration(t.updateRate) * time.Second)

	t.quit = make(chan struct{})

	for {
		select {
		case <-ticker.C:
			err := t.f()

			if err != nil {
				return err
			}
		case <-t.quit:
			break
		}
	}
}

func (t *Timer) Stop() error {
	close(t.quit)
	return nil
}

func NewTimer(f TimerFunc, updateRate int) *Timer {
	return &Timer{
		f:          f,
		updateRate: updateRate,
	}
}
