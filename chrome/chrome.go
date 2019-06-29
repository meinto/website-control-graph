package chrome

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	cdproto "github.com/chromedp/cdproto/cdp"
	cdp "github.com/chromedp/chromedp"
	"github.com/meinto/gqlgen-starter/model"
)

var AddExecPathInBuild string = "no"

func CreateContext() (context.Context, context.CancelFunc) {
	opts := []cdp.ExecAllocatorOption{
		cdp.NoFirstRun,
		cdp.NoDefaultBrowserCheck,
		cdp.Headless,
		cdp.DisableGPU,
	}
	if AddExecPathInBuild == "yes" {
		opts = append(opts, cdp.ExecPath("/headless-shell/headless-shell"))
	}
	ctx, _ := cdp.NewExecAllocator(context.Background(), opts...)
	ctx, _ = cdp.NewContext(ctx, cdp.WithDebugf(log.Printf))
	return context.WithTimeout(ctx, 20*time.Second)
}

func Run(actions []*model.Action, mappings []*model.WebsiteElement) ([]*model.Output, error) {
	ctx, cancel := CreateContext()
	defer cancel()

	var tasks cdp.Tasks

	for _, action := range actions {
		if action != nil {
			fields := reflect.TypeOf(action)
			values := reflect.ValueOf(action)

			num := fields.Elem().NumField()

			for i := 0; i < num; i++ {
				field := fields.Elem().Field(i)
				value := values.Elem().Field(i)
				if !value.IsNil() {
					fmt.Print("Type:", field.Type, ",", field.Name, "=", value.Elem().String(), "\n")

					switch field.Name {
					case "Navigate":
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
					}
				}
			}
		}
	}

	type keyNode struct {
		nodes *[]*cdproto.Node
		key   string
	}
	var keyNodes []keyNode
	for _, mapping := range mappings {
		var nodes []*cdproto.Node
		tasks = append(tasks, cdp.Nodes(mapping.Selector, &nodes, cdp.ByQueryAll))
		keyNodes = append(keyNodes, keyNode{
			&nodes,
			mapping.OutKey,
		})
	}

	if err := cdp.Run(ctx, tasks); err != nil {
		return nil, err
	}

	var outmap []*model.Output
	for _, mapping := range mappings {
		for _, keyNode := range keyNodes {
			if keyNode.key == mapping.OutKey {
				for i, node := range *keyNode.nodes {
					if len(node.Children) > 0 && node.Children[0].NodeType == cdproto.NodeTypeText {
						outmap = append(outmap, &model.Output{
							Key:      mapping.OutKey,
							Value:    node.Children[0].NodeValue,
							Index:    i,
							Selector: mapping.Selector,
						})
					}
				}
			}
		}
	}

	return outmap, nil
}
