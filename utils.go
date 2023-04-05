package gospel

func If(condition bool, value any) any {
	if condition {
		return value
	}
	return nil
}
