
# slowsql-analysis
基于pt-query-digest工具的慢SQL分析工具

# 使用方式
目录cmd template，以及编译的slowsql-analysis这个文件放在一个目录下

执行


```js
./slowsql-analysis -h 可以获取帮助信息
Usage of ./slowsql-analysis:
-endTime string
结束时间 格式：yyyy-mm-dd hh:mm:ss
-f string
慢sq日志文件所在的位置 例子：/var/log/mysql4306-slow.log
-startTime string
开始时间 格式：yyyy-mm-dd hh:mm:ss
```

执行例子：



```js
./slowsql-analysis -f /var/log/mysql4306-slow.log -startTime="2024-04-16 00:00:00" -endTime="2024-04-16 23:00:00"
```

最终成一个html文件，直接打开即可
