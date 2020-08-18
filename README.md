## 3.0划重点

​ 关于数据库访问，Golang 中提供了标准库 database/sql。不过它不是针对某种具体数据库的逻辑实现，而是一套统一抽象的接口。
真正与数据库打交道的，是各个数据库对应的驱动 Driver；在使用时需要先注册对应的驱动库，然后就能通过标准库 sql 中定义的接口来统一操作数据库。

- 创建 sql.DB 连接池

​ 我们来看一下如何创建 sdl.DB 连接池，以 MySQL 为例:

```go
import (
  "log"
  "os"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)
​
func main() {
  db, err := sql.Open("mysql",
    "user:pwd@tcp(127.0.0.1:3306)/testdb")
  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }
  defer db.Close()

  err = db.Ping()
  if err != nil {
     // TODO do something
  }
  // ...
}
```

创建数据库连接是一种比较耗资源的操作，先要完成 TCP 三次握手，连接到数据库后需要进行鉴权和分配连接资源，因此建议使用长连接来避免频繁进行此类操作。在我们的 Go 应用中，sql.DB 自身就是会管理连接池的，一般实现为全局的连接池，不用重复进行 open/close 动作。

- CRUD 接口

```go
// 1. 返回多行数据，手动关闭结果集 (defer rows.Close())
db.Query()
// 2. 返回单行数据，不须手动关闭结果集
db.QueryRow()
// 3. 预先将一条连接(conn)与一条sql语句绑定起来，供重复使用
stmt, err := db.Prepare(sql)
stmt.Query(args)
// 4. 适用于执行增/删/更新等操作(不需要返回结果集)
db.Exec()
```

- 结果集合

Query()方法会返回结果集合，需要通过 rows.Next(), rows.Scan()方法来遍历结果集合。

- 事务

```go
// 开始事务
tx := db.Begin()
// 执行事务
db.Exec()
// ...
// 提交事务
tx.Commit()
// 如果失败的话，回滚
tx.Rollback()
```

- 小结

  用户层面所执行的 sql 语句，在底层其实是将 sql 语句串编码后传输到数据库服务器端再执行。因此本质上可以看作是 C/S 架构程序。

  对于如何访问不同的数据库，Go 的做法是抽离出具体代码逻辑，将数据库操作分为 database/sql 和 driver 两层。database/sql 负责提供统一的用户接口，以及一些不涉及具体数据库的逻辑(如连接池管理)；Driver 层负责实际的数据库通信。

## 4.0划重点

- 理解和使用中间件

一个好的中间件是能够根据需求实现可插拔的。这就意味着可以在接口级别嵌入一个中间件，他就能直接正常的工作。它不会污染原有的代码，不会改变原始的逻辑，也不会影响原有的编码方式。需要它的时候就把它安插在指定的位置，不需要的时候就把它移除掉。

在 Go 语言中，在标准库中，中间件的应用是非常普遍的。一个比较典型的场景，在 net/http 里的某些业务 handler，可以配置 TimeoutHandler 作为中间件，用来处理请求超时的问题；这样业务 handler 专注于处理业务逻辑即可，无需亲自管理超时。

同样，在 API 鉴权方面，中间件也是发挥了不可或缺的作用：API 请求首先经过配置好的中间件，只有校验通过了才会将请求转发到下一级中间件或具体的业务 handler 进行逻辑的处理。以下是 API 拦截器使用示例:

```go
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            r.ParseForm()
            // TODO: 进行身份验证，比如校验cookie或token
            h(w, r)
        })
}

// 然后在创建路由时类似这么调用:
// http.HandleFunc("/test", HTTPInterceptor(YourCustomHandler))
```

## 关于 go modules

在 go1.11 之后，官方推出了 Go Modules 这原生的包管理工具；相对于 vendor, go mod 可以更有效的进行包版本管理。

- 在 1.12 及之前， go modules 需要配置环境变量来开启此功能:

```bash
# 分别有 on off auto
go env -w GO111MODULE=on
```

- 配置代理。因为众所周知的原因，有些包我们国内无法访问，一般需要通过代理(如 goproxy.cn):

```bash
go env -w GOSUMDB=sum.golang.google.cn
go env -w GOPROXY=https://goproxy.cn,direct
# 查看是否成功(go env的输出中包含代理信息)
go env
```

- go mod 初始化

```
go mod init <指定一个module名，如工程名>
```

在项目根目录下成功初始化后，会生成一个 go.mod 和 go.sum 文件。在之后执行 go build 或 go run 时，会触发依赖解析，并且下载对应的依赖包。

更具体的用法可以参考网上其他教程哦(如https://github.com/golang/go/wiki/Modules)。



