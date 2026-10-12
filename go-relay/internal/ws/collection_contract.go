package ws

import (
	"fmt"
	"slices"
)

func IsCollectionContract(kind string) bool {
	return slices.Contains([]string{"actors", "items", "actors-detail", "items-detail", "world"}, kind)
}

// BoundCollectionUser never grants a master key or connection token GM visibility.
func BoundCollectionUser(key *APIKeyValidation, clientID, masterKey string) (string, error) {
	if key == nil || !key.Valid || key.ScopedKeyID <= 0 || masterKey == "" || key.MasterAPIKey != masterKey || !slices.Contains(key.Scopes, "entity:read") {
		return "", fmt.Errorf("CALLER_IDENTITY_REQUIRED")
	}
	if (len(key.ScopedClientIDs) > 0 && !slices.Contains(key.ScopedClientIDs, clientID)) || (len(key.ScopedClientIDs) == 0 && key.ScopedClientID != "" && key.ScopedClientID != clientID) {
		return "", fmt.Errorf("CLIENT_NOT_AUTHORIZED")
	}
	user := key.ScopedUserID
	if len(key.ScopedUserIDs) > 0 {
		user = key.ScopedUserIDs[clientID]
	}
	if user == "" {
		return "", fmt.Errorf("CALLER_IDENTITY_REQUIRED")
	}
	return user, nil
}
