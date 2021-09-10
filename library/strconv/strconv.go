package strconv

func Itoa(arg int) string {
	if arg == 0 {
		return "0"
	}
	var rs = []rune{}
	for arg != 0 {
		rs = append(rs, rune('0'+arg%10))
		arg = arg / 10
	}

	var revRs = []rune{}
	for i := 0; i < len(rs); i = i + 1 {
		revRs = append(revRs, rs[len(rs)-i-1])
	}
	revRs = append(revRs, rune(0))

	return string(revRs)
}
