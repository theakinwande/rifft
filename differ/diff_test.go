package differ

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func makeSpec(paths map[string]*openapi3.PathItem) *openapi3.T {
	p := &openapi3.Paths{}
	for k, v := range paths {
		p.Set(k, v)
	}
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0.0"},
		Paths:   p,
	}
}

func TestDiff_EndpointRemoved(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
		"/orders": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == Breaking && c.Path == "/orders" && c.Description == "Endpoint removed" {
			found = true
		}
	}
	if !found {
		t.Error("expected BREAKING change for removed /orders endpoint")
	}
}

func TestDiff_EndpointAdded(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
		"/products": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == NonBreaking && c.Path == "/products" && c.Description == "New endpoint added" {
			found = true
		}
	}
	if !found {
		t.Error("expected NON_BREAKING change for new /products endpoint")
	}
}

func TestDiff_RequiredParamAdded(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{
				Parameters: openapi3.Parameters{},
				Responses:  &openapi3.Responses{},
			},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{
						Value: &openapi3.Parameter{
							Name:     "page",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
							},
						},
					},
				},
				Responses: &openapi3.Responses{},
			},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == Breaking && c.Field == "page" {
			found = true
		}
	}
	if !found {
		t.Error("expected BREAKING change for required parameter 'page' added")
	}
}

func TestDiff_FieldTypeChanged(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/orders": {
			Post: &openapi3.Operation{
				RequestBody: &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"quantity"},
										Properties: openapi3.Schemas{
											"quantity": &openapi3.SchemaRef{
												Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
											},
										},
									},
								},
							},
						},
					},
				},
				Responses: &openapi3.Responses{},
			},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/orders": {
			Post: &openapi3.Operation{
				RequestBody: &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     &openapi3.Types{"object"},
										Required: []string{"quantity"},
										Properties: openapi3.Schemas{
											"quantity": &openapi3.SchemaRef{
												Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
											},
										},
									},
								},
							},
						},
					},
				},
				Responses: &openapi3.Responses{},
			},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == Breaking && c.Field == "quantity" {
			found = true
		}
	}
	if !found {
		t.Error("expected BREAKING change for field type change on 'quantity'")
	}
}

func TestDiff_EnumValueAdded(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/orders": {
			Get: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{
						Value: &openapi3.Parameter{
							Name: "status",
							In:   "query",
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
									Enum: []interface{}{"pending", "shipped"},
								},
							},
						},
					},
				},
				Responses: &openapi3.Responses{},
			},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/orders": {
			Get: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{
						Value: &openapi3.Parameter{
							Name: "status",
							In:   "query",
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
									Enum: []interface{}{"pending", "shipped", "delivered"},
								},
							},
						},
					},
				},
				Responses: &openapi3.Responses{},
			},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == Warning && c.Field == "status" {
			found = true
		}
	}
	if !found {
		t.Error("expected WARNING change for new enum value on 'status'")
	}
}

func TestDiff_DescriptionChanged(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{
				Description: "old description",
				Responses:   &openapi3.Responses{},
			},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{
				Description: "new description",
				Responses:   &openapi3.Responses{},
			},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == NonBreaking && c.Description == "Description changed" {
			found = true
		}
	}
	if !found {
		t.Error("expected NON_BREAKING change for description change")
	}
}

func TestDiff_NoChanges(t *testing.T) {
	spec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{Responses: &openapi3.Responses{}},
		},
	})

	changes := Diff(spec, spec)
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestDiff_ParameterRemoved(t *testing.T) {
	oldSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{
						Value: &openapi3.Parameter{
							Name: "limit",
							In:   "query",
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
							},
						},
					},
				},
				Responses: &openapi3.Responses{},
			},
		},
	})
	newSpec := makeSpec(map[string]*openapi3.PathItem{
		"/users": {
			Get: &openapi3.Operation{
				Parameters: openapi3.Parameters{},
				Responses:  &openapi3.Responses{},
			},
		},
	})

	changes := Diff(oldSpec, newSpec)
	found := false
	for _, c := range changes {
		if c.Type == Breaking && c.Field == "limit" && c.Description == "Parameter 'limit' removed" {
			found = true
		}
	}
	if !found {
		t.Error("expected BREAKING change for parameter 'limit' removed")
	}
}
