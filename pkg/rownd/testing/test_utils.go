package testing

import (
    "crypto/ed25519"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var (
    // Initialize test key pair
    publicKey, privateKey, _ = ed25519.GenerateKey(nil)
)

// GetKeys returns the test key pair
func GetKeys() (ed25519.PublicKey, ed25519.PrivateKey) {
    return publicKey, privateKey
}

// GenerateTestToken creates a test JWT token for testing
func GenerateTestToken() (string, error) {
    claims := jwt.MapClaims{
        "https://auth.rownd.io/app_user_id": "rownd-test-user-1",
        "iss":                               "dev.rownd.io",
        "aud":                               []string{"app:290167281732813315"},
        "iat":                               time.Now().Unix(),
        "exp":                               time.Now().Add(time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
    return token.SignedString(privateKey)
}

// TestAppConfig provides a mock app configuration for testing
var TestAppConfig = map[string]interface{}{
    "app": map[string]interface{}{
        "name": "Rownd (dev)",
        "id":   "290167281732813315",
        "schema": map[string]interface{}{
            "email": map[string]interface{}{
                "display_name":             "Email",
                "type":                     "string",
                "data_category":            "pii_basic",
                "required":                 false,
                "owned_by":                 "user",
                "user_visible":             true,
                "revoke_after":             "1 month",
                "required_retention":        "none",
                "collection_justification": "This piece of personal data is needed to make your customer experience the best it can be. We do not resell this data.",
                "opt_out_warning":         "By turning off your e-mail, your account will no longer work as designed. You may not be able to log-in, get updates, or reset your password",
},
},
},
}

