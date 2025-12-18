package logger

import "context"

type ContextFieldsFunc func(context.Context) Fields

func mergeFields(base Fields, override Fields) Fields {
	if len(base) == 0 {
		return override
	}
	if len(override) == 0 {
		return base
	}

	out := make(Fields, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
