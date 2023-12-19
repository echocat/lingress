package proxy

import (
	"github.com/echocat/lingress/context"
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

func (this Interceptors) Handle(ctx *context.Context) (proceed bool, err error) {
	proceed = true
	if candidates, ok := this[ctx.Stage]; ok {
		for _, candidate := range candidates {
			if proceed, err = candidate.Handle(ctx); !proceed || err != nil {
				return
			}
		}
	}
	return
}

func (this Interceptors) Add(i Interceptor) Interceptors {
	name := i.Name()
	for _, stage := range i.HandlesStages() {
		if ofStage, found := this[stage]; found {
			ofStage[name] = i
		} else {
			this[stage] = map[string]Interceptor{name: i}
		}
	}
	return this
}

func (this Interceptors) AddFunc(name string, i InterceptorFunc, stages ...context.Stage) Interceptors {
	return this.Add(&interceptorFunc{
		name:    name,
		handler: i,
		stages:  stages,
	})
}

func (this Interceptors) Remove(i Interceptor) Interceptors {
	return this.RemoveByName(i.Name())
}

func (this Interceptors) RemoveByName(name string) Interceptors {
	for stage, candidates := range this {
		delete(candidates, name)
		if len(candidates) <= 0 {
			delete(this, stage)
		}
	}
	return this
}

func (this Interceptors) Clone() Interceptors {
	result := make(Interceptors)
	for stage, candidates := range this {
		resultCandidates := make(map[string]Interceptor)
		for name, candidate := range candidates {
			resultCandidates[name] = candidate
		}
		result[stage] = resultCandidates
	}
	return result
}

type InterceptorFunc func(ctx *context.Context) (proceed bool, err error)

type interceptorFunc struct {
	name    string
	handler InterceptorFunc
	stages  []context.Stage
}

func (this *interceptorFunc) Name() string {
	return this.name
}

func (this *interceptorFunc) Handle(ctx *context.Context) (proceed bool, err error) {
	return this.handler(ctx)
}

func (this *interceptorFunc) HandlesStages() []context.Stage {
	return this.stages
}
