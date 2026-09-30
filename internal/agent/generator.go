package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type QueryRejected struct{ Validation Validation }

func (e *QueryRejected) Error() string {
	messages := []string{}
	for _, issue := range e.Validation.Errors {
		messages = append(messages, issue.Message)
	}
	return strings.Join(messages, "；")
}

// Generate accepts structured conditions only and always enforces validation.
func Generate(document Document) (string, error) {
	validation := Validate(document)
	if !validation.Valid {
		return "", &QueryRejected{Validation: validation}
	}
	return render(*validation.Normalized.Condition), nil
}

func render(n Condition) string {
	if n.Type == "and" || n.Type == "or" {
		parts := make([]string, len(n.Conditions))
		for i, child := range n.Conditions {
			parts[i] = render(child)
		}
		joiner := " && "
		if n.Type == "or" {
			joiner = " || "
		}
		return "(" + strings.Join(parts, joiner) + ")"
	}
	var encoded string
	if value, ok := n.Value.(bool); ok {
		encoded = strconv.FormatBool(value)
	} else {
		var out bytes.Buffer
		encoder := json.NewEncoder(&out)
		encoder.SetEscapeHTML(false)
		_ = encoder.Encode(fmt.Sprint(n.Value))
		encoded = strings.TrimSuffix(out.String(), "\n")
	}
	return n.Field + fields[n.Field].Operators[n.Operator] + encoded
}
