package ansible

type Executor interface {
	Execute(command []string) error
}

type ExecutorFunc func(command []string) error

func (f ExecutorFunc) Execute(command []string) error {
	return f(command)
}
