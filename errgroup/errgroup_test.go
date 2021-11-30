package errgroup

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
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

// 带cancel的errgroup
func TestWithCancel(t *testing.T) {
	g := WithCancel(context.Background())
	g.Go(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("boom")
	})
	var doneErr error
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			doneErr = ctx.Err()
		}
		return doneErr
	})
	_ = g.Wait()
	if doneErr != context.Canceled {
		t.Error("error should be Canceled")
	}
}

func sleep1s(ctx context.Context) error {
	time.Sleep(time.Second)
	return nil
}

// 并发控制的errgroup
func TestGOMAXPROCS(t *testing.T) {
	// 没有并发控制
	g := Group{}
	now := time.Now()
	g.Go(sleep1s)
	g.Go(sleep1s)
	g.Go(sleep1s)
	g.Go(sleep1s)
	err := g.Wait()
	assert.Nil(t, err)
	sec := math.Round(time.Since(now).Seconds())
	if sec != 1 {
		t.FailNow()
	}
	// 限制并发数
	g2 := Group{}
	g2.GOMAXPROCS(2)
	now = time.Now()
	g2.Go(sleep1s)
	g2.Go(sleep1s)
	g2.Go(sleep1s)
	g2.Go(sleep1s)
	err = g2.Wait()
	assert.Nil(t, err)
	sec = math.Round(time.Since(now).Seconds())
	if sec != 2 {
		t.FailNow()
	}
	// context cancel
	var canceled bool
	g3 := WithCancel(context.Background())
	g3.GOMAXPROCS(2)
	g3.Go(func(ctx context.Context) error {
		return fmt.Errorf("error for testing errgroup context")
	})
	g3.Go(func(ctx context.Context) error {
		time.Sleep(time.Second)
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}
		return nil
	})
	err = g3.Wait()
	assert.Nil(t, err)
	if !canceled {
		t.FailNow()
	}
}