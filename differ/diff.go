package differ

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ChangeType classifies the severity of a change.
type ChangeType string

const (
	Breaking    ChangeType = "BREAKING"
	NonBreaking ChangeType = "NON_BREAKING"
	Warning     ChangeType = "WARNING"
)

// Change represents a single detected difference between two specs.
type Change struct {
	Type        ChangeType `json:"type"`
	Method      string     `json:"method"`
	Path        string     `json:"path"`
	Field       string     `json:"field"`
	Description string     `json:"description"`
}

// Diff compares two OpenAPI specs and returns a list of changes.
func Diff(oldSpec, newSpec *openapi3.T) []Change {
	var changes []Change

	oldPaths := flattenPaths(oldSpec)
	newPaths := flattenPaths(newSpec)

	// Check for removed endpoints and changes on existing endpoints.
	for key, oldOp := range oldPaths {
		if newOp, exists := newPaths[key]; exists {
			changes = append(changes, compareOperations(key, oldOp, newOp)...)
		} else {
			method, path := splitKey(key)
			changes = append(changes, Change{
				Type:        Breaking,
				Method:      method,
				Path:        path,
				Description: "Endpoint removed",
			})
		}
	}

	// Check for new endpoints.
	for key := range newPaths {
		if _, exists := oldPaths[key]; !exists {
			method, path := splitKey(key)
			changes = append(changes, Change{
				Type:        NonBreaking,
				Method:      method,
				Path:        path,
				Description: "New endpoint added",
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		order := map[ChangeType]int{Breaking: 0, Warning: 1, NonBreaking: 2}
		if order[changes[i].Type] != order[changes[j].Type] {
			return order[changes[i].Type] < order[changes[j].Type]
		}
		if changes[i].Path != changes[j].Path {
			return changes[i].Path < changes[j].Path
		}
		return changes[i].Method < changes[j].Method
	})

	return changes
}

type pathKey = string // "METHOD /path"

func makeKey(method, path string) pathKey {
	return strings.ToUpper(method) + " " + path
}

func splitKey(key pathKey) (method, path string) {
	parts := strings.SplitN(key, " ", 2)
	return parts[0], parts[1]
}

func flattenPaths(spec *openapi3.T) map[pathKey]*openapi3.Operation {
	result := make(map[pathKey]*openapi3.Operation)
	if spec.Paths == nil {
		return result
	}
	for path, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			result[makeKey(method, path)] = op
		}
	}
	return result
}

func compareOperations(key pathKey, oldOp, newOp *openapi3.Operation) []Change {
	method, path := splitKey(key)
	var changes []Change

	changes = append(changes, compareParameters(method, path, oldOp.Parameters, newOp.Parameters)...)
	changes = append(changes, compareRequestBodies(method, path, oldOp.RequestBody, newOp.RequestBody)...)
	changes = append(changes, compareResponses(method, path, oldOp.Responses, newOp.Responses)...)

	if oldOp.Description != newOp.Description {
		changes = append(changes, Change{
			Type:        NonBreaking,
			Method:      method,
			Path:        path,
			Field:       "description",
			Description: "Description changed",
		})
	}

	return changes
}

func compareParameters(method, path string, oldParams, newParams openapi3.Parameters) []Change {
	var changes []Change

	oldMap := make(map[string]*openapi3.Parameter)
	for _, p := range oldParams {
		param := p.Value
		if param != nil {
			oldMap[param.In+":"+param.Name] = param
		}
	}

	newMap := make(map[string]*openapi3.Parameter)
	for _, p := range newParams {
		param := p.Value
		if param != nil {
			newMap[param.In+":"+param.Name] = param
		}
	}

	// Removed parameters.
	for key, oldParam := range oldMap {
		if _, exists := newMap[key]; !exists {
			changes = append(changes, Change{
				Type:        Breaking,
				Method:      method,
				Path:        path,
				Field:       oldParam.Name,
				Description: fmt.Sprintf("Parameter '%s' removed", oldParam.Name),
			})
		}
	}

	// New or changed parameters.
	for key, newParam := range newMap {
		oldParam, exists := oldMap[key]
		if !exists {
			if newParam.Required {
				changes = append(changes, Change{
					Type:        Breaking,
					Method:      method,
					Path:        path,
					Field:       newParam.Name,
					Description: fmt.Sprintf("Required %s parameter '%s' added", newParam.In, newParam.Name),
				})
			} else {
				changes = append(changes, Change{
					Type:        NonBreaking,
					Method:      method,
					Path:        path,
					Field:       newParam.Name,
					Description: fmt.Sprintf("Optional %s parameter '%s' added", newParam.In, newParam.Name),
				})
			}
			continue
		}

		// Type changed.
		if oldParam.Schema != nil && newParam.Schema != nil &&
			oldParam.Schema.Value != nil && newParam.Schema.Value != nil {
			oldTypes := oldParam.Schema.Value.Type.Slice()
			newTypes := newParam.Schema.Value.Type.Slice()
			if len(oldTypes) > 0 && len(newTypes) > 0 && oldTypes[0] != newTypes[0] {
				changes = append(changes, Change{
					Type:        Breaking,
					Method:      method,
					Path:        path,
					Field:       newParam.Name,
					Description: fmt.Sprintf("Parameter '%s' type changed from '%s' to '%s'",
						newParam.Name, oldTypes[0], newTypes[0]),
				})
			}

			changes = append(changes, compareEnums(method, path, newParam.Name, oldParam.Schema.Value, newParam.Schema.Value)...)
		}
	}

	return changes
}

