## 功能简介
MyIncrLog 用于监控多个日志，并发送到钉钉进行提醒报警  
完全基于Redis，包含多客户端心跳检测，可部署多服务端进行消费  
多服务端时包含服务端心跳检测  
在客户端可开启端口4406作为web服务，点击可从钉钉进入查看错误日志前后最多4Kb的数据

## 使用方式
编译以后 先使用pem包内的genpem生成公钥与私钥，公钥用于服务端，私钥用于客户端

## 配置文件参数说明
客户端
```json
{
  "log-path": "1.log",
  "redisHost": "xxx",
  "redisPort": "6379",
  "redisPass": "xxx",
  "max-len": 1000,
  "netCardName": "以太网",
  "stream-key": "mysql:error",
  "heart-key": "mysql:heart",
  "keywords": ["ERROR", "ROLL BACK"],
  "skip-once": true
}
```
log-path 日志文件路径  
redisHost redisPort redisPass 三个参数为redis配置  
max-len为redis中stream队列的最大长度  
netCardName 对外网卡名称  
stream-key redis中stream的key名称
heart-key redis中客户端心跳检测key名称
keywords 需要监控的关键字
skip-once 第一次检测的时候是否跳过已经包含的关键字

服务端配置
```json
{
  "redisHost": "",
  "redisPort": "",
  "redisPass": "",
  "worker-id": 2,
  "ding-secret": "",
  "ding-url": "",
  "ding-proxy": "",
  "stream-key": "mysql:error",
  "heart-key": "mysql:heart",
  "sheart-key": "mysql:sheart",
  "netCardName": "以太网",
  "groupName": "mysql"
}
```
redisHost redisPort redisPass 三个参数为redis配置   
worker-id 服务端id，如配置多个服务端可配置不同的id  
ding-secret ding-url 可从钉钉创建机器人处获取  
stream-key redis中stream的key名称  
heart-key 客户端心跳检测的key
sheart-key 多个服务端的时候，服务端的心跳检测key
netCardName 网卡名称  
groupName redis中的消费组名称
## 部署案例
### 客户端部署
步骤(以mysql为例)
- 创建目录 /usr/local/dachshund  
- 上传文件client 赋执行权限  
- 上传prvKey.pem私钥  
- 上传config.json配置文件  
- 按照配置文件参数说明修改配置文件
- chown mysql:mysql /usr/local/dachshund /*
- 服务配置
```shell
cat>/usr/lib/systemd/system/dachshund.service<<-'EOF'
[unit]
Description=dachshund Client
#Documentation=

[Service]
User=mysql
Group=mysql
Environment="GIN_MODE=release"
WorkingDirectory=/usr/local/dachshund
ExecStart=/usr/local/dachshund/client
ExecReload=/bin/kill -HUP $MAINPID
ExecStop=/bin/kill -KILL $MAINPID
Type=simple
killMode=control-group
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl start dachshund
```
### 服务端部署
服务端可与redis配置到一台机器上面  
步骤：
- 创建目录 /usr/local/dachshund  
- 上传server 赋执行权限  
- 上传prvKey.pem公钥文件    
- 上传config.json配置文件  
- 按照配置文件参数说明修改配置文件
- 服务配置与启动
```shell
cat>/usr/lib/systemd/system/dachshund-server.service<<-'EOF'
[unit]
Description=dachshund Server
#Documentation=

[Service]
WorkingDirectory=/usr/local/dachshund
ExecStart=/usr/local/dachshund/server
ExecReload=/bin/kill -HUP $MAINPID
ExecStop=/bin/kill -KILL $MAINPID
Type=simple
killMode=control-group
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl start dachshund-server
```
## 设计
### 客户端
客户端每次可增量读取日志文件，
每次记录读取位置，最大化节省资源，在切割日志以后能够自动识别重置读取位置从头读取，
每次读取的内容送到matcher中的通道进行匹配，如果包含设置的关键字，则发送到sender进行发送
我们这里采用的是redis中的stream，作为消息队列，主要考虑为轻量，且效率高，易用。  
在启动以后自动发送当前机器的时间戳到redis中，作为心跳检测。  
在客户端启动了一个小的web服务，web服务为内网服务，用于在钉钉中直接查看并且添加了
rsa加密算法保障安全，防止错误日志外泄，并且设置了访问过期时间。
### 服务端
服务端使用用户组消费redis中的消息队列，默认开启了5个协程用于消费数据，每个协程单独开启协程
用于检测发送失败的消息，由于使用了管道，非常方便地就能进行发送频率限制，防止发送过快钉钉禁用机器人
服务端有私钥用于加密web地址，并发送到钉钉，显示简略信息。在服务端定期循环挑选心跳检测
信息用于检测是否有机器下线，做了防止重复提示机器下线的机制。
由于和客户端松耦合，可以布置多个服务端，可以设置不同的钉钉机器人报警。  
并且服务端之间也有心跳检测机制，如果有服务端掉线超过指定时间也会进行报警。
