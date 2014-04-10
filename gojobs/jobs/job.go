package jobs

//所有的job都必须实现这个接口
type Job interface {
    Start(arg string, result *string) error
    Stop(arg string, result *string) error
    Statu(arg string, result *string) error
}
