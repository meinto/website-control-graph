package model

import (
	"fmt"
)

type ResultSelector struct {
	Selector     Selector
	SubSelectors []*ResultSelector
	Result       *interface{}
}

func NewResultSelector(s Selector) *ResultSelector {
	if s.SubSelectors != nil {
		var subSelectors []*ResultSelector
		for _, ss := range s.SubSelectors {
			if ss != nil {
				resultSelector := NewResultSelector(*ss)
				subSelectors = append(subSelectors, resultSelector)
			}
		}
		return &ResultSelector{
			s,
			subSelectors,
			nil,
		}
	}
	return &ResultSelector{
		s,
		nil,
		nil,
	}
}

func (s *ResultSelector) Iterate(cb func(*ResultSelector)) {
	cb(s)
	if s.SubSelectors != nil {
		for _, ss := range s.SubSelectors {
			ss.Iterate(cb)
		}
	}
}

func (s *ResultSelector) GetJS(runtimeVars []*RuntimeVar, parantType SelectorType, hideDefaultValueKeys *bool) (jsString string) {

	isRootSelector := !parantType.IsValid()

	if isRootSelector {
		jsString += fmt.Sprintf(`
			let removeDefaultValueKeys_%s = false;
			function cleanup_%s(result) {
				if (result) {
					if (result.forEach !== undefined) {
						result.forEach(cleanup_%s)
					} else if (typeof result === 'object') {
						if (removeDefaultValueKeys_%s) delete result.__value
						delete result.node
						Object.keys(result).forEach(key => {
							if (result[key] && typeof result[key] === 'string') result[key] = result[key].trim()
							cleanup_%s(result[key])
						})
					}
				}
			}; 
		`, s.Selector.Key, s.Selector.Key, s.Selector.Key, s.Selector.Key, s.Selector.Key)
	}

	if isRootSelector {
		jsString += fmt.Sprintf(`const %s = `, s.Selector.Key)
	}

	if isRootSelector && s.Selector.CSSSelector != nil {
		if s.Selector.Type == SelectorTypeObjectArray || s.Selector.Type == SelectorTypeStringArray {
			jsString += fmt.Sprintf(`Array.from(document.querySelectorAll("%s"))`, *s.Selector.CSSSelector)
		} else if s.Selector.Type == SelectorTypeObjectProp || s.Selector.Type == SelectorTypeStringProp {
			jsString += fmt.Sprintf(`[document.querySelector("%s")]`, *s.Selector.CSSSelector)
		}
		jsString += `.map(node => ({ __value: node, node: node }))`

	} else if !isRootSelector && s.Selector.CSSSelector != nil {
		if s.Selector.Type == SelectorTypeObjectArray || s.Selector.Type == SelectorTypeStringArray {
			jsString += fmt.Sprintf(`Array.from(object.node.querySelectorAll("%s"))`, *s.Selector.CSSSelector)
		} else if s.Selector.Type == SelectorTypeObjectProp || s.Selector.Type == SelectorTypeStringProp {
			jsString += fmt.Sprintf(`[object.node.querySelector("%s")]`, *s.Selector.CSSSelector)
		}
		jsString += `.map(node => ({ __value: node, node: node }))`
	} else if !isRootSelector {
		jsString += `[object]`
	}

	jsString += s.getResultNodeMutations()
	jsString += s.getResultRegexMutations()

	if s.Selector.Type == SelectorTypeStringArray || s.Selector.Type == SelectorTypeStringProp {
		jsString += `.map(object => object.__value)`
	}

	if s.Selector.Type == SelectorTypeObjectArray || s.Selector.Type == SelectorTypeObjectProp {
		for _, ss := range s.SubSelectors {
			jsString += ss.subSelectorJS(s.Selector.Type, runtimeVars, hideDefaultValueKeys)
		}
	}

	if s.Selector.Type == SelectorTypeObjectProp || s.Selector.Type == SelectorTypeStringProp {
		jsString += `.pop()`
	}

	if isRootSelector {
		if hideDefaultValueKeys != nil && *hideDefaultValueKeys {
			jsString += fmt.Sprintf(`; removeDefaultValueKeys_%s = true`, s.Selector.Key)
		}
		jsString += fmt.Sprintf(`; cleanup_%s(%s); %s`, s.Selector.Key, s.Selector.Key, s.Selector.Key)
	}

	return ReplaceRuntimeTemplates(runtimeVars, jsString)
}

func (s *ResultSelector) subSelectorJS(parantType SelectorType, runtimeVars []*RuntimeVar, hideDefaultValueKeys *bool) (jsString string) {
	return fmt.Sprintf(`.map(object => ({ ...object, %s: %s }))`, s.Selector.Key, s.GetJS(runtimeVars, parantType, hideDefaultValueKeys))
}

func (s *ResultSelector) getResultNodeMutations() (jsString string) {
	if s.Selector.HTMLAttribute != nil {
		jsString += fmt.Sprintf(`.map(object => ({ ...object, __value: object.node.getAttribute("%s") }))`, *s.Selector.HTMLAttribute)
	} else {
		jsString += `.map(object => ({ ...object, __value: object.node.innerHTML }))`
	}
	return jsString
}

func (s *ResultSelector) getResultRegexMutations() (jsString string) {
	if s.Selector.Regex != nil {
		jsString += fmt.Sprintf(`.map(object => ({ ...object, __value: [object.__value.match(%s)].filter(n => n).map(matches => matches.pop()).pop() }))`, *s.Selector.Regex)
	}
	return jsString
}
