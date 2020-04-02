package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/support"
)

var (
	DefaultInterceptors = make(Interceptors)
)

type Interceptor interface {
	Name() string
	Handle(ctx *context.Context) (proceed bool, err error)
	HandlesStages() []context.Stage
}

type Interceptors map[context.Stage]map[string]Interceptor

func (instance Interceptors) Handle(ctx *context.Context) (proceed bool, err error) {
	proceed = true
	if candidates, ok := instance[ctx.Stage]; ok {
		for _, candidate := range candidates {
			if proceed, err = candidate.Handle(ctx); !proceed || err != nil {
				return
			}
		}
	}
	return
}

func (instance Interceptors) Add(i Interceptor) Interceptors {
	name := i.Name()
	for _, stage := range i.HandlesStages() {
		if ofStage, found := instance[stage]; found {
			ofStage[name] = i
		} else {
			instance[stage] = map[string]Interceptor{name: i}
		}
	}
	return instance
}

func (instance Interceptors) AddFunc(name string, i InterceptorFunc, stages ...context.Stage) Interceptors {
	return instance.Add(&interceptorFunc{
		name:    name,
		handler: i,
		stages:  stages,
	})
}

func (instance Interceptors) Remove(i Interceptor) Interceptors {
	return instance.RemoveByName(i.Name())
}

func (instance Interceptors) RemoveByName(name string) Interceptors {
	for stage, candidates := range instance {
		delete(candidates, name)
		if len(candidates) <= 0 {
			delete(instance, stage)
		}
	}
	return instance
}

func (instance Interceptors) Clone() Interceptors {
	result := make(Interceptors)
	for stage, candidates := range instance {
		resultCandidates := make(map[string]Interceptor)
		for name, candidate := range candidates {
			resultCandidates[name] = candidate
		}
		result[stage] = resultCandidates
	}
	return result
}

func (instance Interceptors) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	alreadyHandled := make(map[support.FlagRegistrar]bool)
	for _, interceptors := range instance {
		for _, interceptor := range interceptors {
			if fr, ok := interceptor.(support.FlagRegistrar); ok {
				if _, already := alreadyHandled[fr]; !already {
					if err := fr.RegisterFlag(fe, appPrefix); err != nil {
						return err
					}
					alreadyHandled[fr] = true
				}
			}
		}
	}
	return nil
}

func (instance Interceptors) Init(stop support.Channel) error {
	alreadyHandled := make(map[support.Initializable]bool)
	for _, interceptors := range instance {
		for _, interceptor := range interceptors {
			if fr, ok := interceptor.(support.Initializable); ok {
				if _, already := alreadyHandled[fr]; !already {
					if err := fr.Init(stop); err != nil {
						return err
					}
					alreadyHandled[fr] = true
				}
			}
		}
	}
	return nil
}

type InterceptorFunc func(ctx *context.Context) (proceed bool, err error)

type interceptorFunc struct {
	name    string
	handler InterceptorFunc
	stages  []context.Stage
}

func (instance *interceptorFunc) Name() string {
	return instance.name
}

func (instance *interceptorFunc) Handle(ctx *context.Context) (proceed bool, err error) {
	return instance.handler(ctx)
}

func (instance *interceptorFunc) HandlesStages() []context.Stage {
	return instance.stages
}
