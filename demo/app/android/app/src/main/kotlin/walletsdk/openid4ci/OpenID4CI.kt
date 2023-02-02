package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.DIDJWTSignerCreator
import dev.trustbloc.wallet.sdk.api.DIDResolver
import dev.trustbloc.wallet.sdk.api.VerifiableCredential
import dev.trustbloc.wallet.sdk.openid4ci.AuthorizeResult
import dev.trustbloc.wallet.sdk.openid4ci.ClientConfig
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequestOpts
import dev.trustbloc.wallet.sdk.openid4ci.Interaction

class OpenID4CI constructor(
        private val requestURI: String,
        private val userDID: String,
        private val didJWTSignerCreator: DIDJWTSignerCreator,
        private val didResolver: DIDResolver,
) {
    private var newInteraction: Interaction

    init {
        val cfg = ClientConfig(userDID, "ClientID", didJWTSignerCreator, didResolver, null)

        println("didJWTSignerCreator")
        println(didJWTSignerCreator)

        println("cfg.signerCreator")
        println(cfg.signerCreator)

        newInteraction = Interaction(requestURI, cfg)
    }

    fun authorize(): AuthorizeResult {
        return newInteraction.authorize()
    }


    fun requestCredential(otp: String?): VerifiableCredential? {
        val credReq = CredentialRequestOpts(otp)
        val credsArr = newInteraction.requestCredential(credReq)

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