package gin

import (
    "github.com/gin-gonic/gin"
    "github.com/rownd/go-sdk/pkg/rownd"
)

type AuthOptions struct {
    FetchUserInfo     bool
    ErrOnInvalidToken bool
    ErrOnMissingUser  bool
}

func Authenticate(client *rownd.Client, opts AuthOptions) gin.HandlerFunc {
    return func(c *gin.Context) {
		export async function fetchRowndWellKnownConfig(
			apiUrl: string
		  ): Promise<WellKnownConfig> {
			if (cache.has('oauth-config')) {
			  return cache.get('oauth-config') as WellKnownConfig;
			}
		  
			let resp: WellKnownConfig = await got
			  .get(`${apiUrl}/hub/auth/.well-known/oauth-authorization-server`)
			  .json();
			cache.set('oauth-config', resp);
		  
			return resp;
		  }
		  export async function fetchRowndJwks(
			jwksUrl: string
		  ): Promise<GetKeyFunction<jose.JWSHeaderParameters, jose.FlattenedJWSInput>> {
			if (cache.has('jwks')) {
			  return jose.createLocalJWKSet(cache.get('jwks') as jose.JSONWebKeySet);
			}
		  
			let resp: jose.JSONWebKeySet = await got.get(jwksUrl).json();
			cache.set('jwks', resp);
		  
			return jose.createLocalJWKSet(resp);
		  }
        // Implementation similar to Node.js express middleware
    }
}