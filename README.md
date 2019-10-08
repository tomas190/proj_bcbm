# 奔驰宝马游戏部署

## 文件目录

```bazaar
├─bin
│  ├─conf
│  ├─gamedata
│  └─log
├─pkg
│  └─mod
│      └─cache
└─src
    ├─client
    │  ├─common
    │  └─test
    └─server
        ├─base
        ├─conf
        ├─constant
        ├─game
        │  └─internal
        ├─gamedata
        ├─gate
        │  └─internal
        ├─login
        │  └─internal
        ├─msg
        └─util
```

## 套件支持

```bazaar
golang version go1.13 
mongodb 建议版本 4.0
```

## 配置文件

### 位置 

```
proj_bcbm/bin/conf/server.json
```

### 内容

```
{
	"LogLevel": "debug",
	"LogPath": "",
	"WSAddr": "0.0.0.0:1220",
	"MaxConnNum": 20000,

	"TokenServer":"http://172.16.100.2:9502/Token/getToken",
	"CenterServer": "http://172.16.100.2:9502",
	"CenterServerPort": "9502",
	"DevKey": "new_game_17",
	"DevName": "新游戏开发",
	"GameID": "5b1f3a3cb76a591e7f251716",
	"MongoDB": "mongodb://bcbm:123456@172.16.100.5:27017"
}

```
- ```TokenServer CenterServer CenterServerPort``` 中心服配置
- ```DevKey DevName GameID``` 游戏配置
- ```MongoDB``` MongoDB地址（包含用户名和密码）

## 日志

### 位置

```
proj_bcbm/bin/nohup.out
```

## 部署流程

- clone 代码到对应目录（地址可能不同）
```
git clone http://joel:20190506@git.0717996.com/Joel/proj_bcbm.git
```

- build

```bazaar
sh build.sh
```

- run

```bazaar
sh run.sh
```

- 检查启动提示

```bazaar
tail -f proj_bcbm/bin/nohup.out
```

```
2019/09/11 16:33:09 [debug  ] 读取配置文件 conf/server.json...
2019/09/11 16:33:09 [release] Leaf 1.1.3 starting up
2019/09/11 16:33:09 [debug  ] 连接中心服 ws://swoole.0717996.com
2019/09/11 16:33:09 [debug  ] 请求Token http://swoole.0717996.com/Token/getToken?dev_key=new_game_17&dev_name=%E6%96%B0%E6%B8%B8%E6%88%8F%E5%BC%80%E5%8F%91
2019/09/11 16:33:10 [debug  ] Msg to center {"event":"/GameServer/Login/login","data":{"host":"http://swoole.0717996.com","port":"9502","game_id":"5b1f3a3cb76a591e7f251716","token":"3e8324cbd454a7327702b21f66921d7d31f8550d","dev_key":"new_game_17"}}
2019/09/11 16:33:10 [debug  ] 数据库客户端 mongodb://10.63.90.53:27917 创建成功...
2019/09/11 16:33:10 [debug  ] 数据库连接成功...
2019/09/11 16:33:10 [debug  ] Msg from center {"event":"\/GameServer\/Login\/login","data":{"status":"SUCCESS","code":200,"msg":{"platform_tax_percent":6}}}
2019/09/11 16:33:10 [debug  ] 服务器登陆 SUCCESS 税率 %6 ...
```
