1:首先根据环境修改config的相关配置，通常只有两个配置要改，一个是http_proxy，如果是在可以访问外网的服务器上运行，这个配置设置为false，否则设置为true,并设置httpproxyurl为对于的代理地址
2:如果是做测试test设置为true，会去监听以太坊上测试网的地址，否则会监听主网的地址
3:address配置需要监听的地址
4：启动命令在工程目录下执行go run .