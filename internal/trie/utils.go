package trie

import "strings"

type SubError struct {
	Sub    string
	Reason string
}

func (e *SubError) Error() string {
	if e.Sub == "" {
		return "invalid sub: " + e.Reason
	}
	return "invalid sub \"" + e.Sub + "\": " + e.Reason
}

func subErr(sub, reason string) error {
	return &SubError{Sub: sub, Reason: reason}
}

func validSub(sub string) ([]string, error) {
	if sub == "" {
		return nil, subErr(sub, "must not be empty")
	}

	parts := strings.Split(sub, ".")

	for i, part := range parts {
		switch {
		case part == "":
			return nil, subErr(sub, "empty token")
		case part == "*":
			// valid single-level wildcard
		case part == ">":
			if i != len(parts)-1 {
				return nil, subErr(sub, "'>' must be the last token")
			}
		case strings.ContainsAny(part, ">*"):
			return nil, subErr(sub, "wildcards must be standalone tokens")
		}
	}

	return parts, nil
}
