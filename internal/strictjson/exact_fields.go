package strictjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var (
	jsonRawMessageType = reflect.TypeOf(
		json.RawMessage{},
	)
	jsonUnmarshalerType = reflect.TypeOf(
		(*json.Unmarshaler)(nil),
	).Elem()
)

// decodeTargetType validates a decode target and returns the type stored by
// its outer pointer.
func decodeTargetType(target any) (reflect.Type, error) {
	if target == nil {
		return nil, fmt.Errorf(
			"JSON decode target is nil",
		)
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Pointer {
		return nil, fmt.Errorf(
			"JSON decode target must be a pointer",
		)
	}

	if value.IsNil() {
		return nil, fmt.Errorf(
			"JSON decode target is nil",
		)
	}

	return value.Type().Elem(), nil
}

// validateExactFields checks object keys against the exact JSON field names
// represented by targetType before the destination can be mutated.
func validateExactFields(
	data []byte,
	targetType reflect.Type,
) error {
	decoder := json.NewDecoder(
		bytes.NewReader(data),
	)
	decoder.UseNumber()

	var value any

	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf(
			"decode JSON structure: %w",
			err,
		)
	}

	return validateExactValue(
		value,
		targetType,
		"$",
	)
}

// validateExactValue recursively checks keys for struct-shaped JSON values.
func validateExactValue(
	value any,
	targetType reflect.Type,
	path string,
) error {
	if value == nil {
		return nil
	}

	for targetType.Kind() == reflect.Pointer {
		if isOpaqueJSONType(targetType) {
			return nil
		}

		targetType = targetType.Elem()
	}

	if isOpaqueJSONType(targetType) {
		return nil
	}

	switch targetType.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}

		fields := jsonStructFields(targetType)

		for _, key := range sortedObjectKeys(object) {
			fieldType, exists := fields[key]
			if !exists {
				return fmt.Errorf(
					"unknown JSON object key %q at %s",
					key,
					path,
				)
			}

			if err := validateExactValue(
				object[key],
				fieldType,
				fmt.Sprintf(
					"%s[%q]",
					path,
					key,
				),
			); err != nil {
				return err
			}
		}

	case reflect.Slice, reflect.Array:
		items, ok := value.([]any)
		if !ok {
			return nil
		}

		for index, item := range items {
			if err := validateExactValue(
				item,
				targetType.Elem(),
				fmt.Sprintf(
					"%s[%d]",
					path,
					index,
				),
			); err != nil {
				return err
			}
		}

	case reflect.Map:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}

		for _, key := range sortedObjectKeys(object) {
			if err := validateExactValue(
				object[key],
				targetType.Elem(),
				fmt.Sprintf(
					"%s[%q]",
					path,
					key,
				),
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// sortedObjectKeys returns deterministic JSON object key order for errors.
func sortedObjectKeys(
	object map[string]any,
) []string {
	keys := make(
		[]string,
		0,
		len(object),
	)

	for key := range object {
		keys = append(
			keys,
			key,
		)
	}

	sort.Strings(keys)

	return keys
}

// isOpaqueJSONType reports whether a type owns its JSON representation and
// must therefore validate that representation itself.
func isOpaqueJSONType(
	targetType reflect.Type,
) bool {
	if targetType == jsonRawMessageType {
		return true
	}

	if targetType.Implements(
		jsonUnmarshalerType,
	) {
		return true
	}

	if targetType.Kind() != reflect.Pointer &&
		reflect.PointerTo(targetType).Implements(
			jsonUnmarshalerType,
		) {
		return true
	}

	return false
}

// jsonFieldCandidate describes one field that may represent a JSON object key.
type jsonFieldCandidate struct {
	fieldType reflect.Type
	depth     int
	tagged    bool
}

// jsonStructFields returns the exact JSON names represented by a struct.
func jsonStructFields(
	targetType reflect.Type,
) map[string]reflect.Type {
	candidates := make(
		map[string][]jsonFieldCandidate,
	)

	collectJSONFieldCandidates(
		targetType,
		0,
		make(map[reflect.Type]bool),
		candidates,
	)

	fields := make(
		map[string]reflect.Type,
		len(candidates),
	)

	for name, group := range candidates {
		minimumDepth := group[0].depth

		for _, candidate := range group[1:] {
			if candidate.depth < minimumDepth {
				minimumDepth = candidate.depth
			}
		}

		taggedAtMinimumDepth := false

		for _, candidate := range group {
			if candidate.depth == minimumDepth &&
				candidate.tagged {
				taggedAtMinimumDepth = true
				break
			}
		}

		selected := make(
			[]jsonFieldCandidate,
			0,
			len(group),
		)

		for _, candidate := range group {
			if candidate.depth != minimumDepth {
				continue
			}

			if taggedAtMinimumDepth &&
				!candidate.tagged {
				continue
			}

			selected = append(
				selected,
				candidate,
			)
		}

		if len(selected) == 1 {
			fields[name] = selected[0].fieldType
		}
	}

	return fields
}

// collectJSONFieldCandidates discovers direct and embedded JSON struct fields.
func collectJSONFieldCandidates(
	targetType reflect.Type,
	depth int,
	visiting map[reflect.Type]bool,
	candidates map[string][]jsonFieldCandidate,
) {
	for targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	if targetType.Kind() != reflect.Struct ||
		visiting[targetType] {
		return
	}

	visiting[targetType] = true
	defer delete(
		visiting,
		targetType,
	)

	for index := 0; index < targetType.NumField(); index++ {
		field := targetType.Field(index)
		fieldType := field.Type

		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		if field.PkgPath != "" &&
			(!field.Anonymous ||
				fieldType.Kind() != reflect.Struct) {
			continue
		}

		tag := field.Tag.Get("json")
		name := tag

		if comma := strings.IndexByte(
			name,
			',',
		); comma >= 0 {
			name = name[:comma]
		}

		if name == "-" {
			continue
		}

		tagged := name != ""

		if field.Anonymous &&
			name == "" &&
			fieldType.Kind() == reflect.Struct &&
			!isOpaqueJSONType(field.Type) {
			collectJSONFieldCandidates(
				field.Type,
				depth+1,
				visiting,
				candidates,
			)

			continue
		}

		if name == "" {
			name = field.Name
		}

		candidates[name] = append(
			candidates[name],
			jsonFieldCandidate{
				fieldType: field.Type,
				depth:     depth,
				tagged:    tagged,
			},
		)
	}
}
