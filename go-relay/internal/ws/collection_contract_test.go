package ws

import "testing"

func TestBoundCollectionUser(t *testing.T) {
	tests := []struct {
		name                 string
		edit                 func(*APIKeyValidation)
		client, master, want string
	}{
		{"bound", func(k *APIKeyValidation) {}, "world", "master", "player"},
		{"master key", func(k *APIKeyValidation) { k.ScopedKeyID = 0 }, "world", "master", ""},
		{"invalid", func(k *APIKeyValidation) { k.Valid = false }, "world", "master", ""},
		{"no scope", func(k *APIKeyValidation) { k.Scopes = nil }, "world", "master", ""},
		{"no user", func(k *APIKeyValidation) { k.ScopedUserID = "" }, "world", "master", ""},
		{"wrong master", func(k *APIKeyValidation) {}, "world", "other", ""},
		{"wrong client", func(k *APIKeyValidation) {}, "other", "master", ""},
		{"mapped", func(k *APIKeyValidation) {
			k.ScopedClientIDs = []string{"world", "second"}
			k.ScopedUserIDs = map[string]string{"world": "mapped"}
		}, "world", "master", "mapped"},
		{"map fails closed", func(k *APIKeyValidation) {
			k.ScopedClientIDs = []string{"world", "second"}
			k.ScopedUserIDs = map[string]string{"second": "mapped"}
		}, "world", "master", ""},
		{"multi client denies", func(k *APIKeyValidation) { k.ScopedClientIDs = []string{"second"} }, "world", "master", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &APIKeyValidation{Valid: true, ScopedKeyID: 1, MasterAPIKey: "master", ScopedClientID: "world", ScopedUserID: "player", Scopes: []string{"entity:read"}}
			tt.edit(key)
			user, err := BoundCollectionUser(key, tt.client, tt.master)
			if user != tt.want || (err == nil) != (tt.want != "") {
				t.Fatalf("got user=%q err=%v", user, err)
			}
		})
	}
	if _, err := BoundCollectionUser(nil, "world", "master"); err == nil {
		t.Fatal("nil validation must fail")
	}
}
