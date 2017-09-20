package openbmp

import "unicode"

//https://github.com/asaskevich/govalidator
func camelCaseToUnderscore(str string) string {
	addSegment := func(inrune, segment []rune) []rune {
		if len(segment) == 0 {
			return inrune
		}
		if len(inrune) != 0 {
			inrune = append(inrune, '_')
		}
		inrune = append(inrune, segment...)
		return inrune
	}

	var output []rune
	var segment []rune
	for _, r := range str {
		if !unicode.IsLower(r) {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}
