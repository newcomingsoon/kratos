package p2c

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/node/ewma"
)

const (
	//
	forcePick = time.Second * 3
	// Name is balancer name
	Name = "p2c"
)

// 实现原理： 随机挑选两个node(n1,n2)
// 在挑选的两个node（n1,n2）中
// 一般情况下：选择权重大的那一个, 假设为n2， 那n1则为未被选中，直接返回n2，并退出
// 特殊情况：如果n1距离上次被选中超过了一定时间间隔（forcePick的定义）
// 那么强制选择该节点（n1），放弃之前选择的权重大的节点（n2）， 并返回n1，然后退出

var _ selector.Balancer = &Balancer{}

// WithFilter with select filters
func WithFilter(filters ...selector.Filter) Option {
	return func(o *options) {
		o.filters = filters
	}
}

// Option is random builder option.
type Option func(o *options)

// options is random builder options
type options struct {
	filters []selector.Filter
}

// New creates a p2c selector.
func New(opts ...Option) selector.Selector {
	return NewBuilder(opts...).Build()
}

// Balancer is p2c selector.
type Balancer struct {
	r *rand.Rand
	// 用CAS来充当锁
	lk int64
}

// choose two distinct nodes.
// 预筛选两个不同的node
func (s *Balancer) prePick(nodes []selector.WeightedNode) (nodeA selector.WeightedNode, nodeB selector.WeightedNode) {
	a := s.r.Intn(len(nodes))
	b := s.r.Intn(len(nodes) - 1)
	if b >= a {
		b = b + 1
	}
	nodeA, nodeB = nodes[a], nodes[b]
	return
}

// Pick pick a node.
func (s *Balancer) Pick(ctx context.Context, nodes []selector.WeightedNode) (selector.WeightedNode, selector.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector.ErrNoAvailable
	} else if len(nodes) == 1 {
		done := nodes[0].Pick()
		return nodes[0], done, nil
	}

	var pc, upc selector.WeightedNode
	nodeA, nodeB := s.prePick(nodes)
	// meta.Weight为服务发布者在discovery中设置的权重
	// 从预选中进一步判断权重大小
	// pc： 权重大的，要当选的节点
	// upc： 权重小的，本次落选的
	if nodeB.Weight() > nodeA.Weight() {
		pc, upc = nodeB, nodeA
	} else {
		pc, upc = nodeA, nodeB
	}

	// 如果落选节点在forceGap期间内从来没有被选中一次，则强制选一次
	// 利用强制的机会，来触发成功率、延迟的更新
	if upc.PickElapsed() > forcePick && atomic.CompareAndSwapInt64(&s.lk, 0, 1) {
		pc = upc
		atomic.StoreInt64(&s.lk, 0)
	}
	done := pc.Pick()
	return pc, done, nil
}

// NewBuilder returns a selector builder with p2c balancer
func NewBuilder(opts ...Option) selector.Builder {
	var option options
	for _, opt := range opts {
		opt(&option)
	}
	return &selector.DefaultBuilder{
		Filters:  option.filters,
		Balancer: &Builder{},
		Node:     &ewma.Builder{},
	}
}

// Builder is p2c builder
type Builder struct{}

// Build creates Balancer
func (b *Builder) Build() selector.Balancer {
	return &Balancer{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}
