package service

import "github.com/tidwall/gjson"

// Visit schema positions only: a property/default/example named required is data.
func collectNullRequired(body []byte, schema gjson.Result, depth int, hits *[]openAIResponsesToolSchemaNullType) {
	if depth > 128 || !schema.IsObject() {
		return
	}
	if required := schema.Get("required"); required.Raw == "null" {
		n := len(*hits)
		appendOpenAIResponsesToolSchemaNullType(body, required, hits)
		if len(*hits) > n {
			(*hits)[n].replacement = "[]"
		}
	}
	schema.ForEach(func(key, value gjson.Result) bool {
		switch key.String() {
		case "properties", "patternProperties", "$defs", "definitions", "dependentSchemas":
			value.ForEach(func(_, child gjson.Result) bool { collectNullRequired(body, child, depth+1, hits); return true })
		case "allOf", "anyOf", "oneOf", "prefixItems", "items":
			if value.IsArray() {
				value.ForEach(func(_, child gjson.Result) bool { collectNullRequired(body, child, depth+1, hits); return true })
			} else {
				collectNullRequired(body, value, depth+1, hits)
			}
		case "additionalProperties", "additionalItems", "contains", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems", "contentSchema":
			collectNullRequired(body, value, depth+1, hits)
		}
		return true
	})
}
