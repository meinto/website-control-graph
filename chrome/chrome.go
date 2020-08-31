package chrome

import (
	"context"
	"log"
	"time"

	cdp "github.com/chromedp/chromedp"
	"github.com/meinto/website-control-graph/model"
)

var DockerBuild string = "no"
var ShouldLog string = "no"

type Chrome interface {
	CreateContext() (context.Context, context.CancelFunc)
	Run([]*model.Action, []*model.OutputCollectionMap) (*model.Output, error)
	ActionToCDPTasks([]*model.RuntimeVar, *model.Action) (cdp.Tasks, []*model.RuntimeVar)
	CollectDataCDPTasks([]*model.RuntimeVar, []*model.ResultOutputCollectionMap) cdp.Tasks
}

type chrome struct {
	timeout   time.Duration
	omitEmpty bool
}

type outputValues struct {
	val          *string
	element      string
	key          string
	groupElement string
	groupKey     string
}

func New(timeout time.Duration, omitEmpty *bool) Chrome {
	oe := false
	if omitEmpty != nil {
		oe = *omitEmpty
	}
	return &chrome{timeout, oe}
}

func (c *chrome) CreateContext() (context.Context, context.CancelFunc) {
	allocatorOpts := []cdp.ExecAllocatorOption{
		cdp.NoFirstRun,
		cdp.NoDefaultBrowserCheck,
		cdp.Headless,
		cdp.DisableGPU,
		cdp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"),
		cdp.WindowSize(1600, 1200),
	}
	if DockerBuild == "yes" {
		allocatorOpts = append(allocatorOpts, cdp.ExecPath("/headless-shell/headless-shell"))
	}
	ctx, _ := cdp.NewExecAllocator(
		context.Background(),
		allocatorOpts...,
	)

	cdpContextOpts := []cdp.ContextOption{}
	if ShouldLog == "yes" {
		cdpContextOpts = append(cdpContextOpts, cdp.WithDebugf(log.Printf))
	}
	ctx, _ = cdp.NewContext(ctx, cdpContextOpts...)
	return context.WithTimeout(ctx, c.timeout*time.Second)
}

func (c *chrome) Run(actions []*model.Action, mapping []*model.OutputCollectionMap) (*model.Output, error) {
	ctx, cancel := c.CreateContext()
	defer cancel()

	runtimeVars := make([]*model.RuntimeVar, 0)
	for _, action := range actions {
		tasks, rv := c.ActionToCDPTasks(runtimeVars, action)
		runtimeVars = rv

		if err := cdp.Run(ctx, tasks); err != nil {
			return nil, err
		}
	}

	var resultMapping []*model.ResultOutputCollectionMap
	for _, m := range mapping {
		var resultSelectors []*model.ResultSelector
		for _, s := range m.Selectors {
			if s != nil {
				rs := model.NewResultSelector(*s)
				resultSelectors = append(resultSelectors, rs)
			}
		}
		resultMapping = append(resultMapping, &model.ResultOutputCollectionMap{
			OutputCollectionMap: *m,
			ResultSelectors:     resultSelectors,
		})
	}

	tasks := c.CollectDataCDPTasks(runtimeVars, resultMapping)
	if err := cdp.Run(ctx, tasks); err != nil {
		return nil, err
	}

	outmap := c.generateOutput(resultMapping)

	return &model.Output{
		runtimeVars,
		outmap,
	}, nil
}
