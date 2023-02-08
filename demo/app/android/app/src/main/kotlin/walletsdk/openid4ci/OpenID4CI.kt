package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.openid4ci.AuthorizeResult
import dev.trustbloc.wallet.sdk.openid4ci.ClientConfig
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequestOpts
import dev.trustbloc.wallet.sdk.openid4ci.Interaction

class OpenID4CI constructor(
        private val requestURI: String,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
) {
    private var newInteraction: Interaction

    init {
        val cfg = ClientConfig( "ClientID", crypto, didResolver, null)

        newInteraction = Interaction(requestURI, cfg)
    }

    fun authorize(): AuthorizeResult {
        return newInteraction.authorize()
    }


    fun requestCredential(otp: String?, didVerificationMethod: VerificationMethod): VerifiableCredential? {
        val credReq = CredentialRequestOpts(otp)
        val credsArr = newInteraction.requestCredential(credReq, didVerificationMethod)

        if (credsArr.length() != 0L) {
            return credsArr.atIndex(0)
        }

        return null
    }

    fun resolveCredentialDisplay(): String? {
        val resolvedDisplayData = newInteraction.resolveDisplay("")
        return resolvedDisplayData.serialize()
    }
}