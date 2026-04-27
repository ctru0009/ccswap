package config

func MaskKey(key string) string {
	if key == "" {
		return "<empty>"
	}
	if len(key) < 11 {
		return "****"
	}
	return key[:7] + "..." + key[len(key)-4:]
}