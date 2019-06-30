package chrome

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	cdproto "github.com/chromedp/cdproto/cdp"
	cdp "github.com/chromedp/chromedp"
	"github.com/meinto/website-control-graph/model"
)

var DockerBuild string = "no"

type Chrome interface {
	CreateContext() (context.Context, context.CancelFunc)
	Run(actions []*model.Action, mappings []*model.OutputSelector) ([]*model.Output, error)
	TasksFromActions(tasks cdp.Tasks, actions []*model.Action) cdp.Tasks
	TasksForOutput(tasks cdp.Tasks, mapping []*model.OutputSelector) (cdp.Tasks, []outputNodes)
	MapFoundNodesToOutputStruct(outputNodesList []outputNodes, mapping []*model.OutputSelector) (outmap []*model.Output)
}

type chrome struct {
	timeout time.Duration
}

type outputNodes struct {
	nodes *[]*cdproto.Node
	key   string
}

func New(timeout time.Duration) Chrome {
	return &chrome{timeout}
}

func (c *chrome) CreateContext() (context.Context, context.CancelFunc) {
	opts := []cdp.ExecAllocatorOption{
		cdp.NoFirstRun,
		cdp.NoDefaultBrowserCheck,
		cdp.Headless,
		cdp.DisableGPU,
	}
	if DockerBuild == "yes" {
		opts = append(opts, cdp.ExecPath("/headless-shell/headless-shell"))
	}
	ctx, _ := cdp.NewExecAllocator(context.Background(), opts...)
	ctx, _ = cdp.NewContext(ctx, cdp.WithDebugf(log.Printf))
	return context.WithTimeout(ctx, c.timeout*time.Second)
}

func (c *chrome) Run(actions []*model.Action, mapping []*model.OutputSelector) ([]*model.Output, error) {
	ctx, cancel := c.CreateContext()
	defer cancel()

	var tasks cdp.Tasks
	tasks = c.TasksFromActions(tasks, actions)
	tasks, outputNodes := c.TasksForOutput(tasks, mapping)

	if err := cdp.Run(ctx, tasks); err != nil {
		return nil, err
	}

	outmap := c.MapFoundNodesToOutputStruct(outputNodes, mapping)

	return outmap, nil
}

func (c *chrome) TasksFromActions(tasks cdp.Tasks, actions []*model.Action) cdp.Tasks {
	for _, action := range actions {
		if action != nil {
			fields := reflect.TypeOf(action)
			values := reflect.ValueOf(action)

			num := fields.Elem().NumField()

			tmp := make([]*string, 0)

			for i := 0; i < num; i++ {
				field := fields.Elem().Field(i)
				value := values.Elem().Field(i)

				if !value.IsNil() {
					switch field.Name {
					case "Navigate":
						// url := *action.Navigate
						// r := regexp.Compile("(\$[0-9]+)")
						// url = fmt.Sprintf(url)
						tasks = append(tasks, cdp.Navigate(value.Elem().String()))
						break
					case "Sleep":
						duration := time.Duration(value.Elem().Int()) * time.Second
						tasks = append(tasks, cdp.Sleep(duration))
						break
					case "WaitVisible":
						selector := *action.WaitVisible
						tasks = append(tasks, cdp.WaitVisible(selector, cdp.ByQuery))
						break
					case "SendKeys":
						selector := action.SendKeys.Selector
						val := action.SendKeys.Value
						tasks = append(tasks, cdp.SendKeys(selector, val, cdp.ByQuery))
						break
					case "Click":
						selector := *action.Click
						tasks = append(tasks, cdp.Click(selector, cdp.ByQuery))
						break
					case "EvalJs":
						js := *action.EvalJs
						var res []byte
						tasks = append(tasks, cdp.EvaluateAsDevTools(js, &res))
						break
					case "Store":
						selector := *action.Store
						selectorJS := fmt.Sprintf(`document.querySelector("%s")`, selector.Element)
						if selector.Attribute != nil {
							selectorJS += fmt.Sprintf(`.getAttribute("%s")`, selector.Attribute)
						} else {
							selectorJS += ".innerHTML"
						}
						var res string
						tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &res))
						tmp = append(tmp, &res)
						break
					}
				}
			}
		}
	}

	return tasks
}

func (c *chrome) TasksForOutput(tasks cdp.Tasks, mapping []*model.OutputSelector) (cdp.Tasks, []outputNodes) {
	var outputNodesList []outputNodes
	for _, m := range mapping {
		var nodes []*cdproto.Node
		tasks = append(tasks, cdp.Nodes(m.Selector, &nodes, cdp.ByQueryAll))
		outputNodesList = append(outputNodesList, outputNodes{
			&nodes,
			m.Key,
		})
	}

	return tasks, outputNodesList
}

func (c *chrome) MapFoundNodesToOutputStruct(outputNodesList []outputNodes, mapping []*model.OutputSelector) (outmap []*model.Output) {
	for _, m := range mapping {
		for _, outputNodes := range outputNodesList {
			if outputNodes.key == m.Key {
				for i, node := range *outputNodes.nodes {
					if len(node.Children) > 0 && node.Children[0].NodeType == cdproto.NodeTypeText {
						outmap = append(outmap, &model.Output{
							Key:      m.Key,
							Value:    node.Children[0].NodeValue,
							Index:    i,
							Selector: m.Selector,
						})
					}
				}
			}
		}
	}

	return outmap
}
