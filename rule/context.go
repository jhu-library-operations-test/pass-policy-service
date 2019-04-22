package rule

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// Known variable names
const (
	SubmissionVariable = "submission" // ${submission}
	HeaderVariable     = "header"     // ${header}
)

// PassEntityFetcher retrieves the JSON-LD content at the given url, and
// un-marshals it into the provided struct.
//
// entityPointer is expected to be a pointer to a struct or map
// (i.e. anything encoding/json can un-marshal into).  An error will be returned
// otherwise.
type PassEntityFetcher interface {
	FetchEntity(url string, entityPointer interface{}) error
}

// resolvedObject contains a parsed JSON object, as well as the source
// URI from where it came
type resolvedObject struct {
	src    string
	object map[string]interface{}
}

// Context establishes a rule evaluation/resolution context.
type Context struct {
	SubmissionURI string
	Headers       map[string][]string
	PassClient    PassEntityFetcher
	values        map[string]interface{} // values that have been already resolved
}

// Resolve resolves a variable of the form ${a.b.c.d}, returning
// a list of strings.
func (c *Context) Resolve(vari string) ([]string, error) {

	c.init()

	v, ok := toVariable(vari)
	if !ok {
		return []string{vari}, nil
	}

	// Resolve each part of the variable (e.g. a, a.b, a.b.c, a.b.c.d)
	for part, ok := v.shift(); ok; part, ok = part.shift() {
		err := c.resolvePart(part)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve variable part %s", part.segmentName)
		}
	}

	// Now materialize the resolved value(s?) into a list of strings
	switch v := c.values[v.fullName].(type) {
	case string:
		return []string{v}, nil
	case resolvedObject:
		return []string{v.src}, nil
	case []string:
		return uniq(v), nil
	case []resolvedObject:
		var vals []string
		for _, val := range v {
			vals = append(vals, val.src)
		}
		return uniq(vals), nil
	case nil:
		return []string{}, nil
	}

	return nil, errors.Errorf("variable %s resolved to a %T instead of a string", v.fullName, c.values[v.fullName])

}

// Set the ${submission} and ${header} values, if not set already
func (c *Context) init() {

	// if the values map is already initialized, we're done
	if c.values != nil {
		return
	}

	c.values = map[string]interface{}{
		SubmissionVariable: c.SubmissionURI,
	}

	headers := make(map[string]interface{}, len(c.Headers))
	for k, v := range c.Headers {
		headers[k] = v
	}
	c.values[HeaderVariable] = resolvedObject{object: headers}
}

// Resolve a variable part (e.g ${x.y} out of ${x.y.z})
func (c *Context) resolvePart(varPart variable) (err error) {

	// If we already have a value, no need to re-resolve it  Likewise, no point in
	// resolving ${} from ${x}
	if _, ok := c.values[varPart.segmentName]; ok || varPart.prev().segmentName == "" {
		return nil
	}

	// We need to resolve the previous segment into a map, (or list of maps) if it isn't already,
	// and extract its value
	switch prevValue := c.values[varPart.prev().segmentName].(type) {

	// ${foo} is a JSON object, or list of JSON objects, or a JSON blobs that had previously been resolved from URIs.
	// So just look up key 'bar' and save those values to ${foo.bar}
	case resolvedObject:
		err = c.extractValue(varPart, prevValue)
	case []resolvedObject:
		err = c.extractValues(varPart, prevValue)

	// ${foo} is a string, or list of strings.  In order to find ${foo.bar},
	// see if each foo is a stringified JSON blob, or an http uri.
	// If it's a blob, parse to a JSON object and save it as a resolvedObject to ${foo.bar}.
	// If it is a URI, dereference it and, if it a JSON blob, parse it to a JSON object
	// and save a resolvedObject containing both the URI and the resulting blob to ${foo.bar}
	case string:
		if err = c.resolveToObject(varPart.prev(), prevValue); err != nil {
			return errors.Wrapf(err, "could not resolve %s to an object in ${%s}", prevValue, varPart.prev().segmentName)
		}
		err = c.resolvePart(varPart)
	case []string:
		if err = c.resolveToObjects(varPart.prev(), prevValue); err != nil {
			return errors.Wrapf(err, "could not resolve all uris in ${%s}", varPart.prev().segmentName)
		}
		err = c.resolvePart(varPart)

	// ${foo} is a list of some sort, but we don't know the type of the items.  If they are URIs, dereference the URIs.
	// If the result is a JSON blob, parse it and look up the value of the key 'bar' in each blob.  Save to ${foo.bar}
	case []interface{}:
		var list []string
		for _, item := range prevValue {
			s, ok := item.(string)
			if !ok {
				return errors.Errorf("expecting list items to be strings, instead got %T", item)
			}
			list = append(list, s)
		}

		if err = c.resolveToObjects(varPart.prev(), list); err != nil {
			return errors.Wrapf(err, "could not resolve all uris in ${%s}", varPart.prev().segmentName)
		}
		err = c.resolvePart(varPart)

	// ${bar} is has no value, so of course ${foo.bar} has no value either
	case nil:
		c.values[varPart.segmentName] = []string{}
		c.values[varPart.segment] = []string{}

	// ${bar} is some unexpected type.
	default:
		return errors.Errorf("%s is %T, cannot parse into an object to extract %s",
			varPart.prev().segmentName,
			prevValue,
			varPart.segmentName,
		)
	}

	return err
}

