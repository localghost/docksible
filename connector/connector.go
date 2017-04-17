package connector

type Connector interface {
	Execute(executor Executor, playbook string) error
}
