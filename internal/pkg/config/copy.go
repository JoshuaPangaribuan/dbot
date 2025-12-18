package config

func copyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = deepCopyAny(v)
	}
	return out
}

func copySlice(in []any) []any {
	if in == nil {
		return nil
	}
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = deepCopyAny(v)
	}
	return out
}

func deepCopyAny(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		return copyMap(typed)
	case []any:
		return copySlice(typed)
	default:
		return v
	}
}
