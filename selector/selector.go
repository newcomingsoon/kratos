package selector

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
)

// ErrNoAvailable is no available node.
var ErrNoAvailable = errors.ServiceUnavailable("no_available_node", "")

// Selector is node pick balancer.
// Node节点筛选器
// 其中Rebalancer是用来做节点rebalance更新用的，当有新增或删除节点时
// 将最新的节点信息同步到筛选器中
type Selector interface {
	Rebalancer

	// Select nodes
	// if err == nil, selected and done must not be empty.
	// 用来真正筛选Node的实现： 其中可能是多种不同负载均衡的具体实现
	// 常见的包括，轮询，权重， 最少链接数等等实现
	Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error)
}

// Rebalancer is nodes rebalancer.
type Rebalancer interface {
	// apply all nodes when any changes happen
	// 这个方法被调用后，内部才有了节点相关的数据
	// 具体Node需要定义哪些数据，自行定义在Node中即可
	Apply(nodes []Node)
}

// Builder build selector
// Selector 通过 Builder 抽象来构建一个选择器
// 这样就意味着 Builder的实现也是可定制的
// 这样是通过接口抽象的好处，可以不用关心具体的实现细节
type Builder interface {
	Build() Selector
}

// Node is node interface.
// Node也定义成了一个接口抽象，具体需要出传入的数据也变的更加灵活
// 这个接口就定义一些基础必备的参数即可， 如果有特殊需要，可以通过嵌入该Node引入新的接口实现
// 传递更多自定义的的参数到selector中
type Node interface {
	// Address is the unique address under the same service
	Address() string

	// ServiceName is service name
	ServiceName() string

	// InitialWeight is the initial value of scheduling weight
	// if not set return nil
	InitialWeight() *int64

	// Version is service node version
	Version() string

	// Metadata is the kv pair metadata associated with the service instance.
	// version,namespace,region,protocol etc..
	Metadata() map[string]string
}

// DoneInfo is callback info when RPC invoke done.
// 自定义一个DoneInfo字段，来接收调用结束的消息数据
// 简单的只是将grpc内部balancer中DoneInfo的数据进行装换
// 将内部metata.MD 的结构转换为ReplyMeta，保留一些重要字段
// 这样做的目的，可能是需要处理Done函数被调用后的一些自定义逻辑
// 通过将PickResult中Done注入DoneFunc后，可以在调用Done是执行注入的DoneFunc来做一些自定义的逻辑
type DoneInfo struct {
	// Response Error
	Err error
	// Response Metadata
	ReplyMeta ReplyMeta

	// BytesSent indicates if any bytes have been sent to the server.
	BytesSent bool
	// BytesReceived indicates if any byte has been received from the server.
	BytesReceived bool
}

// ReplyMeta is Reply Metadata.
// 可以通过定义一个自定义类型来 `type Tailer metadata.MD`
// 来简化对metadata.MD进行操作, 这个自定义类型实现ReplyMeta接口即可
// 这里本质对应的还是metadata.MD数据结构, 通过操作接口化，来简化从map中取出key值对应value的操作
// 通过具体实现的Get方法，可以处理一些可能的异常判断逻辑，使得取值更加通用简洁
// 这里有没有体会到接口的强大之处👍👍
type ReplyMeta interface {
	Get(key string) string
}

// DoneFunc is callback function when RPC invoke done.
type DoneFunc func(ctx context.Context, di DoneInfo)
