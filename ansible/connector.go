package ansible

type Connector interface {
	//Connect(source *docker.Container, target *docker.Container) string, error // returned string is target host name
	Execute(executor Executor, playbook string) error
	//Disconnect() error
}
