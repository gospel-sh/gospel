package gospel

import (
	"fmt"
)

type Context interface {
	AddState(variable ContextStateVariable)
	AddCallback(callback ContextCallbackFunction)
}

type DefaultContext struct {
	variables []ContextStateVariable
	callbacks []ContextCallbackFunction
}

func (d *DefaultContext) AddCallback(callback ContextCallbackFunction) {
	d.callbacks = append(d.callbacks, callback)
	Log.Info("Adding callback no. %d", len(d.callbacks))
	callback.SetId(fmt.Sprintf("%d", len(d.callbacks)))

}

func (d *DefaultContext) AddState(variable ContextStateVariable) {
	d.variables = append(d.variables, variable)
	Log.Info("Adding state variable no. %d", len(d.variables))
	variable.SetId(fmt.Sprintf("%d", len(d.variables)))
}
