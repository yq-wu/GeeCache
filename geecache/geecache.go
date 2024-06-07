package geecache

import (
	"GeeCache/geecachepb"
	"GeeCache/singleflight"
	"fmt"
	"log"
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
	peers     PeerPicker
	loader    *singleflight.Group
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
		loader:    &singleflight.Group{},
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

func (g *Group) Load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) { // 匿名函数的作用是向其他缓存节点获取数据，
		// 用do来保证后面的匿名函数只实现一次即可 // do里面保证多个goroutine同时访问时， 只有1个goroutine在执行后面的匿名函数，其他的都在等待
		if g.peers != nil {
			if peer, ok := g.peers.PeerPick(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer")
			}
		}
		return g.GetLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &geecachepb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &geecachepb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
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

func (g *Group) RegisterPeers(peer PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peer
}
