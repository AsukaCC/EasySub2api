package service

import (
	"github.com/tidwall/gjson"
	"strings"
)

// collectSystemOneInput moderates every client-controlled text of a System One
// request: question IDs, every question field except the validated type,
// unknown top-level extension fields, and the evaluated state. Object keys are
// sent to Jev as part of the JSON, so they are moderated like values.
func collectSystemOneInput(body []byte, parts *[]string) {
	root := gjson.ParseBytes(body)
	questions := root.Get("questions")
	if !questions.IsObject() {
		collectSystemOneText(questions, parts)
	}
	questions.ForEach(func(id, question gjson.Result) bool {
		addSystemOneModerationText(parts, id.String())
		if !question.IsObject() {
			collectSystemOneText(question, parts)
			return true
		}
		question.ForEach(func(field, value gjson.Result) bool {
			switch field.String() {
			case "type":
				return true
			case "instructions", "criteria":
			default:
				addSystemOneModerationText(parts, field.String())
			}
			collectSystemOneText(value, parts)
			return true
		})
		return true
	})
	root.ForEach(func(field, value gjson.Result) bool {
		switch field.String() {
		case "model", "stream", "state", "questions":
			return true
		}
		addSystemOneModerationText(parts, field.String())
		collectSystemOneText(value, parts)
		return true
	})
	collectSystemOneText(root.Get("state"), parts)
}

func collectSystemOneText(value gjson.Result, parts *[]string) {
	switch {
	case !value.Exists():
		return
	case value.Type == gjson.String:
		addSystemOneModerationText(parts, value.String())
	case value.IsArray():
		value.ForEach(func(_, child gjson.Result) bool {
			collectSystemOneText(child, parts)
			return true
		})
	case value.IsObject():
		value.ForEach(func(key, child gjson.Result) bool {
			addSystemOneModerationText(parts, key.String())
			collectSystemOneText(child, parts)
			return true
		})
	}
}
func addSystemOneModerationText(parts *[]string, text string) {
	if text = strings.TrimSpace(text); text != "" {
		*parts = append(*parts, text)
	}
}
