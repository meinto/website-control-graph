package chrome

import (
	"fmt"
	"log"
	"reflect"
	"time"

	cdp "github.com/chromedp/chromedp"
	"github.com/meinto/website-control-graph/model"
)

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
