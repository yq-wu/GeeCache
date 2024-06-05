package GeeCache

import (
	"fmt"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache Cache
}

var (
	groups = make(map[string]*Group)
	mu     sync.RWMutex
)

func NewGroup(name string, cacheByte int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      name,
		mainCache: Cache{cacheBytes: cacheByte},
		getter:    getter,
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if value, ok := g.mainCache.Get(key); ok {
		return value, nil
	}
	return g.Load(key)
}

func (g *Group) Load(key string) (ByteView, error) {
	return g.GetLocally(key)
}

func (g *Group) GetLocally(key string) (ByteView, error) {
	v, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: CloneByte(v)}
	g.PopulateCache(key, value)
	return value, nil
}

func (g *Group) PopulateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}