func compareRequestBodies(method, path string, oldBody, newBody *openapi3.RequestBodyRef) []Change {
	var changes []Change

	if oldBody == nil && newBody == nil {
		return changes
	}

	oldSchema := extractBodySchema(oldBody)
	newSchema := extractBodySchema(newBody)

	if oldSchema == nil && newSchema == nil {
		return changes
	}

	if oldSchema == nil && newSchema != nil {
		changes = append(changes, compareSchemaFields(method, path, &openapi3.Schema{}, newSchema)...)
		return changes
	}
	if oldSchema != nil && newSchema == nil {
		return changes
	}

	changes = append(changes, compareSchemaFields(method, path, oldSchema, newSchema)...)
	return changes
}

func extractBodySchema(body *openapi3.RequestBodyRef) *openapi3.Schema {
	if body == nil || body.Value == nil {
		return nil
	}
	content := body.Value.Content
	if content == nil {
		return nil
	}
	mt := content.Get("application/json")
	if mt == nil {
		// Try first available media type.
		for _, v := range content {
			mt = v
			break
		}
	}
	if mt == nil || mt.Schema == nil || mt.Schema.Value == nil {
		return nil
	}
	return mt.Schema.Value
}

func compareSchemaFields(method, path string, oldSchema, newSchema *openapi3.Schema) []Change {
	var changes []Change

	oldProps := schemaProperties(oldSchema)
	newProps := schemaProperties(newSchema)

	newRequired := make(map[string]bool)
	for _, r := range newSchema.Required {
		newRequired[r] = true
	}

	// Removed fields.
	for name := range oldProps {
		if _, exists := newProps[name]; !exists {
			changes = append(changes, Change{
				Type:        Breaking,
				Method:      method,
				Path:        path,
				Field:       name,
				Description: fmt.Sprintf("Field '%s' removed from request body", name),
			})
		}
	}

	// New or changed fields.
	for name, newProp := range newProps {
		oldProp, exists := oldProps[name]
		if !exists {
			if newRequired[name] {
				changes = append(changes, Change{
					Type:        Breaking,
					Method:      method,
					Path:        path,
					Field:       name,
					Description: fmt.Sprintf("Required field '%s' added to request body", name),
				})
			} else {
				changes = append(changes, Change{
					Type:        NonBreaking,
					Method:      method,
					Path:        path,
					Field:       name,
					Description: fmt.Sprintf("New optional field '%s' added", name),
				})
			}
			continue
		}

		// Type changed.
		if oldProp.Value != nil && newProp.Value != nil {
			oldTypes := oldProp.Value.Type.Slice()
			newTypes := newProp.Value.Type.Slice()
			if len(oldTypes) > 0 && len(newTypes) > 0 && oldTypes[0] != newTypes[0] {
				changes = append(changes, Change{
					Type:        Breaking,
					Method:      method,
					Path:        path,
					Field:       name,
					Description: fmt.Sprintf("Field '%s' type changed from '%s' to '%s'", name, oldTypes[0], newTypes[0]),
				})
			}

			changes = append(changes, compareEnums(method, path, name, oldProp.Value, newProp.Value)...)
		}
	}

	return changes
}

func compareEnums(method, path, fieldName string, oldSchema, newSchema *openapi3.Schema) []Change {
	var changes []Change

	oldEnums := make(map[string]bool)
	for _, v := range oldSchema.Enum {
		oldEnums[fmt.Sprintf("%v", v)] = true
	}

	for _, v := range newSchema.Enum {
		val := fmt.Sprintf("%v", v)
		if !oldEnums[val] {
			changes = append(changes, Change{
				Type:        Warning,
				Method:      method,
				Path:        path,
				Field:       fieldName,
				Description: fmt.Sprintf("New enum value added to '%s'", fieldName),
			})
			break // One warning per field is enough.
		}
	}

	return changes
}

func compareResponses(method, path string, oldResp, newResp *openapi3.Responses) []Change {
	var changes []Change
	if oldResp == nil || newResp == nil {
		return changes
	}

	for status := range oldResp.Map() {
		if newResp.Value(status) == nil {
			changes = append(changes, Change{
				Type:        NonBreaking,
				Method:      method,
				Path:        path,
				Field:       status,
				Description: fmt.Sprintf("Response status '%s' removed", status),
			})
		}
	}

	return changes
}

func schemaProperties(s *openapi3.Schema) map[string]*openapi3.SchemaRef {
	if s == nil || s.Properties == nil {
		return make(map[string]*openapi3.SchemaRef)
	}
	return s.Properties
}
