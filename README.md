# 简易命令行秒表

大概长这样子

```txt
13s (13s)
56s (43s)
01m:05s (10s)
====
01m:06s (01s)^C
2020/12/14 14:00:11 - 2020/12/14 14:01:18
```

按回车键记下一条记录

ctl-c 暂停，之后再 ctl-c 退出，或回车继续

应该不支持 Windows （主要是清屏的命令跟 *nix 不一样）
