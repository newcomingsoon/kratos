# LEARNING

## metadata

通过区分server和client的传入context中的数据，定义一些操作context中数据的方法

## middleware

### 1. metadata
一般context中的数据都是在中间件中注入的
server读取client中带来的数据
client初始化化一些数据，请求带给server

## Transport

### 1. gRPC

接口抽象定义：

- Balancer -> Selector  -> reBalancer

reBalancer 通过调用Apply(Nodes), 最终注入通过注册中心或者直连方式拿到的Node数据， 这样selector就可以基于具体策略来进行Node筛选了。

- resolver -> registry.Discovery

resolver 通过接入注册中心，来获取服务有效的endpoint，最后注入服务对应的连接（ClientConn接口的具体实现）， 其中selector中也是通过这个连接来进行筛选的（准确来说应该是SubConn），比如负载均衡中的最少链接法，就是每次筛选出来的链接中通过引入计数器来记入该链接上的客户端数量，最后通过比较来筛选出最少建立链接数的那个


#### [Selector](/selector/selector.go)

负载均衡算法： 随机，加权轮询, p2c


### 2. Http

### binding

通过将proto.message中定义的字段，映射到URL中定义placeholder中

### status

http返回Code和grpc返回code的映射，它们之间可以相互装换
 
