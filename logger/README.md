# logger
convenient log package

# 1. 使用说明
```go
    // 配置logger，如果不配置时默认为控制台输出，等级为DEBUG
    logger.SetLogger(`{"Console": {"level": "DEBG"}`)
    // 配置说明见下文

    // 设置完成后，即可在控制台和日志文件app.log中看到如下输出
    logger.Debug("this is Debug")
    logger.Trace("this is Trace")
    logger.Alert("this is Alert")
    logger.Error("this is Error")
    logger.Panic("this is Panic")
    logger.Fatal("this is Fatal")
```

# 2. 日志等级

当前日志输出等级共6种，对应的等级由底到高，当配置为某个输出等级时，只有大于等于该等级的日志才会输出
