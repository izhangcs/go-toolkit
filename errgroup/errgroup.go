package errgroup

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Group struct {
	err error
	wg sync.WaitGroup
	errOne sync.Once

	workerOnce sync.Once
	ch chan func(ctx context.Context) error
	chs []func(ctx context.Context) error

	ctx context.Context
	cancel func()
}

func WithContext(ctx context.Context) *Group {
	return &Group{ctx: ctx}
}

func WithCancel(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{ctx:ctx, cancel: cancel}
}

// 设置最大执行协程数
func (g *Group) GOMAXPROCS(n int) {
	if n <= 0 {
		panic("errgroup: GOMAXPROCS must great than 0")
	}
	g.workerOnce.Do(func() {
		// 创建对应元素个数的channel
		g.ch = make(chan func(context.Context) error, n)
		for i := 0; i < n; i++ {
			go func() {
				for f := range g.ch {
					g.do(f)
				}
			}()
		}
	})
}

func (g *Group) do(f func(ctx context.Context) error) {
	ctx := g.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	defer func() {
		// recover
		if r := recover(); r != nil {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			err = fmt.Errorf("errgroup: panic recovered: %s\n%s", r, buf)
		}
		if err != nil {
			pl := fmt.Sprintf("%s ERROR:errgroup call err: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
			// 只有panic才会报警
			if strings.Contains(err.Error(), "panic") {
				fmt.Fprintf(os.Stderr, pl)
			}
			g.err = err
			if g.cancel != nil {
				g.cancel()
			}
		}
		g.wg.Done()
	}()
	err = f(ctx)
}

func (g *Group) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)
	if g.ch != nil {
		select{
		case g.ch <- f:
		default:
			g.chs = append(g.chs, f)
		}
		return
	}
	go g.do(f)
}

func (g *Group) Wait() error{
	if g.ch != nil {
		for _, f := range g.chs {
			g.ch <- f
		}
	}
	g.wg.Wait()
	if g.ch != nil {
		close(g.ch)
	}
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}