# log

Alternative log writer for Go.

## Summary

* Writter in pure Go
* Require Go version > 1.8
* Async write log to files
* Can write log to stdin and stderr
* Support formatted records
* Automatical log rotate
* Support log levels
* MIT license

## Install

```
go get github.com/wmentor/log
```

## Usage

### Open/Close log

```go
package main

import (
  "github.com/wmentor/log"
)

func main() {
  log.Open("name=app path=. period=day level=info keep=15")
  defer log.Close() // close log

  log.Info("log record")
}
```

*period* can take one of the following values: *minute*, *hour*, *day* or *month*. The *keep* parameter specifies the number of log files to save. Logs are saved in the *path* directory and has names (app-YYYYMMDD.log if period is day, app-YYYYMM.log if period is month and app-YYYYMMDDHH.log if period is hour).

When a new period comes, the old log file will be closed and a new one will be created.

You can use multiple log writers in your application. Add params *global=0* and *log.Open* return custom log object.

```go
package main

import (
  "github.com/wmentor/log"
)

func main() {
  log1, err1 := log.Open("name=app1 path=. period=day level=info keep=15 global=0")
  if err1 != nil {
    panic(err1)
  }
  defer log1.Close()

  log2, err2 := log.Open("name=app2 path=. period=day level=info save=15 global=0")
  if err2 != nil {
    panic(err2)
  }
  defer log2.Close()

  log1.Infof("log record %d %s", 1, "log1")
  log2.Errorf("log record %d %s", 1, "log2")
}
```

Add parameter *stderr=1* or *stdout=1* if you want print log message to *stderr* or *stdout*.

### Log levels

Log object support log levels (trace/debug/info/warn/error/fatal/none). Log writes nothing if selected level is *none*.
All *trace* and *debug* messages will be skipped if you select the logging level *info*.

You can set log level in *Open* (like *level=debug*).


On each level we have calls Trace/Debug/Info/Warn/Error/Fatal.

Moreover, *Fatal* writes message to log file and kills application.

### Formatted messages

Log object has functions Tracef/Debugf/Infof/Warnf/Errorf/Fatalf like Trace/Debug/Info/Warn/Error/Fatal, but they take format string and params list and print message to log file. The format options are same of fmt package.

```go
log.Infof("Hello, %s!", "Mike")
```

## Gin integration

```go
package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/wmentor/log"
)


func main() {
  
  gin.SetMode(gin.ReleaseMode)
  
  router := gin.New()

  log.Open("name= path=. stderr=1 period=day level=info")

  router.Use(log.GinLogger())
  router.Use(gin.Recovery())

  router.GET("/user/:name", func(c *gin.Context) {
    name := c.Param("name")
    c.String(http.StatusOK, "Hello %s", name)
  })

  router.Run(":8080")
}

```
