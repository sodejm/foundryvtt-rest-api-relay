package handler

import (
	"math"
	"net/http"
	"regexp"

	"github.com/ThreeHats/foundryvtt-rest-api-relay/go-relay/internal/handler/helpers"
)

var documentID = regexp.MustCompile(`^[A-Za-z0-9]{16}$`)

func collectionRoute(kind string, detail bool) helpers.APIRouteConfig {
	cfg := helpers.APIRouteConfig{Type: kind, RequireBoundUser: true, OptionalParams: []helpers.ParamDef{clientIDParam()}}
	if kind == "world" {
		return cfg
	}
	if detail {
		cfg.RequiredParams = []helpers.ParamDef{{Name: "id", From: []helpers.ParamSource{helpers.FromParams}, Type: helpers.TypeString, Required: true}}
	} else {
		for _, name := range []string{"page", "limit"} {
			cfg.OptionalParams = append(cfg.OptionalParams, helpers.ParamDef{Name: name, From: []helpers.ParamSource{helpers.FromQuery}, Type: helpers.TypeNumber})
		}
		for _, name := range []string{"query", "type", "snapshotId"} {
			cfg.OptionalParams = append(cfg.OptionalParams, helpers.ParamDef{Name: name, From: []helpers.ParamSource{helpers.FromQuery}, Type: helpers.TypeString})
		}
	}
	cfg.ValidateParams = func(p helpers.Params, _ *http.Request) (map[string]interface{}, bool) {
		invalid := false
		if detail {
			invalid = !documentID.MatchString(p.GetString("id"))
		}
		for name, maximum := range map[string]float64{"page": 9007199254740991, "limit": 100} {
			if p.Has(name) {
				v := p.GetFloat(name)
				invalid = invalid || math.IsNaN(v) || math.IsInf(v, 0) || v < 1 || v > maximum || v != math.Floor(v)
			}
		}
		invalid = invalid || len(p.GetString("query")) > 1024 || len(p.GetString("type")) > 128 || len(p.GetString("snapshotId")) > 1024
		if invalid {
			return map[string]interface{}{"error": "INVALID_COLLECTION_QUERY"}, true
		}
		return nil, false
	}
	cfg.BuildPayload = func(p helpers.Params, _ *http.Request) map[string]interface{} {
		payload := make(map[string]interface{})
		for k, v := range p {
			if k != "clientId" && k != "type" {
				payload[k] = v
			}
		}
		if !detail {
			payload["filterType"] = p.GetString("type")
			if !p.Has("page") {
				payload["page"] = 1
			}
			if !p.Has("limit") {
				payload["limit"] = 100
			}
		}
		return payload
	}
	return cfg
}

func actorsCollection() helpers.APIRouteConfig { return collectionRoute("actors", false) }
func itemsCollection() helpers.APIRouteConfig  { return collectionRoute("items", false) }
func actorsDetail() helpers.APIRouteConfig     { return collectionRoute("actors-detail", true) }
func itemsDetail() helpers.APIRouteConfig      { return collectionRoute("items-detail", true) }
func worldIdentity() helpers.APIRouteConfig    { return collectionRoute("world", false) }
