package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.openid4ci.Opts
import dev.trustbloc.wallet.sdk.vcparse.Vcparse

class OpenID4CI constructor(
        private val requestURI: String,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val activityLogger: ActivityLogger,
) {
    private var newInteraction: Interaction

    init {
        val args = Args(requestURI, "ClientID", crypto, didResolver)

        val opts = Opts()
        opts.setActivityLogger(activityLogger)

        newInteraction = Interaction(args, opts)
    }

    fun authorize(): AuthorizeResult {
        return newInteraction.authorize()
    }

    fun issuerURI(): String {
        return newInteraction.issuerURI()
    }

    fun requestCredential(didVerificationMethod: VerificationMethod, otp: String?): VerifiableCredential? {
        val credsArr = newInteraction.requestCredentialWithPIN(didVerificationMethod, otp)

        if (credsArr.length() != 0L) {
            return credsArr.atIndex(0)
        }

        return null
    }

    fun serializeDisplayData(issuerURI: String?, vcCredentials: VerifiableCredentialsArray): String? {
        return Display.resolve(vcCredentials, issuerURI, null).serialize()
    }
}