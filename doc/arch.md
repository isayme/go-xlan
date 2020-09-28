# 目标

通过公网访问家里内网的服务.

# 限制条件

家里的宽带没有公网 IP, 无法通过路由器的端口转发功能实现.

# 解决方案

方案为`内网穿透`, 目前有开源产品 [frp](https://github.com/fatedier/frp).

# 设计

整个链路由 4 部分组成: 用户(user), 内网穿透服务端(xlan-s), 内网穿透客户端(xlan-c), 最终服务(service), 其中 xlan-s 有公网 IP 地址.

# 流程

1. xlan-s 启动并监听端口 P1, 等待 xlan-c 连接;
2. xlan-c 启动并与 xlan-s 端口 P1 建立连接 C1;
3. 连接建立后, 发送 registerAsControl 命令, 声明这个连接用于控制;
4. 使用控制连接 C1, 向服务端发送 registerService 命令注册服务;
5. xlan-s 收到 registerService 后为启动对应服务, 监听端口 P2;
6. user 向端口 P2 发起请求建立连接 C2, xlan-s 通过连接 C1 向 xlan-c 发送 NewConnectionFromUser 命令;
7. xlan-c 收到命令后, 向 service 建立连接 C3;
8. xlan-c 收到命令后, 向 xlan-s 端口 P1 建立连接 C4 并发送 RegisterAsProxy 声明此连接用于传输用户数据;
9. xlan-s 将连接 C2 和 C4 的数据进行交换, xlan-c 将连接 C3 和 C4 的数据进行交换;
10. 最终实现 C2 和 C3 的数据交换, 等同于 user 与 service 进行通信, 达到穿透目的.
