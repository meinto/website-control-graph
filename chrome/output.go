package chrome

import (
	"log"

	cdp "github.com/chromedp/chromedp"
	"github.com/meinto/website-control-graph/model"
)

func (c *chrome) CollectDataCDPTasks(runtimeVars []*model.RuntimeVar, mapping []*model.ResultOutputCollectionMap) cdp.Tasks {
	var tasks cdp.Tasks
	for _, m := range mapping {
		for _, rs := range m.ResultSelectors {
			selectorJS := rs.GetJS(runtimeVars, "")

			log.Println(selectorJS)

			tasks = append(tasks, cdp.EvaluateAsDevTools(selectorJS, &rs.Result))
		}
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

		data := make(map[string]interface{})

		for _, rs := range m.ResultSelectors {
			values := rs.Result
			data[rs.Selector.Key] = values
		}

		data["collection"] = m.Name

		output[collection] = data
	}
	return output
}
