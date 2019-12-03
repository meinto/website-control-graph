package model

import (
	"fmt"
	"strings"
)

type ResultSelector struct {
	Selector     Selector
	CSSSelector  string
	SubSelectors *[]ResultSelector
	Result       *string
}

func NewResultSelector(s Selector, parentCSSSelector string) ResultSelector {
	cssSelector := ""
	if len(parentCSSSelector) > 0 {
		cssSelector = parentCSSSelector
	}
	if s.CSSSelector != nil {
		cssSelector = cssSelector + " " + *s.CSSSelector
	}

	if s.SubSelectors != nil {
		var subSelectors []ResultSelector
		for _, ss := range s.SubSelectors {
			subSelectors = append(subSelectors, NewResultSelector(*ss, cssSelector))
		}
		return ResultSelector{
			s,
			cssSelector,
			&subSelectors,
			nil,
		}
	}
	return ResultSelector{
		s,
		cssSelector,
		nil,
		nil,
	}
}

func (s *ResultSelector) Iterate(cb func(*ResultSelector)) {
	cb(s)
	if s.SubSelectors != nil {
		for _, ss := range *s.SubSelectors {
			ss.Iterate(cb)
		}
	}
}

func (s *ResultSelector) CSSSelectorToJS(runtimeVars []*RuntimeVar) string {
	return ReplaceRuntimeTemplates(
		runtimeVars,
		fmt.Sprintf(`Array.from(document.querySelectorAll("%s"))`, s.CSSSelector),
	)
}

func (s *ResultSelector) AddHTMLAttributeSelector(jsString string, runtimeVars []*RuntimeVar) string {
	return jsString + ReplaceRuntimeTemplates(
		runtimeVars,
		fmt.Sprintf(`.map(node => node.getAttribute("%s"))`, *s.Selector.HTMLAttribute),
	)
}

func (s *ResultSelector) AddInnerHTMLSelector(jsString string, runtimeVars []*RuntimeVar) string {
	return jsString + ".map(node => node.innerHTML)"
}

func (s *ResultSelector) GetResultJSONArray() (values []map[string]string) {
	var stringValues []string
	if s.Result != nil {
		stringValues = strings.Split(*s.Result, ";;")
	}

	values = make([]map[string]string, 0)
	for _, stringValue := range stringValues {
		value := make(map[string]string)
		if s.Selector.Key != nil {
			selectorKey := *s.Selector.Key
			value[selectorKey] = stringValue
		}
		values = append(values, value)
	}

	// subSelectorsKey := "subSelectors"
	// if s.Selector.SubSelectorsKey != nil {
	// 	subSelectorsKey = *s.Selector.SubSelectorsKey
	// }

	// } else {
	// 	values = make([]string, 0)
	// 	for _, s := range stringValues {
	// 		values = append(values.([]string), s)
	// 	}
	// }

	return values
}
