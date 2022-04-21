package typeexpr

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

var (
	valueComparer = cmp.Comparer(cty.Value.RawEquals)
)

func TestDefaults_Apply(t *testing.T) {
	simpleObject := cty.ObjectWithOptionalAttrs(map[string]cty.Type{
		"a": cty.String,
		"b": cty.Bool,
	}, []string{"b"})
	nestedObject := cty.ObjectWithOptionalAttrs(map[string]cty.Type{
		"c": simpleObject,
		"d": cty.Number,
	}, []string{"c"})

	testCases := map[string]struct {
		defaults *Defaults
		value    cty.Value
		want     cty.Value
	}{
		"no defaults": {
			defaults: &Defaults{
				Type: cty.Map(cty.String),
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("bar"),
			}),
			want: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("bar"),
			}),
		},
		"simple object with defaults applied": {
			defaults: &Defaults{
				Type: simpleObject,
				DefaultValues: map[string]cty.Value{
					"b": cty.True,
				},
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.True,
			}),
		},
		"simple object with optional attributes specified": {
			defaults: &Defaults{
				Type: simpleObject,
				DefaultValues: map[string]cty.Value{
					"b": cty.True,
				},
			},
			value: cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("false"),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"a": cty.StringVal("foo"),
				"b": cty.StringVal("false"),
			}),
		},
		"list of objects with defaults applied": {
			defaults: &Defaults{
				Type: cty.List(simpleObject),
				Children: map[string]*Defaults{
					"": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.True,
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.True,
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("bar"),
					"b": cty.True,
				}),
			}),
		},
		"nested object with defaults applied": {
			defaults: &Defaults{
				Type: nestedObject,
				Children: map[string]*Defaults{
					"c": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.False,
						},
					},
				},
			},
			value: cty.ObjectVal(map[string]cty.Value{
				"c": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				"d": cty.NumberIntVal(5),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"c": cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.False,
				}),
				"d": cty.NumberIntVal(5),
			}),
		},
		"tuple of objects with defaults applied": {
			defaults: &Defaults{
				Type: cty.Tuple([]cty.Type{simpleObject, nestedObject}),
				Children: map[string]*Defaults{
					"0": {
						Type: simpleObject,
						DefaultValues: map[string]cty.Value{
							"b": cty.False,
						},
					},
					"1": {
						Type: nestedObject,
						DefaultValues: map[string]cty.Value{
							"c": cty.ObjectVal(map[string]cty.Value{
								"a": cty.StringVal("default"),
								"b": cty.True,
							}),
						},
					},
				},
			},
			value: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(5),
				}),
			}),
			want: cty.TupleVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.StringVal("foo"),
					"b": cty.False,
				}),
				cty.ObjectVal(map[string]cty.Value{
					"c": cty.ObjectVal(map[string]cty.Value{
						"a": cty.StringVal("default"),
						"b": cty.True,
					}),
					"d": cty.NumberIntVal(5),
				}),
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.defaults.Apply(tc.value)
			if !cmp.Equal(tc.want, got, valueComparer) {
				t.Errorf("wrong result\n%s", cmp.Diff(tc.want, got, valueComparer))
			}
		})
	}
}
