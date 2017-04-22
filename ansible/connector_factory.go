package ansible

import "fmt"

func CreateConnector(connectorType string) Connector {
	switch connectorType {
	case "docker-exec":
		return NewDockerConnector()
	case "ssh":
		return NewSshConnector()
	}
	panic(fmt.Sprintf("Connector %s is not supported.", connectorType))
}
