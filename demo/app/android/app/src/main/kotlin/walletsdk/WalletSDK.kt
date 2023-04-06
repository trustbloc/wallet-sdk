package walletsdk

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.mem.ActivityLogger
import walletsdk.openid4ci.OpenID4CI
import walletsdk.openid4vp.OpenID4VP
import dev.trustbloc.wallet.sdk.localkms.Store

class WalletSDK {
    private var kms: KMS? = null
    var didResolver: DIDResolver? = null
    private var crypto: Crypto? = null
    var activityLogger: ActivityLogger? = null


    fun InitSDK(kmsStore: Store) {
        val kms = Localkms.newKMS(kmsStore)

        val opts = ResolverOpts()
        opts.setResolverServerURI("http://localhost:8072/1.0/identifiers")
        didResolver = Resolver(opts)

        crypto = kms.crypto
        activityLogger = ActivityLogger()
        this.kms = kms
    }

    fun createDID(didMethodType: String, didKeyType: String): DIDDocResolution {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val createDIDOpts = CreateOpts()
        createDIDOpts.setKeyType(didKeyType)

        val creatorDID = Creator(kms as KeyWriter)

        return creatorDID.create(didMethodType, createDIDOpts)
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

        val activityLogger = this.activityLogger
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        return OpenID4VP(crypto, didResolver, activityLogger)
    }
}