package chrome

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	cdp "github.com/chromedp/chromedp"
	"github.com/meinto/website-control-graph/model"
)

var DockerBuild string = "no"
var ShouldLog string = "no"

type Chrome interface {
	CreateContext() (context.Context, context.CancelFunc)
	Run([]*model.Action, []*model.Selector) (*model.Output, error)
	TasksForAction([]*model.RuntimeVar, *model.Action) (cdp.Tasks, []*model.RuntimeVar)
	TasksForOutput([]*model.RuntimeVar, []*model.Selector) (cdp.Tasks, []outputValues)
	MapFoundNodesToOutputStruct([]outputValues, []*model.Selector) []*model.OutputElement
}

type chrome struct {
	timeout time.Duration
}

type outputValues struct {
	val *string
	key string
}

func New(timeout time.Duration) Chrome {
	return &chrome{timeout}
}

func (c *chrome) CreateContext() (context.Context, context.CancelFunc) {
	allocatorOpts := []cdp.ExecAllocatorOption{
		cdp.NoFirstRun,
		cdp.NoDefaultBrowserCheck,
		cdp.Headless,
		cdp.DisableGPU,
	}
	if DockerBuild == "yes" {
		allocatorOpts = append(allocatorOpts, cdp.ExecPath("/headless-shell/headless-shell"))
	}
	ctx, _ := cdp.NewExecAllocator(context.Background(), allocatorOpts...)

	cdpContextOpts := []cdp.ContextOption{}
	if ShouldLog == "yes" {
		cdpContextOpts = append(cdpContextOpts, cdp.WithDebugf(log.Printf))
	}
	ctx, _ = cdp.NewContext(ctx, cdpContextOpts...)
	return context.WithTimeout(ctx, c.timeout*time.Second)
}

func (c *chrome) Run(actions []*model.Action, mapping []*model.Selector) (*model.Output, error) {
	ctx, cancel := c.CreateContext()
	defer cancel()

	runtimeVars := make([]*model.RuntimeVar, 0)
	for _, action := range actions {
		tasks, rv := c.TasksForAction(runtimeVars, action)
		runtimeVars = rv

		if err := cdp.Run(ctx, tasks); err != nil {
			return nil, err
		}
	}

	tasks, outputNodes := c.TasksForOutput(runtimeVars, mapping)
	if err := cdp.Run(ctx, tasks); err != nil {
		return nil, err
	}

	outmap := c.MapFoundNodesToOutputStruct(outputNodes, mapping)

	return &model.Output{
		runtimeVars,
		outmap,
	}, nil
}

func (c *chrome) TasksForAction(runtimeVars []*model.RuntimeVar, action *model.Action) (cdp.Tasks, []*model.RuntimeVar) {
	var tasks cdp.Tasks

	if action != nil {
		fields := reflect.TypeOf(action)
		values := reflect.ValueOf(action)

		num := fields.Elem().NumField()

		for i := 0; i < num; i++ {
			field := fields.Elem().Field(i)
			value := values.Elem().Field(i)

			if !value.IsNil() {
				switch field.Name {
				case "Navigate":
					url := c.ReplaceRuntimeTemplates(runtimeVars, *action.Navigate)
					tasks = append(tasks, cdp.Navigate(url))
					break
				case "Sleep":
					duration := time.Duration(value.Elem().Int()) * time.Second
					tasks = append(tasks, cdp.Sleep(duration))
					break
				case "WaitVisible":
					selector := c.ReplaceRuntimeTemplates(runtimeVars, *action.WaitVisible)
					tasks = append(tasks, cdp.WaitVisible(selector, cdp.ByQuery))
					break
				case "SendKeys":
					selector := c.ReplaceRuntimeTemplates(runtimeVars, action.SendKeys.Selector)
					val := c.ReplaceRuntimeTemplates(runtimeVars, action.SendKeys.Value)
					tasks = append(tasks, cdp.SendKeys(selector, val, cdp.ByQuery))
					break
				case "Click":
					selector := c.ReplaceRuntimeTemplates(runtimeVars, *action.Click)
					tasks = append(tasks, cdp.Click(selector, cdp.ByQuery))
					break
				case "EvalJs":
					js := c.ReplaceRuntimeTemplates(runtimeVars, *action.EvalJs)
					var res []byte
					tasks = append(tasks, cdp.EvaluateAsDevTools(js, &res))
					break
				case "RuntimeVar":
					selector := *action.RuntimeVar
					selectorJS := c.ReplaceRuntimeTemplates(
						runtimeVars,
						fmt.Sprintf(`document.querySelector("%s")`, selector.Element),
					)
					if selector.Attribute != nil {
						selectorJS += c.ReplaceRuntimeTemplates(
							runtimeVars,
							fmt.Sprintf(`.getAttribute("%s")`, *selector.Attribute),
						)
					} else {
						selectorJS += ".innerHTML"
					}
					var res string
					tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &res))
					runtimeVars = append(runtimeVars, &model.RuntimeVar{
						fmt.Sprintf("$%d", len(runtimeVars)),
						selector.Attribute,
						selector.Element,
						&res,
					})
					break
				}
			}
		}
	}

	return tasks, runtimeVars
}

func (c *chrome) ReplaceRuntimeTemplates(runtimeVars []*model.RuntimeVar, sourceString string) string {
	s := sourceString
	for _, v := range runtimeVars {
		s = strings.ReplaceAll(s, v.Name, *v.Value)
		log.Println(s, v.Name, *v.Value)
	}
	return s
}

func (c *chrome) TasksForOutput(runtimeVars []*model.RuntimeVar, mapping []*model.Selector) (cdp.Tasks, []outputValues) {
	var tasks cdp.Tasks
	var ov []outputValues
	for _, m := range mapping {
		selector := m.Element
		selectorJS := c.ReplaceRuntimeTemplates(
			runtimeVars,
			fmt.Sprintf(`Array.prototype.slice.call(document.querySelectorAll("%s"))`, selector),
		)
		if m.Attribute != nil {
			selectorJS += c.ReplaceRuntimeTemplates(
				runtimeVars,
				fmt.Sprintf(`.map(node => node.getAttribute("%s"))`, *m.Attribute),
			)
		} else {
			selectorJS += ".map(node => node.innerHTML)"
		}
		selectorJS += ".join(';;')"
		var res string
		tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &res))
		ov = append(ov, outputValues{
			&res,
			*m.Key,
		})
	}

	return tasks, ov
}

func (c *chrome) MapFoundNodesToOutputStruct(outputValues []outputValues, mapping []*model.Selector) (outmap []*model.OutputElement) {
	for _, m := range mapping {
		for _, ov := range outputValues {
			if m.Key != nil && ov.key == *m.Key {
				vs := strings.Split(*ov.val, ";;")
				for i, v := range vs {
					outmap = append(outmap, &model.OutputElement{
						Key:      *m.Key,
						Value:    v,
						Index:    i,
						Selector: m.Element,
					})
				}
			}
		}
	}

	return outmap
}
