package main
//
//import (
//	"context"
//	"fmt"
//	"nore/types"
//	"nore/engine"
//	"nore/event"
//	"nore/executors"
//	"nore/planner"
//	"nore/registry"
//	"nore/resolver"
//	runtimemodel "nore/runtime"
//	"nore/scheduler"
//	"sync"
//	"time"
//)
//
//func cmd() {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	celCompiler, err := resolver.NewCELCompiler(resolver.DefaultCELConfig())
//	must(err)
//	systemCompiler, err := planner.NewCompiler(celCompiler)
//	must(err)
//	blueprint, err := systemCompiler.Compile(buildSystem())
//	must(err)
//
//	bus := event.NewBus()
//	defer bus.Close()
//	store := runtimemodel.NewMemoryStore()
//	reg := registry.New()
//	must(reg.Register("set", executors.SetExecutor{}))
//	must(reg.Register("ai", executors.AIMockExecutor{}))
//	must(reg.Register("log", executors.LogExecutor{}))
//
//	sched, err := scheduler.New(bus, store)
//	must(err)
//	execEngine, err := engine.NewExecutorEngine(bus, reg, store, 8)
//	must(err)
//
//	completed, err := bus.Subscribe(event.ExecutionCompleted, 8)
//	must(err)
//	defer completed.Close()
//	failed, err := bus.Subscribe(event.ExecutionFailed, 8)
//	must(err)
//	defer failed.Close()
//
//	var group sync.WaitGroup
//	group.Add(2)
//	go func() { defer group.Done(); _ = sched.Run(ctx) }()
//	go func() { defer group.Done(); _ = execEngine.Run(ctx) }()
//
//	execution, err := runtimemodel.NewExecution(blueprint, types.NewID("req_"))
//	must(err)
//	must(store.Add(execution))
//	must(bus.Publish(ctx, event.New(event.ExecutionStarted, execution.ID, execution.CorrelationID, "", event.ExecutionStartedPayload{
//		Input: map[string]any{"name": "Amina", "email": "amina@example.com", "vip": true},
//	})))
//
//	select {
//	case <-completed.Events():
//		fmt.Println("Execution completed")
//	case e := <-failed.Events():
//		fmt.Println("Execution failed:", e.Payload)
//	case <-ctx.Done():
//		fmt.Println("Timed out")
//	}
//	cancel()
//	group.Wait()
//}