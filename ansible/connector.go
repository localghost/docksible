package ansible

import "github.com/localghost/docksible/docker"

type Connector interface {
	Execute(executor Executor, playbook string) error

	Connect(source *docker.Container, target *docker.Container) error
	Name() string
	Host() string
	Args() []string
	Disconnect() error
}
