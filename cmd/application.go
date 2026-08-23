package main
//
//import (
//	"context"
//	customer_system "development/systems/customer-systems"
//	"log"
//	"nore/analytics"
//	"nore/resolver"
//	"sync"
//
//	"api/server"
//	"nore/engine"
//	"nore/event"
//	"nore/planner"
//	"nore/registry"
//	runtimemodel "nore/runtime"
//	"nore/scheduler"
//
//	"log/slog"
//)
//
//func main() {
//	// Global context that lives as long as the application is Running
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	celCompiler, err := resolver.NewCELCompiler(resolver.DefaultCELConfig())
//	must(err)
//	systemCompiler, err := planner.NewCompiler(celCompiler)
//	must(err)
//	blueprint, err := systemCompiler.Compile(*customer_system.Sys.Build())
//	must(err)
//
//	bus := event.NewBus()
//	defer bus.Close()
//	store := runtimemodel.NewMemoryStore()
//	reg := registry.New()
//
//	reg.RegisterCoreServiceExecutors()
//
//	an, _ := analytics.New(bus, slog.Default())
//	sched, err := scheduler.New(bus, store)
//
//	executorEngine, err := engine.NewExecutorEngine(bus, reg, store, 8)
//	must(err)
//
//	go func() {
//		if err := sched.Run(ctx); err != nil {
//			log.Printf("Scheduler stopped: %v\n", err)
//		}
//	}()
//
//	go func() {
//		if err := an.Serve(ctx); err != nil {
//			log.Printf("Analytics stopped: %v\n", err)
//		}
//	}()
//
//	go func() {
//		if err := executorEngine.Run(ctx); err != nil {
//			log.Printf("Engine stopped: %v\n", err)
//		}
//	}()
//
//	activeRequests := &sync.Map{}
//	srv := server.NewServer(bus, store, *blueprint, activeRequests)
//
//	// Start the event router to listen for completions in the background
//	srv.StartEventRouter(ctx)
//
//	if err := srv.Start(":8080"); err != nil {
//		log.Fatalf("Server crashed: %v\n", err)
//	}
//}
//
//func must(err error) {
//	if err != nil {
//		panic(err)
//	}
//}