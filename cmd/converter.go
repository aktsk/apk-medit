package cmd

func intToUTF8bytes(arg string) []byte {
	rs := []rune(arg)
	return []byte(string(rs))
}
