package ansible

import "github.com/localghost/docksible/docker"

type Connector interface {
	Connect(source *docker.Container, target *docker.Container) error
	Name() string
	Host() string
	ExtraArgs() []string
	Disconnect() error
}
