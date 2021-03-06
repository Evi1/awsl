# awsl
A WebSocket Linker  

## How To

windows10可使用wsl

1. 从release下载对应版本
2. 运行`awsl -c $configfile`

没有`-c $configfile`时读取/etc/awsl/config.json
配置文件参考`test/server.config.json`和`test/client.config.json` 以及`Caddfile`，配合caddy使用时server端不需要key和crt

## WSL解决方案
1. 创建wsl_run.sh
```
#!/bin/bash
# service privoxy start >/dev/null 2>&1    # 可使用privoxy完成http -> socks5 目前已支持http
$GOPATH/bin/awsl > $logfile 2>&1
```
2. `chmod +x wsl_run.sh`
3. `sudo visudo`  
添加  `$username ALL=(root) NOPASSWD: $path_to_wsl_run.sh`
4. windows中新建`wsl.vbs`  
`$wsl_name` 可在cmd中使用 `wsl -l`查看
```
Set ws = CreateObject("Wscript.Shell") 
ws.run "wsl -d $wsl_name -e sudo $path_to_wsl_run.sh", vbhide
```
5. 使用任务计划程序添加vbs  
## windows
[使用vbs管理awsl.exe](https://blog.bilibili.network/posts/vbs_service/ "vbs")

## awsl.service demo
```
[Unit]
Description=awsl
After=network-online.target network.target
Wants=network-online.target systemd-networkd-wait-online.service network.target

[Service]
Restart=on-failure
User=user
Group=users
ExecStart=awsl -m 80
ExecReload=/bin/kill $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=5s
StandardOutput=file:awsl.log
StandardError=file:awsl.error.log

[Install]
WantedBy=multi-user.target
```
