package subjectregistry

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

type LookupError struct {
	Subject string
	Reason  string
}

func (e *LookupError) Error() string {
	if e.Subject == "" {
		return "invalid lookup: " + e.Reason
	}
	return "invalid lookup \"" + e.Subject + "\": " + e.Reason
}

func lookupErr(subject, reason string) error {
	return &LookupError{Subject: subject, Reason: reason}
}

func validLookup(subject string) ([]string, error) {
	if subject == "" {
		return nil, lookupErr(subject, "must not be empty")
	}

	parts := strings.Split(subject, ".")

	for _, part := range parts {
		switch {
		case part == "":
			return nil, lookupErr(subject, "empty token")
		case strings.ContainsAny(part, ">*"):
			return nil, lookupErr(subject, "wildcards not allowed in lookup")
		}
	}

	return parts, nil
}
