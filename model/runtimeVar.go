package model

import (
	"log"
	"strings"
)

func ReplaceRuntimeTemplates(runtimeVars []*RuntimeVar, sourceString string) string {
	s := sourceString
	for _, v := range runtimeVars {
		s = strings.ReplaceAll(s, v.Name, *v.Value)
		log.Println(s, v.Name, *v.Value)
	}
	return s
}
