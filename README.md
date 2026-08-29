# Neuron

Neuron is a runtime for building and operating complex software systems from composable, executable capabilities.

It is built around a simple idea:

> Software should be composed from things that can do something, connected by explicit relationships, and operated by a runtime that does not need to understand what those things are.

Neuron is not a workflow platform.

It is not tied to a particular kind of application, service, language, or execution model.

It is closer to a small operating environment, a micro-kernel-like foundation for systems whose capabilities can be composed, connected, and operated independently of the technology used to implement them.

---

## The Idea

Modern software is usually built as a collection of applications, services, workers, libraries, queues, databases, APIs, and infrastructure.

As systems grow, the difficulty is rarely writing one individual component.

The difficulty is making all of those components work together as one coherent system.

Neuron approaches this differently.

Instead of making the runtime understand every kind of application or service, it defines a small set of primitives:

- **System** — what a complete software system is made of
- **Service** — an executable capability
- **Connector** — how capabilities communicate
- **Executor** — how a capability is actually run
- **Instance** — a living realization of a System
- **N.O.R.E.** — the runtime environment in which Systems exist and operate

These concepts deliberately remain independent.

A service does not need to be a function.

A connector does not need to be an HTTP request.

An executor does not need to be written in the same language as the system using it.

And N.O.R.E. does not need to know what a service actually does.

It only needs to know how to operate it.

---

## A Different Model of Software

Neuron treats software as a composition of capabilities rather than a collection of applications.

A capability can be almost anything:

- a function
- a library
- an API
- a database operation
- a model
- a filesystem operation
- a browser automation task
- a GitHub operation
- a native program
- a WebAssembly module
- a process running on another machine
- a service written in another programming language
- or another system exposed as a capability

Neuron does not need to understand the implementation. It only needs a contract describing what the capability provides and how it can be reached. This makes the boundary between "application", "service", "worker", and "external system" much less important. They can all become **Services**.

---

# Core Concepts

## **System**

A **System** is the definition of a software system. It describes the capabilities that belong together and the relationships between them. A System can be small:

> receive a request → validate it → return a result

Or it can become extremely large:

> authentication → billing → inventory → notifications → analytics → external APIs → background processing

The important distinction is that a System describes **what exists and how it is connected**, rather than forcing everything to be implemented as one application. A System is therefore a composition.

```text
                 System
                    │
        ┌───────────┼───────────┐
        │           │           │
     Service     Service     Service
        │           │           │
        └─────── Connectors ────┘

```

Systems can themselves become building blocks for larger systems.

## **Service**

A Service is an executable capability.

Neuron intentionally does not restrict what a Service is.

For example:

* github.read
* github.commit
* customer.verify
* payment.authorize
* database.query
* image.generate
* model.predict
* filesystem.read
* email.send
* browser.open

The name does not determine how the capability is implemented.

One Service could run locally.
Another could run remotely.
Another could be a WebAssembly module.
Another could be implemented in Rust, Go, Python, JavaScript, or another language.
Another could simply be an interface to an external system.

From Neuron's perspective, they are all capabilities that can participate in a System.

## **Connector**

A Connector describes how one capability communicates with another.

This distinction is important.

A Service describes what can be done.
A Connector describes how it can be reached or connected.

Depending on the environment, a Connector could represent communication through:

* an in-process interface
* a local process
* a Unix socket
* a network connection
* an RPC protocol
* a message channel
* an external API
* another runtime
* or another supported transport

This allows the architecture of a System to remain independent from the transport used underneath it.

```text
Service A
   │
   │ Connector
   ▼
Service B

```

The same logical relationship can exist whether both services are on the same machine or separated across a network.

## **Executor**

An Executor is the mechanism that actually runs a Service.

Neuron separates the description of a capability from the mechanism used to execute it.

This allows different execution environments to coexist.

For example:

```text
Service
   │
   ▼
Executor
   │
   ├── Native process
   ├── WebAssembly
   ├── Local runtime
   ├── Remote runtime
   └── External service

```

The Service remains the logical capability.

The Executor provides the machinery required to make that capability operate.

This separation is what allows Neuron to support capabilities implemented using different technologies without turning the core runtime into a collection of special cases.

## **Instance**

A System is a definition.
An Instance is a living realization of that definition.

A System might describe:

> Customer Verification

An Instance represents an actual running or available realization of that system with its own state, resources, configuration, and activity.

This distinction allows the same System definition to have many independent instances.

```text
                 System
          Customer Verification
                   │
          ┌────────┼────────┐
          │        │        │
       Instance  Instance  Instance
          A        B        C

```

