package pipeline

import "fmt"

// ErrorPolicy controls how the pipeline handles a processor failure.
type ErrorPolicy string

const (
	// ErrorPolicyUsePrevious restores the text from before the failed
	// processor and continues with the next processor.
	ErrorPolicyUsePrevious ErrorPolicy = "use_previous"

	// ErrorPolicyUseOriginal restores the pipeline's original input text and
	// continues with the next processor.
	ErrorPolicyUseOriginal ErrorPolicy = "use_original"

	// ErrorPolicySkip restores the text from before the failed processor and
	// stops the remaining processors without returning an error.
	ErrorPolicySkip ErrorPolicy = "skip"

	// ErrorPolicyAbort restores the text from before the failed processor and
	// returns the processor error.
	ErrorPolicyAbort ErrorPolicy = "abort"
)

// ParseErrorPolicy converts a configured value into an ErrorPolicy.
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
