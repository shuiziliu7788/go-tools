package process

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

type Option struct {
	Ctx   context.Context
	Reset bool
}

type HandleProcess func(ctx context.Context)

type OnceProcess struct {
	ctx    context.Context
	cancel context.CancelFunc
	handle func()
	status bool
	mutex  sync.Mutex
	wg     sync.WaitGroup
}

func (o *OnceProcess) Run(ctx context.Context, handle HandleProcess, reset bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.status {
		if reset {
			o.cancel()
		} else {
			return
		}
	}
	o.wg.Wait()
	o.ctx = nil
	o.cancel = nil
	// 重置ctx
	if ctx != nil {
		o.ctx, o.cancel = context.WithCancel(ctx)
	} else {
		o.ctx, o.cancel = context.WithCancel(context.Background())
	}
	o.wg.Add(1)
	o.status = true
	go func() {
		defer func() {
			o.cancel()
			o.status = false
			o.wg.Done()
			if err := recover(); err != nil {
				fmt.Println(err)
				buf := make([]byte, 1<<16)
				if stack := runtime.Stack(buf, true); stack > 0 {
					fmt.Println(string(buf[:stack]))
				}
			}
		}()
		handle(o.ctx)
	}()
}

func (o *OnceProcess) Close() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.cancel != nil {
		o.cancel()
	}
}

func (o *OnceProcess) Wait() {
	o.wg.Wait()
}

func (o *OnceProcess) Status() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.status
}
