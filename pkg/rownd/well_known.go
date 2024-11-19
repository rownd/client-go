package auth

type WellKnownConfig struct {
    Issuer                                     string   `json:"issuer"`
    TokenEndpoint                              string   `json:"token_endpoint"`
    JwksURI                                    string   `json:"jwks_uri"`
    UserinfoEndpoint                           string   `json:"userinfo_endpoint"`
    ResponseTypesSupported                     []string `json:"response_types_supported"`
    IDTokenSigningAlgValuesSupported          []string `json:"id_token_signing_alg_values_supported"`
    TokenEndpointAuthMethodsSupported         []string `json:"token_endpoint_auth_methods_supported"`
    CodeChallengeMethodsSupported             []string `json:"code_challenge_methods_supported"`
}

func FetchWellKnownConfig(client *http.Client, baseURL string) (*WellKnownConfig, error) {
    // Implementation
}