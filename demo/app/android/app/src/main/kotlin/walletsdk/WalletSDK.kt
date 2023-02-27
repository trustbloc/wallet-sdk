package walletsdk

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirementArray
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.ld.DocLoader
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.mem.ActivityLogger
import walletsdk.openid4ci.OpenID4CI
import walletsdk.openid4vp.OpenID4VP
import dev.trustbloc.wallet.sdk.localkms.Store
import dev.trustbloc.wallet.sdk.openid4ci.AuthorizeResult

class WalletSDK {
    private var kms: KMS? = null
    private var didResolver: DIDResolver? = null
    private var documentLoader: LDDocumentLoader? = null
    private var crypto: Crypto? = null
    var activityLogger: ActivityLogger? = null


    fun InitSDK(kmsStore: Store) {
        val kms = Localkms.newKMS(kmsStore)
        didResolver = Resolver("http://localhost:8072/1.0/identifiers")
        crypto = kms.crypto
        documentLoader = DocLoader()
        activityLogger = ActivityLogger()
        this.kms = kms
    }

    fun createDID(didMethodType: String): DIDDocResolution {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val creatorDID = Creator(kms as KeyWriter)

        return creatorDID.create(didMethodType, CreateDIDOpts())
    }

    fun createOpenID4CIInteraction(requestURI: String) : OpenID4CI {
        val didResolver = this.didResolver
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val crypto = this.crypto
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val activityLogger = this.activityLogger
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        return OpenID4CI(
                requestURI,
                crypto,
                didResolver,
                activityLogger
        )
    }

    fun createOpenID4VPInteraction(): OpenID4VP {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")
        val crypto = this.crypto
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")
        val didResolver = this.didResolver
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")
        val documentLoader = this.documentLoader
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val activityLogger = this.activityLogger
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")


        return OpenID4VP(kms, crypto, didResolver, documentLoader, activityLogger)
    }
}