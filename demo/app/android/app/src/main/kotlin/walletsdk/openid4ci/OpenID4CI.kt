package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.openid4ci.Openid4ci.resolveDisplay
import dev.trustbloc.wallet.sdk.vcparse.Vcparse

class OpenID4CI constructor(
        private val requestURI: String,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val activityLogger: ActivityLogger,
) {
    private var newInteraction: Interaction

    init {
        val cfg = ClientConfig("ClientID", crypto, didResolver, activityLogger)

        newInteraction = Interaction(requestURI, cfg)
    }

    fun authorize(): AuthorizeResult {
        return newInteraction.authorize()
    }

    fun issuerURI(): String {
        return newInteraction.issuerURI()
    }

    fun requestCredential(otp: String?, didVerificationMethod: VerificationMethod): VerifiableCredential? {
        val credReq = CredentialRequestOpts(otp)
        val credsArr = newInteraction.requestCredential(credReq, didVerificationMethod)

        if (credsArr.length() != 0L) {
            return credsArr.atIndex(0)
        }

        return null
    }

    fun resolveCredentialDisplay(issuerURI: String?, vcCredentials: VerifiableCredentialsArray): String? {
        return resolveDisplay(vcCredentials, issuerURI, "").serialize()
    }

    fun getCredID(vcCredentials: ArrayList<String>) : String?{
        val opts = Vcparse.newOpts(true, null)
        val credIds = ArrayList<String>()
        for (cred in vcCredentials) {
            val parsedVC = Vcparse.parse(cred, opts)
            var credID = parsedVC.id()
            credIds.add(credID)
        }
        return credIds[0];
    }
}