package simple_response

import (
	"fmt"
	"simple_response/configs"
	"simple_response/services"

	"github.com/Muhammad-Jay/neuron/application/sdk"
)

var System = sdk.New("Simple Response", "1.0.0")

func init() {
	System.AddService(services.Service1)
	System.AddService(services.Service2)

	// Connector 1
	Connector1 := System.Connector(services.Service1, services.Service2)
	Connector1.Metadata("http-to-ai", "HTTP Response → AI Analysis")
	Connector1.AddMappings(configs.Mappings.CommandToLog...)

	// Connector 2
	Connector2 := System.Connector(services.Service1, services.Service2)
	Connector2.Metadata("http-to-llm", "HTTP Response → AI Analysis")
	Connector2.AddMappings(configs.Mappings.AIToCommand...)

	System.AddConnector(Connector1.Core())
	System.AddConnector(Connector2.Core())
}

func main() {
	fmt.Println(System.Build())
}
