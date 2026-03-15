package domain

// Policy represents an implicit reactive rule: WHEN [EVENT] THEN [COMMAND].
type Policy struct {
	Name    string    // unique identifier for the policy
	Trigger EventType // domain event that activates this policy
	Action  string    // description of the resulting command
}

// PolicyEngine dispatches events to registered handlers.
type PolicyEngine struct {
	handlers map[EventType][]PolicyHandler
}

// PolicyHandler is a function that reacts to a domain event.
type PolicyHandler func(event Event)

// NewPolicyEngine creates a PolicyEngine with no registered handlers.
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		handlers: make(map[EventType][]PolicyHandler),
	}
}

// Register adds a handler for the given event type.
func (pe *PolicyEngine) Register(eventType EventType, handler PolicyHandler) {
	pe.handlers[eventType] = append(pe.handlers[eventType], handler)
}

// Dispatch sends an event to all handlers registered for its type.
func (pe *PolicyEngine) Dispatch(event Event) {
	for _, h := range pe.handlers[event.Type] {
		h(event)
	}
}

// Policies registers all known implicit policies in dominator.
var Policies = []Policy{
	{Name: "ScriptGeneratedNotify", Trigger: EventScriptGenerated, Action: "NotifyGeneration"},
	{Name: "GenerationFailedAlert", Trigger: EventGenerationFailed, Action: "AlertFailure"},
}
