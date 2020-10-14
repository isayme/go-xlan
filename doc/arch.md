# 目标

通过公网访问家里内网的服务.

# 限制条件

家里的宽带没有公网 IP, 无法通过路由器的端口转发功能实现.

# 解决方案

方案为`内网穿透`, 目前有开源产品 [frp](https://github.com/fatedier/frp).

# 设计

见: [xlan: Golang 实现一个内网穿透工具](https://github.com/isayme/blog/issues/52)
