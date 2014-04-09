package jobs

type Job interface {
    Start(arg string, result *string) error
    Stop(arg string, result *string) error
    Statu(arg string, result *string) error
}
