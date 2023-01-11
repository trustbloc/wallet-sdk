package openid4ci

const (
	module = "OCI"

	NoClientConfigProvidedCode  = module + "0-0001"
	NoClientConfigProvidedError = "NO_CLIENT_CONFIG_PROVIDED"

	ClientConfigNoUserDidProvidedCode  = module + "0-0002"
	ClientConfigNoUserDidProvidedError = "CLIENT_CONFIG_NO_USER_DID_PROVIDED"

	ClientConfigNoClientIDProvidedCode  = module + "0-0003"
	ClientConfigNoClientIDProvidedError = "CLIENT_CONFIG_NO_CLIENT_ID_PROVIDED"

	ClientConfigNoSignerProviderProvidedCode  = module + "0-0004"
	ClientConfigNoSignerProviderProvidedError = "CLIENT_CONFIG_NO_SIGNER_PROVIDER_PROVIDED"

	ClientConfigNoDIDResolverProvidedCode  = module + "0-0005"
	ClientConfigNoDIDResolverProvidedError = "CLIENT_CONFIG_DID_RESOLVER_PROVIDED"

	PreAuthorizedCodeRequiredCode  = module + "0-0001"
	PreAuthorizedCodeRequiredError = "PRE_AUTHORIZED_CODE_REQUIRED"
)
