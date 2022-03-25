# MySQL binlog 跟踪工具

没啥，就是个模拟mysql从库，去主库里去订阅binlog，然后把binlog往kafka这样的消息队列里去丢，以后可能会加入其它的mq



## 安装

release 里面的包免安装，直接运行就行了。懒的话就直接下载运行吧，别安装了。

## 编译

- GoLang (1.13 +) 
- node 14.16  (如果你要编译web代码的话)
- 环境变量：GO111MODULE=on

###### 编译web(可选)

1. 进入web目录   

   ```bash
   $ cd web
   ```

2. 安装依赖  

   ```bash
   $ yarn install
   ```

3. 编译  

   ```bash
   $ yarn run build
   ```

4. 集成到程序

   1. 安装go-bindata   

      ```shell
      $ go get -u github.com/go-bindata/go-bindata/...
      ```

   2. 生成 bindata.go 文件

      ```shell
      $ go-bindata ./assets/...
      ```

###### 生成运行文件

```bash
$ go build -o bin/go-mysql-replication 
```



## 大致原理

![](https://raw.githubusercontent.com/buzhiyun/go-mysql-replication/master/screenshot/screenshot-2.png)



## 运行

###### 配置文件

- yaml 文件：

  ```yaml
  # mysql配置
  addr: localhost:3306
  user: mysql_user
  pass: mysql_passwd
  charset : utf8
  slave_id: 1001 #slave ID
  flavor: mysql #mysql or mariadb,默认mysql
  from_gtid_file: gtid.txt
  #只dump 某些db
  only_dbs:
    - log
  # 只dump 某些table , 一个db只能配置一组
  #only_tables:
  #  - log_202106

  logger:
    level: info #日志级别；支持：debug|info|warn|error，默认info
  
  #maxprocs: 50 #并发协（线）程数量，默认为: CPU核数*2；一般情况下不需要设置此项
  bulk_size: 100 #每批处理数量，不写默认100，可以根据带宽、机器性能等调整;如果是全量数据初始化时redis建议设为1000，其他接收端酌情调大
  
  
  #web admin相关配置
  enable_web_admin: true #是否启用web admin，默认false
  web_admin_port: 8080 #web监控端口,默认8060
  
  
  #目标类型
  target: kafka # 支持redis、mongodb、elasticsearch、rocketmq、kafka、rabbitmq
  
  #kafka连接配置
  kafka_addrs: kafka.local:9092 #kafka连接地址，多个用逗号分隔
  kafka_topic: log
  kafka_version: 0.10.2.2
  #kafka_sasl_user:  #kafka SASL_PLAINTEXT认证模式 用户名
  #kafka_sasl_password: #kafka SASL_PLAINTEXT认证模式 密码
  
  
  notice: # 告警通知
    septnet_msg:
      api: http://septnet/
      to_users: "177"   # 接收企业微信的用户ID，多个用户用|分隔
  ```



###### 运行方式

> 没有依赖，直接运行即可

- 直接运行

  ```shell
  ./bin/go-mysql-replication -c config.yml
  ```

  

- docker运行
  
  ```shell
  sudo docker run  --rm --name=test -p 8080:8099 -v /tmp/config.yml:/app/config.yml -v -v /tmp/gtid.txt:/app/gtid.txt test:latest
  ```
  
- docker-compose 运行：

  ```yaml
  version: '3.1'
  services:
    go-mysql-replication:
      container_name: mysql-replication
      image: test:latest
      ports:
        - "8080:8080"
      volumes:
        - ./config.yml:/app/config.yml
        - ./gtid.txt:/app/gtid.txt
  ```



## 运行web截图

![](https://raw.githubusercontent.com/buzhiyun/go-mysql-replication/master/screenshot/screenshot-1.png)