Each instance can therefore exist independently while being derived from the same underlying definition.

This is one of the foundations that allows Neuron to move beyond simple task execution.

## **N.O.R.E.**

**Neuron Operating Runtime Environment**

N.O.R.E. is the runtime environment of Neuron.

It is where Systems are registered, instantiated, operated, and connected to the capabilities they require.

N.O.R.E. provides the environment in which a System can become something that actually exists.

Conceptually:

```text
                    Neuron
                      │
                      ▼
                   N.O.R.E.
                      │
          ┌───────────┼───────────┐
          │           │           │
       Systems     Instances    Services
          │           │           │
          └───────────┼───────────┘
                      │
                 Executors
                      │
          ┌───────────┼───────────┐
          │           │           │
       Local       WASM        Remote

```

N.O.R.E. is intentionally not responsible for understanding the business meaning of the capabilities it operates.

It provides the runtime primitives.
The capabilities provide the behavior.

## The Runtime as a Small Operating Environment

The operating-system analogy is useful, but with an important distinction.

Neuron is not trying to become an operating system for hardware.
It is an operating environment for software capabilities.

An operating system provides primitives such as:
processes, memory, resources, communication, isolation, scheduling, identity, and persistence.

Neuron applies a similar philosophy at a higher level.
It provides a foundation around:
Systems, Instances, Services, Executors, Connectors, resources, state, communication, and execution.

The goal is to make complex software composable in the same way that an operating system makes complex programs possible.

This is why N.O.R.E. can be thought of as a micro-kernel-like runtime for Neuron.
The kernel should remain small.
Capabilities should live outside it.

## Everything Is a Capability

One of the most important ideas in Neuron is that a Service does not have to correspond to a traditional "microservice".

A Service can represent a capability at any level.

For example:

```text
                 System
                    │
        ┌───────────┼───────────┐
        │           │           │
    Database      Model       GitHub
     Service     Service      Service
        │           │           │
        └───────────┼───────────┘
                    │
                Application

```

Or:

```text
                 System
                    │
             ┌──────┴──────┐
             │             │
          Service        Service
             │             │
          WASM          Remote API

```

Or even:

```text
System
  │
  └── Another System
          │
          ├── Service
          ├── Service
          └── Service

```

This creates a recursive model of software composition.

Complex systems can be constructed from smaller systems without requiring the runtime to treat them as fundamentally different things.

## Composition

Neuron is designed around composition rather than a single prescribed architecture.

A System can combine capabilities implemented in completely different environments.

For example:

```text
                    Order System
                         │
        ┌────────────────┼────────────────┐
        │                │                │
     Validate         Payment         Inventory
      Order            │                │
        │              │                │
      Local           Remote           WASM
     Service          Service          Service

```

The System does not need to care that these capabilities are implemented differently.

The runtime maintains the boundaries.
The connectors maintain communication.
The executors provide execution.
The System provides composition.

## Local and Distributed

Neuron is designed so that location does not have to define the architecture.

A capability may exist:

```text
Same process
     ↓
Same machine
     ↓
Another local process
     ↓
Another machine
     ↓
Another runtime
     ↓
Another environment

```

The logical System can remain the same while its physical deployment changes.

This makes it possible to begin with a completely local system and gradually distribute individual capabilities as the system grows.

## State and Identity

A System defines behavior.
An Instance provides a place where that behavior can exist.

This separation allows state to belong to the appropriate level.

For example, a System can define:

> Customer Verification

while an Instance can represent:

```text
Customer Verification
│
├── configuration
├── resources
├── current state
└── activity

```

This makes it possible for the same definition to be reused without forcing every realization to share the same state.

## Execution Without a Single Execution Model

Neuron is not built around the assumption that every operation should become a background job.

Some capabilities may be:
request/response, long-running, interactive, streaming, asynchronous, scheduled, event-driven, or continuously active.

The runtime therefore separates the concept of a capability from the way that capability is operated.

A request can produce an immediate result.
Another operation can continue independently.
Another can expose intermediate state while it is active.

The model is intentionally broader than traditional workflow execution.

## Why Neuron Is Not a Workflow Engine

Workflow engines generally begin with a predefined abstraction:

> A workflow consists of a sequence of tasks.

Neuron starts somewhere else:

> A System consists of capabilities and relationships.

That difference matters.

A workflow is one possible thing that can be represented using Neuron.
It is not the boundary of the platform.

