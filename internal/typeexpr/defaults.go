package typeexpr

import (
	"github.com/zclconf/go-cty/cty"
)

// Defaults represents a type tree which may contain default values for
// optional object attributes at any level. This is used to apply nested
// defaults to an input value before converting it to the concrete type.
type Defaults struct {
	// Type of the node for which these defaults apply. This is necessary in
	// order to determine how to inspect the Defaults and Children collections.
	Type cty.Type

	// DefaultValues contains the default values for each object attribute,
	// indexed by attribute name.
	DefaultValues map[string]cty.Value

	// Children is a map of Defaults for elements contained in this type. This
	// only applies to structural and collection types.
	//
	// The map is indexed by string instead of cty.Value because cty.Number
	// instances are non-comparable, due to embedding a *big.Float.
	//
	// Collections have a single element type, which is stored at key "".
	Children map[string]*Defaults
}

// Apply walks the given value, applying specified defaults wherever optional
// attributes are missing. The input and output values may have different
// types, and the result may still require type conversion to the final desired
// type.
//
// This function is permissive and does not report errors, assuming that the
// caller will have better context to report useful type conversion failure
// diagnostics.
func (d *Defaults) Apply(val cty.Value) cty.Value {
	val, err := cty.Transform(val, func(path cty.Path, v cty.Value) (cty.Value, error) {
		// Cannot apply defaults to an unknown value
		if !v.IsKnown() {
			return v, nil
		}

		// Look up the defaults for this path.
		defaults := d.traverse(path)

		// If we have no defaults, nothing to do.
		if len(defaults) == 0 {
			return v, nil
		}

		// Ensure we are working with an object or map
		vt := v.Type()
		if !vt.IsObjectType() && !vt.IsMapType() {
			// Cannot apply defaults because the value type is incompatible.
			// We'll ignore this and let the later conversion stage display a
			// more useful diagnostic.
			return v, nil
		}

		// Apply defaults where attributes are missing, constructing a new
		// value with the same marks.
		v, valMarks := v.Unmark()
		attrs := v.AsValueMap()

		for attr, defaultValue := range defaults {
			if _, ok := attrs[attr]; !ok {
				attrs[attr] = defaultValue
			}
		}

		// We construct an object even if the input value was a map, as the
		// type of an attribute's default value may be incompatible with the
		// map element type.
		return cty.ObjectVal(attrs).WithMarks(valMarks), nil
	})

	// Our transform callback above should never return an error.
	if err != nil {
		panic(err)
	}

	return val
}

func (d *Defaults) traverse(path cty.Path) map[string]cty.Value {
	if len(path) == 0 {
		return d.DefaultValues
	}

	pathStep := path[0]
	switch s := pathStep.(type) {
	case cty.GetAttrStep:
		if d.Type.IsObjectType() {
			if child, ok := d.Children[s.Name]; ok {
				return child.traverse(path[1:])
			}
		}

		return nil
	case cty.IndexStep:
		if d.Type.IsTupleType() {
			// Tuples can have different types for each element, so we look
			// up the defaults based on the index key.
			if child, ok := d.Children[s.Key.AsBigFloat().String()]; ok {
				return child.traverse(path[1:])
			}
		} else if d.Type.IsCollectionType() {
			// Defaults for collection element types are stored with a blank
			// key, so we disregard the index key.
			if child, ok := d.Children[""]; ok {
				return child.traverse(path[1:])
			}
		}
		return nil
	default:
		// At time of writing there are no other path step types.
		return nil
	}
}
