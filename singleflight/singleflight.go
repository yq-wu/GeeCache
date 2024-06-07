package singleflight

import "sync"

type Call struct {
	val interface{}
	err error
	wg  sync.WaitGroup
}

// waitgroup 针对 goroutine，让一个goroutine进行，其他的阻塞等待goroutine完成。
// mutex 针对 资源，只有一个goroutine修改，其他的不能修改。
type Group struct {
	mu sync.Mutex
	m  map[string]*Call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*Call)
	}
	if call, ok := g.m[key]; ok {
		g.mu.Unlock()
		call.wg.Wait()
		return call.val, call.err
	}
	call := new(Call)
	call.wg.Add(1)
	g.m[key] = call
	g.mu.Unlock()
	call.val, call.err = fn()
	call.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return call.val, call.err
}
