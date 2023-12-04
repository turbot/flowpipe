package sanitize

func SanitizeLogEntries(keysAndValues []interface{}) []interface{} {
	if len(keysAndValues)%2 != 0 {
		// empty the whole thing if the keys and values are not in pairs
		return nil
	}

	sanitizeKeyAndValues := make([]interface{}, len(keysAndValues))
	for i := 0; i < len(keysAndValues); i += 2 {
		sanitizeKeyAndValues[i] = keysAndValues[i]

		//nolint:gocritic // TODO: just leave this for now (1 case type swich with asignnment)
		switch keysAndValues[i].(type) {
		case string:
			if Instance.FieldExcluded(keysAndValues[i].(string)) {
				sanitizeKeyAndValues[i+1] = redactedStr
			} else {
				sanitizeKeyAndValues[i+1] = Instance.Sanitize(keysAndValues[i+1])
			}
		default:
			sanitizeKeyAndValues[i+1] = Instance.Sanitize(keysAndValues[i+1])
		}
	}

	return sanitizeKeyAndValues
}
