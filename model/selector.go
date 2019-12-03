package model

import "fmt"

// 		HTMLAttribute: String
//   innerHTML: Boolean
//   regex: String
//   subSelector: Selector
//   cssSelector: String!

func (s *Selector) CSSSelectorToJS(runtimeVars []*RuntimeVar, isSubSelector bool) string {
	selectorJS := ""

	// if !isSubSelector {
	selectorJS += ReplaceRuntimeTemplates(
		runtimeVars,
		fmt.Sprintf(`Array.from(document.querySelectorAll("%s"))`, s.CSSSelector),
	)
	// } else {
	// 	selectorJS += ReplaceRuntimeTemplates(
	// 		runtimeVars,
	// 		fmt.Sprintf(`.map(group => Array.from(group.querySelectorAll("%s"))`, s.SubSelector),
	// 	)
	// }

	// if s.SubSelector != nil {
	// 	selectorJS += s.SubSelector.CSSSelectorToJS(runtimeVars, true)
	// }

	return selectorJS
}

func (s *Selector) AddHTMLAttributeSelector(jsString string, runtimeVars []*RuntimeVar) string {
	return jsString + ReplaceRuntimeTemplates(
		runtimeVars,
		fmt.Sprintf(`.map(node => node.getAttribute("%s"))`, *s.HTMLAttribute),
	)
}

func (s *Selector) AddInnerHTMLSelector(jsString string, runtimeVars []*RuntimeVar) string {
	return jsString + ".map(node => node.innerHTML)"
}
