package command

var (
	Neuron = "neuron"

	Run = "run"

	Init = "init [Target]"

	Instance = "instance"
	InstanceList = "list [instance-id]"
	InstanceRemove = "remove [instance-id|system-key]"
	InstanceClear = "clear"

	Register = "register"

	Daemon = "daemon"

	Executor      = "executor"
	ExecutorList    = "list"
	ExecutorInspect = "inspect [name@version]"

	Add = "add [name@version]"
	Remove = "remove [name@version]"
)

