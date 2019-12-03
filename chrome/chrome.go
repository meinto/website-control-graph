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

	var resultMapping = []*model.ResultOutputCollectionMap{}
	for _, m := range mapping {
		resultMapping = append(resultMapping, &model.ResultOutputCollectionMap{
			OutputCollectionMap: *m,
			Result:              nil,
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

func (c *chrome) ActionToCDPTasks(runtimeVars []*model.RuntimeVar, action *model.Action) (cdp.Tasks, []*model.RuntimeVar) {
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
					url := model.ReplaceRuntimeTemplates(runtimeVars, *action.Navigate)
					tasks = append(tasks, cdp.Navigate(url))
					break
				case "Sleep":
					duration := time.Duration(value.Elem().Int()) * time.Second
					tasks = append(tasks, cdp.Sleep(duration))
					break
				case "WaitVisible":
					selector := model.ReplaceRuntimeTemplates(runtimeVars, *action.WaitVisible)
					tasks = append(tasks, cdp.WaitVisible(selector, cdp.ByQuery))
					break
				case "SendKeys":
					selector := model.ReplaceRuntimeTemplates(runtimeVars, action.SendKeys.CSSSelector)
					val := model.ReplaceRuntimeTemplates(runtimeVars, action.SendKeys.Value)
					tasks = append(tasks, cdp.SendKeys(selector, val, cdp.ByQuery))
					break
				case "Click":
					selector := model.ReplaceRuntimeTemplates(runtimeVars, *action.Click)
					tasks = append(tasks, cdp.Click(selector, cdp.ByQuery))
					break
				case "EvalJs":
					js := model.ReplaceRuntimeTemplates(runtimeVars, *action.EvalJs)
					var res []byte
					tasks = append(tasks, cdp.EvaluateAsDevTools(js, &res))
					break
				case "RuntimeVar":
					selector := *action.RuntimeVar
					if selector.CSSSelector != nil {
						selectorJS := model.ReplaceRuntimeTemplates(
							runtimeVars,
							fmt.Sprintf(`document.querySelector("%s")`, *selector.CSSSelector),
						)
						if selector.HTMLAttribute != nil {
							selectorJS += model.ReplaceRuntimeTemplates(
								runtimeVars,
								fmt.Sprintf(`.getAttribute("%s")`, *selector.HTMLAttribute),
							)
						} else {
							selectorJS += ".innerHTML"
						}
						var res string
						tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &res))
						runtimeVars = append(runtimeVars, &model.RuntimeVar{
							fmt.Sprintf("$%d", len(runtimeVars)),
							selector.HTMLAttribute,
							*selector.CSSSelector,
							&res,
						})
					} else {
						log.Println("missing css selector for runtime var")
					}
					break
				}
			}
		}
	}

	return tasks, runtimeVars
}

func (c *chrome) CollectDataCDPTasks(runtimeVars []*model.RuntimeVar, mapping []*model.ResultOutputCollectionMap) cdp.Tasks {
	var tasks cdp.Tasks
	for _, m := range mapping {

		selectorJS := m.Selector.CSSSelectorToJS(runtimeVars, false)

		if m.Selector.HTMLAttribute != nil {
			selectorJS = m.Selector.AddHTMLAttributeSelector(selectorJS, runtimeVars)
		} else {
			selectorJS = m.Selector.AddInnerHTMLSelector(selectorJS, runtimeVars)
		}

		selectorJS += ".join(';;')"
		log.Println(selectorJS)

		tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &m.Result))
	}

	return tasks
}

func (c *chrome) generateOutput(resultMapping []*model.ResultOutputCollectionMap) (output map[string]interface{}) {
	output = make(map[string]interface{})
	for _, m := range resultMapping {
		collection := m.Name
		if m.Key != nil {
			collection = *m.Key
		}

		var stringValues []string
		if m.Result != nil {
			stringValues = strings.Split(*m.Result, ";;")
		}

		var values interface{}
		if m.Selector.Flat == nil || *m.Selector.Flat == false {
			values = make([]map[string]string, 0)
			for _, s := range stringValues {
				value := make(map[string]string)
				selectorKey := "value"
				if m.Selector.Key != nil {
					selectorKey = *m.Selector.Key
				}
				value[selectorKey] = s
				values = append(values.([]map[string]string), value)
			}
		} else {
			values = make([]string, 0)
			for _, s := range stringValues {
				values = append(values.([]string), s)
			}
		}

		data := make(map[string]interface{})

		selectorKey := "data"
		if m.SelectorKey != nil {
			selectorKey = *m.SelectorKey
		}
		data[selectorKey] = values
		data["collection"] = m.Name

		output[collection] = data
	}
	return output
}