// Set ${foo.bar} to foo[bar]
func (c *Context) extractValue(v variable, resolved resolvedObject) error {

	val, ok := resolved.object[v.segment]
	if !ok {
		c.values[v.segmentName] = []string{}
		c.values[v.segment] = []string{}
	}

	c.values[v.segmentName] = val
	c.values[v.segment] = val // this is the shortcut ${properties}, instead of ${x.y.properties}

	return nil
}

// Append foo[bar] to ${foo.bar} for each foo
func (c *Context) extractValues(v variable, resolvedList []resolvedObject) error {
	var vals []string

	for _, resolved := range resolvedList {
		if val, ok := resolved.object[v.segment]; ok {

			switch typedVal := val.(type) {
			case string:
				vals = append(vals, typedVal)
			case []interface{}:
				for _, item := range typedVal {
					strval, ok := item.(string)
					if !ok {
						return errors.Errorf("%s is a list of %T, not strings", v.segmentName, item)
					}
					vals = append(vals, strval)
				}
			default:
				return errors.Errorf("%s is a %T, not a string", v.segmentName, val)
			}
		}
	}

	if len(vals) == 0 {
		return errors.Errorf("no values of %s have key %s", v.prev().segmentName, v.segment)
	}

	c.values[v.segmentName] = vals
	c.values[v.segment] = vals // this is the shortcut ${properties}, instead of ${x.y.properties}

	return nil
}

// resolve a string to an object.  This will only work if the string is a
// http URI, or a JSON blob
func (c *Context) resolveToObject(v variable, s string) error {

	resolved := resolvedObject{
		src:    s,
		object: make(map[string]interface{}, 10),
	}

	c.values[v.segmentName] = resolved
	c.values[v.segment] = resolved

	// If it's a URI, try resolving it
	if strings.HasPrefix(s, "http") {
		return c.PassClient.FetchEntity(s, &resolved.object)
	}

	// Otherwise, attempt to decode it as a JSON blob

	return json.Unmarshal([]byte(s), &resolved.object)
}

// resolve each of a list of strins to an object.  This will only work if each string is a
// http URI, or a JSON blob
func (c *Context) resolveToObjects(v variable, vals []string) error {
	var objs []resolvedObject

	for _, s := range vals {
		resolved := resolvedObject{
			src:    s,
			object: make(map[string]interface{}, 10),
		}
		// If it's a URI, try resolving it
		if strings.HasPrefix(s, "http") {
			if err := c.PassClient.FetchEntity(s, &resolved.object); err != nil {
				return errors.Wrapf(err, "error fetching %s", s)
			}
		} else {

			err := json.Unmarshal([]byte(s), &resolved.object)
			if err != nil {
				return errors.Wrap(err, "error parsing json blob")
			}
		}

		objs = append(objs, resolved)
	}

	c.values[v.segmentName] = objs
	c.values[v.segment] = objs

	return nil
}

func uniq(vals []string) []string {
	uniqueVals := []string{}
	encountered := make(map[string]bool, len(vals))

	for _, str := range vals {
		_, ok := encountered[str]
		if !ok {
			uniqueVals = append(uniqueVals, str)
			encountered[str] = true
		}
	}

	return uniqueVals
}
