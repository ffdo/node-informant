package data

import "log"

func merge(parent, child interface{}) (interface{}, error) {
	point := &child

	switch c := (*point).(type) {
	case []interface{}:
		if resParent, err := mergeArray(parent.([]interface{}), c); err != nil {
			return nil, err
		} else {
			parent = resParent
		}
	case map[string]interface{}:
		if resParent, err := mergeMap(parent.(map[string]interface{}), child.(map[string]interface{})); err != nil {
			return nil, err
		} else {
			parent = resParent
		}
	case interface{}:
		parent = child
	default:
		log.Printf("This shouldn't happen: Type %v", c)
	}
	return parent, nil
}

func mergeMap(parent, child map[string]interface{}) (map[string]interface{}, error) {
	knownKeys := make([]string, 0, 10)
	for k, v := range child {
		if _, exists := parent[k]; !exists {
			parent[k] = v
		} else {
			knownKeys = append(knownKeys, k)
		}
	}
	var err error
	for _, k := range knownKeys {
		parent[k], err = merge(parent[k], child[k])
		if err != nil {
			return nil, err
		}
	}
	return parent, nil
}

func mergeArray(parent, child []interface{}) ([]interface{}, error) {
	foreignElements := make([]interface{}, 0, 10)
	knownElements := make([]interface{}, 0, 10)
	knownElementsParent := make([]interface{}, 0, 10)
	for _, elem := range child {
		if exists, parentElem := isInArray(parent, elem); !exists {
			foreignElements = append(foreignElements, elem)
		} else {
			knownElements = append(knownElements, elem)
			knownElementsParent = append(knownElementsParent, parentElem)
		}
	}
	parent = append(parent, foreignElements...)
	var err error
	for i, elem := range knownElements {
		parent[i], err = merge(knownElementsParent[i], elem)
		if err != nil {
			return nil, err
		}
	}
	return parent, nil
}

func isInArray(array []interface{}, elem interface{}) (bool, interface{}) {
	for _, arrElem := range array {
		if arrElem == elem {
			return true, arrElem
		}
	}
	return false, nil
}
