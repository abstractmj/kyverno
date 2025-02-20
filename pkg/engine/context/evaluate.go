package context

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/kyverno/kyverno/pkg/engine/jmespath"
)

// Query the JSON context with JMESPATH search path
func (ctx *context) Query(query string) (interface{}, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("invalid query (nil)")
	}
	// compile the query
	queryPath, err := jmespath.New(query)
	if err != nil {
		logger.Error(err, "incorrect query", "query", query)
		return nil, fmt.Errorf("incorrect query %s: %v", query, err)
	}
	// search
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()
	var data interface{}
	if err := json.Unmarshal(ctx.jsonRaw, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}
	result, err := queryPath.Search(data)
	if err != nil {
		return nil, fmt.Errorf("JMESPath query failed: %w", err)
	}
	return result, nil
}

func (ctx *context) HasChanged(jmespath string) (bool, error) {
	objData, err := ctx.Query("request.object." + jmespath)
	if err != nil {
		return false, fmt.Errorf("failed to query request.object: %w", err)
	}
	if objData == nil {
		return false, fmt.Errorf("request.object.%s not found", jmespath)
	}
	oldObjData, err := ctx.Query("request.oldObject." + jmespath)
	if err != nil {
		return false, fmt.Errorf("failed to query request.object: %w", err)
	}
	if oldObjData == nil {
		return false, fmt.Errorf("request.oldObject.%s not found", jmespath)
	}
	return !reflect.DeepEqual(objData, oldObjData), nil
}