A System may represent:
an API backend, an automation system, an AI application, a data-processing environment, an interactive application, a distributed service, an internal platform, a long-running process, a collection of external capabilities, or something that does not fit neatly into traditional application categories.

Neuron is intended to provide the underlying runtime model rather than dictate the application model.

## The Architecture in One Picture

```text
                         SYSTEM
                            │
             ┌──────────────┼──────────────┐
             │              │              │
          SERVICE         SERVICE        SERVICE
             │              │              │
             └───────┐      │      ┌───────┘
                     │      │      │
                  CONNECTORS / RELATIONSHIPS
                            │
                            ▼
                         INSTANCE
                            │
                            ▼
                          N.O.R.E.
                            │
             ┌──────────────┼──────────────┐
             │              │              │
          EXECUTOR       EXECUTOR       EXECUTOR
             │              │              │
          Native           WASM          Remote
             │              │              │
             └──────────────┴──────────────┘
                            │
                            ▼
                       Real Software

```

The important boundary is simple:

* System defines composition.
* Service defines capability.
* Connector defines relationships.
* Executor provides execution.
* Instance provides a living realization.
* N.O.R.E. provides the environment in which those pieces operate.

## Built for Heterogeneous Software

Neuron does not require an ecosystem where everything is written in the same language or deployed in the same way.

A System can combine capabilities written in different languages and running in different environments.

This is particularly important as software becomes increasingly heterogeneous.

A capability might be implemented using:
Go, Rust, C/C++, Python, JavaScript, WebAssembly, Remote services, or External APIs.

The implementation language should not become part of the fundamental System model.

Neuron cares about the capability and its contract.

## WebAssembly

WebAssembly provides one possible execution environment for Services.

Its value in Neuron is not simply that it is "fast".
It provides a portable and controlled execution boundary.

That makes it useful when a Service needs to be:
portable, isolated, distributed as a single artifact, executable across different environments, implemented independently from the main runtime, or loaded dynamically.

WebAssembly therefore fits naturally at the Executor boundary.

```text
Service
   │
   ▼
WASM Executor
   │
   ▼
WebAssembly Module

```

WASM is one execution mechanism among many, rather than something every Service must become.

## From Definition to Runtime

The conceptual lifecycle of a Neuron System is:

```text
Definition
    │
    ▼
Resolution
    │
    ▼
System
    │
    ▼
Registration
    │
    ▼
N.O.R.E.
    │
    ▼
Instance
    │
    ▼
Capabilities
    │
    ▼
Execution

```

The important part is that a System definition is not the same thing as a running Instance.

A definition describes what should exist.
N.O.R.E. provides the environment.
An Instance makes that definition operational.
Services provide the actual capabilities.

## Designed for Growth

Neuron is intended to work at different scales without changing its fundamental model.

A small system might look like:

```text
Application
    │
    ├── Service
    └── Service

```

A larger system might look like:

```text
Application
    │
    ├── System
    │    ├── Service
    │    └── Service
    │
    ├── System
    │    ├── Service
    │    └── Service
    │
    └── External Service

```

The same primitives remain useful.
This is deliberate.

The goal is not to create a different abstraction for every scale of software.
The goal is to have a small number of abstractions that remain useful as the system grows.

## The Philosophy

Neuron is based on a few principles.

**Capabilities over applications**
A useful capability should not have to become an entire application before it can participate in a larger system.

**Composition over coupling**
Systems should be assembled from independent capabilities rather than tightly coupled implementations.

**Contracts over implementation details**
The runtime should depend on what a capability provides, not how it was implemented.

**Runtime over framework**
Neuron provides an environment in which software can operate rather than forcing every application into one programming model.

**Local first, distributed when necessary**
A system should be able to start locally without requiring infrastructure that it does not need.
As its requirements grow, individual capabilities can move to other processes, machines, or environments.

**Small primitives, large systems**
The core model should remain understandable even when the systems built on top of it become complex.

## What Neuron Is Becoming

Neuron is an attempt to create a different foundation for software composition.

Instead of asking:
"What kind of application are you building?"

Neuron asks:
"What capabilities does your system have, how are they connected, and where should they operate?"

That distinction opens the door to systems that are not constrained by traditional application boundaries.

A database operation can be a Service.
A machine-learning model can be a Service.
A remote API can be a Service.
A WebAssembly module can be a Service.
An entire System can become a Service.

And all of them can participate in a larger System through explicit connections.

N.O.R.E. provides the environment that makes this possible.

## Neuron

Compose capabilities.
Connect systems.
Operate software as a living environment.

```
