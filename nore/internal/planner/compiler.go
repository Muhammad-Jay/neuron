package planner

import (
	"fmt"
	"strings"

	"github.com/Muhammad-Jay/neuron/nore/internal/resolver"
	"github.com/Muhammad-Jay/neuron/nore/internal/types"
	shared "github.com/Muhammad-Jay/neuron/shared/types/core"
)

type Compiler struct {
	expressions resolver.Compiler
}

func NewCompiler(expressions resolver.Compiler) (*Compiler, error) {
	if expressions == nil {
		return nil, fmt.Errorf("expression compiler is required")
	}
	return &Compiler{expressions: expressions}, nil
}

func (c *Compiler) Compile(system shared.System) (*types.ExecutionBlueprint, error) {
	services := make(map[shared.ID]shared.Service)
	triggerIDs := make([]shared.ID, 0, len(system.Specification.Triggers))

	for _, trigger := range system.Specification.Triggers {
		service := cloneService(trigger.Service)
		if err := addService(services, service); err != nil {
			return nil, err
		}
		triggerIDs = append(triggerIDs, service.Metadata.ID)
	}
	for _, service := range system.Specification.Services {
		if err := addService(services, cloneService(service)); err != nil {
			return nil, err
		}
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("systems must contain at least one service")
	}

	nodes := make(map[shared.ID]types.ExecutionNode, len(services))
	incoming := make(map[shared.ID]int, len(services))
	for id, service := range services {
		compiledConfig, err := c.expressions.CompileServiceConfigurations(service.ServiceConfigurations)
		if err != nil {
			return nil, fmt.Errorf("compile configurations for service %s: %w", id, err)
		}
		nodes[id] = types.ExecutionNode{Service: service, Configurations: compiledConfig}
		incoming[id] = 0
	}

	for _, original := range system.Specification.Connectors {
		connector := cloneConnector(original)
		fromService, fromExists := services[connector.From.ServiceID]
		if !fromExists {
			return nil, fmt.Errorf("connector %s references missing source service %s", connector.Metadata.ID, connector.From.ServiceID)
		}
		toService, toExists := services[connector.To.ServiceID]
		if !toExists {
			return nil, fmt.Errorf("connector %s references missing target service %s", connector.Metadata.ID, connector.To.ServiceID)
		}
		if connector.From.Port != "" && !hasPort(fromService.Outputs, connector.From.Port) {
			return nil, fmt.Errorf("source output port %q does not exist on service %s", connector.From.Port, fromService.Metadata.ID)
		}
		if connector.To.Port != "" && !hasPort(toService.Inputs, connector.To.Port) {
			return nil, fmt.Errorf("target input port %q does not exist on service %s", connector.To.Port, toService.Metadata.ID)
		}

		incoming[connector.To.ServiceID]++
		if incoming[connector.To.ServiceID] > 1 {
			return nil, fmt.Errorf("service %s has multiple incoming connectors; use an aggregation service", connector.To.ServiceID)
		}

		transition, err := c.compileTransition(connector)
		if err != nil {
			return nil, err
		}
		node := nodes[connector.From.ServiceID]
		node.Next = append(node.Next, transition)
		nodes[connector.From.ServiceID] = node
	}

	entryIDs := triggerIDs
	if len(entryIDs) == 0 {
		for serviceID, count := range incoming {
			if count == 0 {
				entryIDs = append(entryIDs, serviceID)
			}
		}
	}
	if len(entryIDs) == 0 {
		return nil, fmt.Errorf("systems has no entry service")
	}
	if err := validateAcyclic(nodes, incoming); err != nil {
		return nil, err
	}
	if err := validateReachability(nodes, entryIDs); err != nil {
		return nil, err
	}

	metadata := cloneMetadata(system.Metadata)
	if metadata.ID == "" {
		metadata.ID = shared.NewID("blueprint_")
	}
	if metadata.Version == "" {
		metadata.Version = "1"
	}
	return &types.ExecutionBlueprint{Metadata: metadata, Nodes: nodes, EntryServiceIDs: entryIDs}, nil
}

