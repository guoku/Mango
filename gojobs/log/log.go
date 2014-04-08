package log

import (
    "database/sql"
    "fmt"
    "io"
    "os"
    "runtime"
    "strings"
    "sync"
    "time"

    "github.com/astaxie/beego"
    _ "github.com/go-sql-driver/mysql"
)

const (
    Linfo = iota
    Lwarn
    Lerror
)
const (
    MYSQL = "mysql"
)

var Level int = Linfo
var Output io.Writer
var Std *Logger
var levels = []string{
    "INFO",
    "WARN",
    "ERROR",
}

func init() {
    lvl := beego.AppConfig.String("log::loglevel")
    if lvl == "warn" {
        Level = Lwarn
    }
    if lvl == "error" {
        Level = Lerror
    }
    op := beego.AppConfig.String("log::output")
    if op == MYSQL {
        mysqluser := beego.AppConfig.String("log::mysqluser")
        mysqlpass := beego.AppConfig.String("log::mysqlpass")
        mysqlurl := beego.AppConfig.String("log::mysqlurl")
        mysqlport := beego.AppConfig.String("log::mysqlport")
        mysqldb := beego.AppConfig.String("log::mysqldb")
        mysqltable := beego.AppConfig.String("log::mysqltable")
        db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", mysqluser, mysqlpass, mysqlurl, mysqlport, mysqldb))
        if err != nil {
            panic(err)
        }
        err = db.Ping()
        if err != nil {
            panic(err)
        }
        stmt, err := db.Prepare(fmt.Sprintf("insert into %s(level, logtype, file, line,time, reason) values (?, ?, ?, ?, ?, ?)", mysqltable))
        if err != nil {
            panic(err)
        }
        Output = MysqlWriter{User: mysqluser, Password: mysqlpass, Url: mysqlurl, Port: mysqlport, DB: mysqldb, Table: mysqltable, stmt: stmt}
        Std = New(Output, Level)
    } else {
        Std = New(os.Stderr, Level)
    }
}

type MysqlWriter struct {
    User     string
    Password string
    Url      string
    Port     string
    DB       string
    Table    string
    stmt     *sql.Stmt
}

func (W MysqlWriter) Write(p []byte) (n int, err error) {
    s := string(p)
    sarray := strings.Split(s, "\t")
    lvl := sarray[0]
    logtype := sarray[1]
    file := sarray[2]
    line := sarray[3]
    now := sarray[4]
    reason := sarray[5]
    _, err = W.stmt.Exec(lvl, logtype, file, line, now, reason)
    if err != nil {
        fmt.Println(err, " log.go")
    }
    return 1, err
}

type Logger struct {
    mu    sync.Mutex
    Level int
    Out   io.Writer
}

func New(out io.Writer, level int) *Logger {
    return &Logger{Out: out, Level: level}
}

func (l *Logger) PrintOut(level int, logtype string, s string) error {
    if level < l.Level {
        return nil
    }
    now := time.Now()
    var file string
    var line int

    _, file, line, _ = runtime.Caller(2)
    l.mu.Lock()
    defer l.mu.Unlock()
    nowFormat := now.Format("2006-01-02 15:04:05")
    outStr := fmt.Sprintf("%s\t%s\t%s\t%d\t%s\t%s\n", levels[level], logtype, file, line, nowFormat, s)
    _, err := l.Out.Write([]byte(outStr))
    return err
}
func Info(v ...interface{}) {
    Std.PrintOut(Linfo, "", fmt.Sprintln(v...))
}

func Error(v ...interface{}) {
    Std.PrintOut(Lerror, "", fmt.Sprintln(v...))
}

func Warn(v ...interface{}) {
    Std.PrintOut(Lwarn, "", fmt.Sprintln(v...))
}
func Infof(format string, v ...interface{}) {
    Std.PrintOut(Linfo, "", fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
    Std.PrintOut(Lerror, "", fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
    Std.PrintOut(Lwarn, "", fmt.Sprintf(format, v...))
}

//通过logtype，可以设置不同的log类型，比如http类型，文本处理异常类型
//提供这些类型的作用，是便于筛选和分析log数据
func InfofType(logtype string, format string, v ...interface{}) {
    Std.PrintOut(Linfo, logtype, fmt.Sprintf(format, v...))
}

func ErrorfType(logtype string, format string, v ...interface{}) {
    Std.PrintOut(Lerror, logtype, fmt.Sprintf(format, v...))
}

func WarnfType(logtype string, format string, v ...interface{}) {
    Std.PrintOut(Lerror, logtype, fmt.Sprintf(format, v...))
}
