package jobs

type Job interface {
    Start(arg interface{}, result *string) error
    Stop(arg interface{}, result *string) error
    Run()
}
