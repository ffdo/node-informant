package data

import (
	"fmt"
	"strconv"
	"strings"
)

// Set returns an error, in case when it is not possible to
// establish the value obtained in accordance with given dotted path.
func set(data interface{}, path string, value interface{}) error {
	parts := strings.Split(path, ".")
	// Normalize path.
	for k, v := range parts {
		if v == "" {
			if k == 0 {
				parts = parts[1:]
			} else {
				return fmt.Errorf("Invalid path %q", path)
			}
		}
	}

	point := &data
	for pos, part := range parts {
		switch c := (*point).(type) {
		case []interface{}:
			if i, error := strconv.ParseInt(part, 10, 0); error == nil {
				// 1. normalize slice capacity
				if int(i) >= cap(c) {
					c = append(c, make([]interface{}, int(i)-cap(c)+1, int(i)-cap(c)+1)...)
				}

				// 2. set value or go further
				if pos+1 == len(parts) {
					c[i] = value
				} else {

					// if exists just pick the pointer
					if va := c[i]; va != nil {
						point = &va
					} else {
						// is next part slice or map?
						if i, err := strconv.ParseInt(parts[pos+1], 10, 0); err == nil {
							va = make([]interface{}, int(i)+1, int(i)+1)
						} else {
							va = make(map[string]interface{})
						}
						c[i] = va
						point = &va
					}

				}

			} else {
				return fmt.Errorf("Invalid list index at %q",
					strings.Join(parts[:pos+1], "."))
			}
		case map[string]interface{}:
			if pos+1 == len(parts) {
				c[part] = value
			} else {
				// if exists just pick the pointer
				if va, ok := c[part]; ok {
					point = &va
				} else {
					// is next part slice or map?
					if i, err := strconv.ParseInt(parts[pos+1], 10, 0); err == nil {
						va = make([]interface{}, int(i)+1, int(i)+1)
					} else {
						va = make(map[string]interface{})
					}
					c[part] = va
					point = &va
				}
			}
		default:
			return fmt.Errorf(
				"Invalid type at %q: expected []interface{} or map[string]interface{}; got %T",
				strings.Join(parts[:pos+1], "."), data)
		}
	}

	return nil
}
