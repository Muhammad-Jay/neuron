package parser

import (
	"fmt"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/project"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

type Parser struct {
	data       *project.ResolvedProject
	varResolver *VariableResolver
}

func NewParser(data *project.ResolvedProject) Parser {
	return Parser{
		data:        data,
		varResolver: NewVariableResolver(&data.Project),
	}
}

func (p Parser) ParseSystem() (*core.System, error) {
	sysMeta := core.Metadata{
		ID:          core.NewID("system_"),
		Name:        p.data.System.Definition.Metadata.Name,
		Description: p.data.System.Definition.Metadata.Description,
		Version:     p.data.System.Definition.Metadata.Version,
	}

	services := make([]core.Service, 0, len(p.data.System.Services))
	serviceMap := make(map[string]core.Service)

	for _, rs := range p.data.System.Services {
		svc := p.convertService(rs)
		serviceMap[rs.Ref] = svc
		services = append(services, svc)
	}

	connectors := make([]core.Connector, 0, len(p.data.System.Connectors))
	incomingCount := make(map[core.ID]int)

	for _, rc := range p.data.System.Connectors {
		conn, err := p.convertConnector(rc, serviceMap)
		if err != nil {
			return nil, err
		}
		connectors = append(connectors, conn)
		incomingCount[conn.To.ServiceID]++
	}

	var triggers []core.Trigger
	var nonTriggerServices []core.Service

	if len(connectors) == 0 {
		for _, rs := range p.data.System.Services {
			triggers = append(triggers, core.Trigger{Service: serviceMap[rs.Ref]})
		}
	} else {
		for _, rs := range p.data.System.Services {
			svc := serviceMap[rs.Ref]
			if incomingCount[svc.Metadata.ID] == 0 {
				triggers = append(triggers, core.Trigger{Service: svc})
				continue
			}
			nonTriggerServices = append(nonTriggerServices, svc)
		}

		if len(triggers) == 0 && len(nonTriggerServices) > 0 {
			return nil, fmt.Errorf("system contains a connector cycle; no entry service found")
		}
	}

	return &core.System{
		Metadata: sysMeta,
		Specification: core.SystemSpec{
			Services:   nonTriggerServices,
			Triggers:   triggers,
			Connectors: connectors,
		},
	}, nil
}

func (p Parser) convertService(rs project.ResolvedService) core.Service {
	spec := rs.Definition.Spec

	var inputs, outputs []core.Port
	for _, m := range spec.Mappings {
		portType := core.ValueAny
		if strings.HasSuffix(m.Target, "_id") || strings.HasSuffix(m.Target, "_count") {
			portType = core.ValueNumber
		} else if strings.HasSuffix(m.Target, "_enabled") || strings.HasSuffix(m.Target, "_active") {
			portType = core.ValueBool
		}

		if m.Direction == "input" {
			inputs = append(inputs, core.Port{
				Name:     m.Target,
				Type:     portType,
				Required: true,
			})
		} else if m.Direction == "output" {
			outputs = append(outputs, core.Port{
				Name: m.Target,
				Type: portType,
			})
		}
	}

	rtConfig := core.RuntimeConfigurations{}
	if spec.Execution != nil {
		rtConfig.Timeout = spec.Execution.Timeout
		rtConfig.Retry = core.RetryPolicy{
			MaxAttempts: spec.Execution.Retries,
			Backoff:     "exponential",
		}
	}

	svcConfig := p.varResolver.Resolve(spec.Config).(map[string]any)

	return core.Service{
		Metadata: core.Metadata{
			ID:          core.ID(rs.Ref),
			Name:        rs.Definition.Metadata.Name,
			Description: rs.Definition.Metadata.Description,
			Version:     rs.Definition.Metadata.Version,
		},
		Type:                    core.ServiceType(spec.Executor.Type),
		ServiceConfigurations:   svcConfig,
		RuntimeConfigurations:   rtConfig,
		Inputs:                  inputs,
		Outputs:                 outputs,
	}
}

func (p Parser) convertConnector(rc project.ResolvedConnector, serviceMap map[string]core.Service) (core.Connector, error) {
	fromSvc, ok := serviceMap[rc.Definition.From]
	if !ok {
		return core.Connector{}, fmt.Errorf("connector %s: from service %q not found", rc.Ref, rc.Definition.From)
	}
	toSvc, ok := serviceMap[rc.Definition.To]
	if !ok {
		return core.Connector{}, fmt.Errorf("connector %s: to service %q not found", rc.Ref, rc.Definition.To)
	}

	var mappings []core.MappingRule
	for _, m := range rc.Definition.Mappings {
		mappings = append(mappings, core.MappingRule{
			TargetPath: m.Target,
			Expression: p.varResolver.ResolveString(m.Expression),
		})
	}

	var validations []core.ValidationRule
	for _, v := range rc.Definition.Validations {
		validations = append(validations, core.ValidationRule{
			Expression: p.varResolver.ResolveString(v.Expression),
			Message:    p.varResolver.ResolveString(v.Message),
		})
	}

	return core.Connector{
		Metadata: core.Metadata{
			ID:          core.NewID("connector_"),
			Name:        rc.Ref,
			Description: rc.Definition.Metadata.Description,
			Version:     rc.Definition.Metadata.Version,
		},
		From: core.Endpoint{
			ServiceID: fromSvc.Metadata.ID,
		},
		To: core.Endpoint{
			ServiceID: toSvc.Metadata.ID,
		},
		Mappings:    mappings,
		Validations: validations,
	}, nil
}

func (p Parser) GetInstanceKey() protocol.InstanceKey {
	hash, _ := protocol.HashBlueprint(p.data.System)
	return protocol.InstanceKey{
		SystemID: p.data.Project.Metadata.Name,
		Version:  p.data.Project.Metadata.Version,
		Hash:     hash,
		Env:      p.getEnvironment(),
	}
}

func (p Parser) getEnvironment() string {
	if p.data.Project.Runtime.Execution.DefaultMode != "" {
		return p.data.Project.Runtime.Execution.DefaultMode
	}
	return "development"
}

func (p Parser) GetRuntimeConfig() project.RuntimeConfig {
	return p.data.Project.Runtime
}

func (p Parser) GetStorageConfig() project.StorageConfig {
	return p.data.Project.Storage
}

func (p Parser) GetExecutorSources() []project.ExecutorSource {
	return p.data.Project.Executors.Sources
}

func (p Parser) GetInspectorConfig() project.InspectorConfig {
	return p.data.Project.Inspector
}