func (c *Compiler) compileTransition(connector shared.Connector) (types.ExecutionTransition, error) {
	compiledMappings := make([]types.CompiledMapping, 0, len(connector.Mappings))
	targets := make(map[string]struct{}, len(connector.Mappings))
	for index, mapping := range connector.Mappings {
		targetPath := strings.TrimSpace(mapping.TargetPath)
		expression := strings.TrimSpace(mapping.Expression)
		if targetPath == "" {
			return types.ExecutionTransition{}, fmt.Errorf("connector %s mapping %d has no target path", connector.Metadata.ID, index)
		}
		if expression == "" {
			return types.ExecutionTransition{}, fmt.Errorf("connector %s mapping %q has no expression", connector.Metadata.ID, targetPath)
		}
		if _, exists := targets[targetPath]; exists {
			return types.ExecutionTransition{}, fmt.Errorf("connector %s contains duplicate target path %q", connector.Metadata.ID, targetPath)
		}
		targets[targetPath] = struct{}{}
		program, err := c.expressions.CompileTransitionExpression(expression)
		if err != nil {
			return types.ExecutionTransition{}, fmt.Errorf("connector %s mapping %q: %w", connector.Metadata.ID, targetPath, err)
		}
		compiledMappings = append(compiledMappings, types.CompiledMapping{TargetPath: targetPath, Expression: expression, Program: program})
	}

	compiledValidations := make([]types.CompiledValidation, 0, len(connector.Validations))
	for index, rule := range connector.Validations {
		expression := strings.TrimSpace(rule.Expression)
		if expression == "" {
			return types.ExecutionTransition{}, fmt.Errorf("connector %s validation %d has no expression", connector.Metadata.ID, index)
		}
		program, err := c.expressions.CompileTransitionExpression(expression)
		if err != nil {
			return types.ExecutionTransition{}, fmt.Errorf("connector %s validation %d: %w", connector.Metadata.ID, index, err)
		}
		message := strings.TrimSpace(rule.Message)
		if message == "" {
			message = "connector validation failed"
		}
		compiledValidations = append(compiledValidations, types.CompiledValidation{Expression: expression, Message: message, Program: program})
	}

	return types.ExecutionTransition{
		ConnectorID: connector.Metadata.ID, TargetServiceID: connector.To.ServiceID,
		Mappings: compiledMappings, Validations: compiledValidations,
	}, nil
}

func addService(services map[shared.ID]shared.Service, service shared.Service) error {
	if service.Metadata.ID == "" {
		return fmt.Errorf("service ID is required")
	}
	if service.Type == "" {
		return fmt.Errorf("service %s has no service type", service.Metadata.ID)
	}
	if _, exists := services[service.Metadata.ID]; exists {
		return fmt.Errorf("duplicate service ID %s", service.Metadata.ID)
	}
	services[service.Metadata.ID] = service
	return nil
}

func hasPort(ports []shared.Port, name string) bool {
	for _, port := range ports {
		if port.Name == name {
			return true
		}
	}
	return false
}

func validateAcyclic(nodes map[shared.ID]types.ExecutionNode, incoming map[shared.ID]int) error {
	counts := make(map[shared.ID]int, len(incoming))
	queue := make([]shared.ID, 0, len(nodes))
	for id, count := range incoming {
		counts[id] = count
		if count == 0 {
			queue = append(queue, id)
		}
	}
	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		visited++
		for _, transition := range nodes[current].Next {
			target := transition.TargetServiceID
			counts[target]--
			if counts[target] == 0 {
				queue = append(queue, target)
			}
		}
	}
	if visited != len(nodes) {
		return fmt.Errorf("systems contains a connector cycle")
	}
	return nil
}

func validateReachability(nodes map[shared.ID]types.ExecutionNode, entries []shared.ID) error {
	visited := make(map[shared.ID]bool, len(nodes))
	queue := append([]shared.ID(nil), entries...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		for _, transition := range nodes[current].Next {
			queue = append(queue, transition.TargetServiceID)
		}
	}
	if len(visited) != len(nodes) {
		return fmt.Errorf("systems contains unreachable services")
	}
	return nil
}

func cloneMetadata(metadata shared.Metadata) shared.Metadata {
	labels := make(map[string]string, len(metadata.Labels))
	for key, value := range metadata.Labels {
		labels[key] = value
	}
	metadata.Labels = labels
	return metadata
}

func cloneService(service shared.Service) shared.Service {
	service.Metadata = cloneMetadata(service.Metadata)
	service.ServiceConfigurations = cloneConfiguration(service.ServiceConfigurations)
	service.Inputs = append([]shared.Port(nil), service.Inputs...)
	service.Outputs = append([]shared.Port(nil), service.Outputs...)
	return service
}

func cloneConnector(connector shared.Connector) shared.Connector {
	connector.Metadata = cloneMetadata(connector.Metadata)
	if connector.Metadata.ID == "" {
		connector.Metadata.ID = shared.NewID("connector_")
	}
	connector.Mappings = append([]shared.MappingRule(nil), connector.Mappings...)
	connector.Validations = append([]shared.ValidationRule(nil), connector.Validations...)
	return connector
}

func cloneConfiguration(source shared.ServiceConfigurations) shared.ServiceConfigurations {
	result := make(shared.ServiceConfigurations, len(source))
	for key, value := range source {
		result[key] = cloneConfigurationValue(value)
	}
	return result
}

func cloneConfigurationValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = cloneConfigurationValue(item)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = cloneConfigurationValue(item)
		}
		return result
	default:
		return typed
	}
}
