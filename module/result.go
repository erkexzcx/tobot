package module

type Result struct {
	CanRepeat bool  // 'true' if OK, 'false' if inventory is full or resources (needed for activity) has depleted
	Error     error // E.g. banned or anything else unexpected
}
