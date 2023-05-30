package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.openid4ci.Opts
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.verifiable.Credential
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

class OpenID4CI constructor(
        private val requestURI: String,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val activityLogger: ActivityLogger,
) {
    private var newInteraction: Interaction

    init {
        val trace = Otel.newTrace()

        val args = Args(requestURI, crypto, didResolver)

        val opts = Opts()
        opts.addHeader(trace.traceHeader())
        opts.setActivityLogger(activityLogger)

        newInteraction = Interaction(args, opts)
    }

    fun checkFlow(): String {
        var issuerCapabilities = newInteraction.issuerCapabilities()
        if (issuerCapabilities.authorizationCodeGrantTypeSupported()) {
            return "auth-code-flow"
        }
        if (issuerCapabilities.preAuthorizedCodeGrantTypeSupported()){
            return "preauth-code-flow"
        }
        return ""
    }

    fun createAuthorizationURLWithScopes(scopes:ArrayList<String>, clientID: String, redirectURI: String): String {
        var issuerCapabilities = newInteraction.issuerCapabilities()
        if (!issuerCapabilities.authorizationCodeGrantTypeSupported()) {
            return "Not implemented"
        }
        val scopesArr = StringArray();
        for (scope in scopes) {
            scopesArr.append(scope);
        }
        return newInteraction.createAuthorizationURLWithScopes(
            clientID,
            redirectURI,
            scopesArr
        )
    }

    fun createAuthorizationURL(clientID: String, redirectURI: String): String {
      return  newInteraction.createAuthorizationURL(
          clientID,
          redirectURI
      )
    }

    fun pinRequired(): Boolean {
        var issuerCapabilities = newInteraction.issuerCapabilities()
        if (!issuerCapabilities.preAuthorizedCodeGrantTypeSupported()) {
            return false
        }
        return  newInteraction.issuerCapabilities().preAuthorizedCodeGrantParams().pinRequired()
    }

    fun issuerURI(): String {
        return newInteraction.issuerURI()
    }

    fun requestCredential(didVerificationMethod: VerificationMethod, otp: String?): Credential? {
        val credsArr = newInteraction.requestCredentialWithPIN(didVerificationMethod, otp)

        if (credsArr.length() != 0L) {
            return credsArr.atIndex(0)
        }

        return null
    }

    fun requestCredentialWithAuth(didVerificationMethod: VerificationMethod, redirectURIWithParams: String) : Credential? {
        var credentials = newInteraction.requestCredentialWithAuth(didVerificationMethod, redirectURIWithParams)
            return credentials.atIndex(0);
    }

    fun serializeDisplayData(issuerURI: String?, vcCredentials: CredentialsArray): String? {
        return Display.resolve(vcCredentials, issuerURI, null).serialize()
    }
}