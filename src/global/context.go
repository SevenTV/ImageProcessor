package global

import (
	"context"
	"sync"

	"github.com/seventv/EmoteProcessor/src/configure"
)

type Context interface {
	context.Context
	Instances() *Instances
	Config() *configure.Config
	AddTask(n int)
	DoneTask()
	Wait()
}

type GlobalContext struct {
	context.Context
	Insts *Instances
	Cfg   *configure.Config
	wg    *sync.WaitGroup
}

func New(ctx context.Context, config *configure.Config) Context {
	return &GlobalContext{
		Context: ctx,
		Insts:   &Instances{},
		Cfg:     config,
		wg:      &sync.WaitGroup{},
	}
}

func (g *GlobalContext) Instances() *Instances {
	return g.Insts
}

func (g *GlobalContext) Config() *configure.Config {
	return g.Cfg
}

func (g *GlobalContext) AddTask(n int) {
	g.wg.Add(n)
}

func (g *GlobalContext) DoneTask() {
	g.wg.Done()
}

func (g *GlobalContext) Wait() {
	g.wg.Wait()
}
