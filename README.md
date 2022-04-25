# restreamer
一个方便用来管理推流转播的小工具。
需要环境内已经安装ffmpeg。
# 使用方法
## Docker
``` bash
~$ docker run -v ./restreamer/data:/root/data/ -p 13232:13232 --name restreamer ghcr.io/yuuinih/restreamer:latest
```
随后用浏览器访问127.0.0.1:13232即可。