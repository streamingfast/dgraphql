package static

import (
	"encoding/json"

	"go.uber.org/zap"
)

type GraphqlDocument string

func (g GraphqlDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(g))
}

type GraphqlVariablesByNetwork map[string]json.RawMessage

func (v GraphqlVariablesByNetwork) MarshalJSON() ([]byte, error) {
	for network, rawJSON := range v {
		zlog.Debug("looking if value is a symlink to another network in the list", zap.String("network", network))
		if _, isSymlinkToOtherNetwork := v[string(rawJSON)]; isSymlinkToOtherNetwork {
			zlog.Debug("network is a symlink to another network, resolve it here", zap.String("network", network), zap.String("points_to", string(rawJSON)))
			v[network] = v[string(rawJSON)]
		}
	}

	return json.Marshal(map[string]json.RawMessage(v))
}

// GraphqlExample represents an history example that should be pre-filled in the History
// tab of the GraphiQL web interface.
type GraphqlExample struct {
	Label string `json:"label"`
	// Document represents the GraphQL query to perform, serialized as `query` because it's what
	// GraphiQL uses under the hood, so JSON-wise, we respect the name.
	Document GraphqlDocument `json:"query"`

	// Variables represents the GraphQL variables that should be used to populated the variables box
	// in GraphiQL. The `json.RawMessage` is used so we keep control of the order of variables. The
	// pretty-printing is done client client side by our helper script
	Variables GraphqlVariablesByNetwork `json:"variables"`
}
