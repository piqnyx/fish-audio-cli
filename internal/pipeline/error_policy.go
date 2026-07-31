package pipeline

import "fmt"

type ErrorPolicy string

const (
	ErrorPolicyUsePrevious ErrorPolicy = "use_previous"
	ErrorPolicyUseOriginal ErrorPolicy = "use_original"
	ErrorPolicySkip        ErrorPolicy = "skip"
	ErrorPolicyAbort       ErrorPolicy = "abort"
)

func ParseErrorPolicy(value string) (ErrorPolicy, error) {
	policy := ErrorPolicy(value)

	switch policy {
	case ErrorPolicyUsePrevious,
		ErrorPolicyUseOriginal,
		ErrorPolicySkip,
		ErrorPolicyAbort:
		return policy, nil
	default:
		return "", fmt.Errorf("unsupported pipeline error policy %q", value)
	}
}
