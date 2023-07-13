package walletsdk.openid4ci

import android.annotation.SuppressLint
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.stderr.MetricsLogger
import dev.trustbloc.wallet.sdk.verifiable.Credential
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

class OpenID4CI constructor(
        private val requestURI: String,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val activityLogger: ActivityLogger,
) {
    private var newInteraction: IssuerInitiatedInteraction

    init {
        val trace = Otel.newTrace()

        val args = IssuerInitiatedInteractionArgs(requestURI, crypto, didResolver)

        val opts = InteractionOpts()
        opts.addHeader(trace.traceHeader())
        opts.setActivityLogger(activityLogger)
        opts.setMetricsLogger(MetricsLogger())

        newInteraction = IssuerInitiatedInteraction(args, opts)
    }

    fun checkFlow(): String {
        if (newInteraction.authorizationCodeGrantTypeSupported()) {
            return "auth-code-flow"
        }
        if (newInteraction.preAuthorizedCodeGrantTypeSupported()){
            return "preauth-code-flow"
        }
        return ""
    }

    fun createAuthorizationURLWithScopes(scopes: StringArray, clientID: String, redirectURI: String): String {
        if (!newInteraction.authorizationCodeGrantTypeSupported()) {
            return "Not implemented"
        }

        val opts = CreateAuthorizationURLOpts().setScopes(scopes)

        return newInteraction.createAuthorizationURL(
            clientID,
            redirectURI,
            opts
        )
    }

    fun createAuthorizationURL(clientID: String, redirectURI: String): String {
      return  newInteraction.createAuthorizationURL(
          clientID,
          redirectURI,
          null,
      )
    }

    fun pinRequired(): Boolean {
        if (!newInteraction.preAuthorizedCodeGrantTypeSupported()) {
            return false
        }
        return  newInteraction.preAuthorizedCodeGrantParams().pinRequired()
    }

    fun issuerURI(): String {
        return newInteraction.issuerURI()
    }

    fun requestCredential(didVerificationMethod: VerificationMethod, otp: String?): Credential? {
        val opts = RequestCredentialWithPreAuthOpts().setPIN(otp)
        val credsArr = newInteraction.requestCredentialWithPreAuth(didVerificationMethod, opts)

        if (credsArr.length() != 0L) {
            return credsArr.atIndex(0)
        }

        return null
    }

    @SuppressLint("SuspiciousIndentation")
    fun requestCredentialWithAuth(didVerificationMethod: VerificationMethod, redirectURIWithParams: String) : Credential? {
        var credentials = newInteraction.requestCredentialWithAuth(didVerificationMethod, redirectURIWithParams, null)
            return credentials.atIndex(0);
    }

    fun serializeDisplayData(issuerURI: String?, vcCredentials: CredentialsArray): String? {
        return Display.resolve(vcCredentials, issuerURI, null).serialize()
    }

    fun dynamicRegistrationSupported(): Boolean {
        return newInteraction.dynamicClientRegistrationSupported()
    }

    fun dynamicRegistrationEndpoint(): String {
        return newInteraction.dynamicClientRegistrationEndpoint()
    }

    fun getAuthorizationCodeGrantParams(): AuthorizationCodeGrantParams {
        return newInteraction.authorizationCodeGrantParams()
    }

}