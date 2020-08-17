package mysql

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init()  {
	db, _ = sql.Open("mysql", "root:123456@tcp(127.0.0.1:13306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err" + err.Error())
		os.Exit(1)
	}
}

// 返回数据连接对象
func DBConn() *sql.DB {
	return db
}

