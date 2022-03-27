package pkg

import (
	"devzat/pkg/commands/dm"
	"errors"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type commandRegistry = map[string]CommandFunc

type Registrar struct {
	commandRegistry
}

func (r *Registrar) Register(cr CommandRegistration) error {
	if cr == nil {
		return errors.New("empty command registration given")
	}

	if cr.Name() == "" {
		return errors.New("empty command name given")
	}

	if cr.Fn == nil {
		return errors.New("nil command func given")
	}

	if r.commandRegistry == nil {
		r.commandRegistry = make(commandRegistry)
	}

	r.commandRegistry[cr.Name()] = cr.Fn

	return nil
}

func (r *Registrar) init() error {
	commandsToRegister := []CommandRegistration{
		&dm.Command{},
	}

	for _, cr := range commandsToRegister {
		if err := r.Register(cr); err != nil {
			return err
		}
	}

	return nil
}
