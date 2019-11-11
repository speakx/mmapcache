package byteio

// BytesCmp 判断2个[]byte是否完全相等
// 比默认包里面的Reflect.DeepEqual快
func BytesCmp(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
