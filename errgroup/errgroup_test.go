package errgroup

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

var ExitError = errors.New("exit error")


// 普通的errgroup 不会因为一个任务导致全部任务失败
func TestNormal(t *testing.T) {
	var (
		g    Group
		err  error
	)
	g.Go(func(context.Context) (err error) {
		time.Sleep(time.Second)
		fmt.Println("sleep 1s")
		return ExitError
	})
	g.Go(func(context.Context) (err error) {
		time.Sleep(2*time.Second)
		fmt.Println("sleep 2s")
		return
	})
	if err = g.Wait(); err != ExitError {
		t.Log(err)
	}
}

func TestWithCancel(t *testing.T) {
	//ctx, cancel := context.WithCancel(ctx)
}