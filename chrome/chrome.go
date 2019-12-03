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
	Run([]*model.Action, []*model.OutputCollections) (*model.Output, error)
	ActionToCDPTasks([]*model.RuntimeVar, *model.Action) (cdp.Tasks, []*model.RuntimeVar)
	CollectDataCDPTasks([]*model.RuntimeVar, []*model.OutputCollections) (cdp.Tasks, []outputValues)
	MapFoundNodesToOutputStruct([]outputValues) []*model.OutputCollection
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

func (c *chrome) Run(actions []*model.Action, mapping []*model.OutputCollections) (*model.Output, error) {
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

	tasks, outputNodes := c.CollectDataCDPTasks(runtimeVars, mapping)
	if err := cdp.Run(ctx, tasks); err != nil {
		return nil, err
	}

	outmap := c.MapFoundNodesToOutputStruct(outputNodes)

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
					selector := c.ReplaceRuntimeTemplates(runtimeVars, action.SendKeys.CSSSelector)
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
						fmt.Sprintf(`document.querySelector("%s")`, selector.CSSSelector),
					)
					if selector.HTMLAttribute != nil {
						selectorJS += c.ReplaceRuntimeTemplates(
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
						selector.CSSSelector,
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

func (c *chrome) CollectDataCDPTasks(runtimeVars []*model.RuntimeVar, mapping []*model.OutputCollections) (cdp.Tasks, []outputValues) {
	var tasks cdp.Tasks
	var ov []outputValues
	for _, m := range mapping {
		groupElement := ""
		groupKey := "default"

		selectorJS := ""
		if m.GroupElement != nil {
			groupElement = *m.GroupElement
			selectorJS += c.ReplaceRuntimeTemplates(
				runtimeVars,
				fmt.Sprintf(`Array.from(document.querySelectorAll("%s"))`, groupElement),
			)
		}
		if m.GroupKey != nil {
			groupKey = *m.GroupKey
		}

		if m.GroupElement != nil {
			groupElement = *m.GroupElement
			selectorJS += c.ReplaceRuntimeTemplates(
				runtimeVars,
				fmt.Sprintf(`.map(group => Array.from(group.querySelectorAll("%s"))`, m.Element),
			)
		} else {
			selectorJS += c.ReplaceRuntimeTemplates(
				runtimeVars,
				fmt.Sprintf(`Array.from(document.querySelectorAll("%s"))`, m.Element),
			)
		}

		if m.Attribute != nil {
			selectorJS += c.ReplaceRuntimeTemplates(
				runtimeVars,
				fmt.Sprintf(`.map(node => node.getAttribute("%s"))`, *m.Attribute),
			)
		} else {
			selectorJS += ".map(node => node.innerHTML)"
		}

		selectorJS += ".join(';;')"
		if m.GroupElement != nil {
			selectorJS += ").join('##')"
		}

		var res string
		tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &res))

		ov = append(ov, outputValues{
			val:          &res,
			element:      m.Element,
			key:          m.Key,
			groupElement: groupElement,
			groupKey:     groupKey,
		})
	}

	return tasks, ov
}

func (c *chrome) MapFoundNodesToOutputStruct(outputValues []outputValues) (outmap []*model.OutputCollection) {
	for _, ov := range outputValues {
		groups := strings.Split(*ov.val, "##")
		for gi, group := range groups {
			vs := strings.Split(group, ";;")
			for i, v := range vs {
				if c.omitEmpty == false || (c.omitEmpty == true && len(v) > 0) {
					outmap = append(outmap, &model.OutputCollection{
						Key:          ov.key,
						Value:        v,
						Index:        i,
						Element:      ov.element,
						GroupElement: ov.groupElement,
						GroupIndex:   gi,
						GroupKey:     ov.groupKey,
					})
				}
			}
		}
	}

	return outmap
}
