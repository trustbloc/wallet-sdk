/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.*
import dev.trustbloc.wallet.sdk.didion.Didion
import dev.trustbloc.wallet.sdk.didjwk.Didjwk
import dev.trustbloc.wallet.sdk.didkey.Didkey
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.mem.ActivityLogger
import walletsdk.openid4ci.OpenID4CI
import walletsdk.openid4vp.OpenID4VP
import dev.trustbloc.wallet.sdk.localkms.Store
import walletsdk.openid4ci.WalletInitiatedOpenID4CI

class WalletSDK {
    private var kms: KMS? = null
    var didResolver: DIDResolver? = null
    private var crypto: Crypto? = null
    var activityLogger: ActivityLogger? = null


    fun initSDK(kmsStore: Store, didResolverURI: String) {
        val kms = Localkms.newKMS(kmsStore)

        val opts = ResolverOpts()
        if (didResolverURI != "") {
            opts.setResolverServerURI(didResolverURI)
        }
        didResolver = Resolver(opts)

        crypto = kms.crypto
        activityLogger = ActivityLogger()
        this.kms = kms
    }

    fun createDID(didMethodType: String, didKeyType: String): DIDDocResolution {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val jwk = kms.create(didKeyType)

        println("Created a new key. The key ID is ${jwk.id()}")

        val doc: DIDDocResolution

        if (didMethodType == "key") {
            doc = Didkey.create(jwk)
        } else if (didMethodType == "jwk") {
            doc = Didjwk.create(jwk)
        } else if (didMethodType == "ion") {
            doc = Didion.createLongForm(jwk)
        } else {
            throw java.lang.Exception("DID method type $didMethodType not supported")
        }

        println("Successfully created a new did:${didMethodType} DID. The DID is ${doc.id()}")

        return doc

    }

    fun createOpenID4CIInteraction(requestURI: String): OpenID4CI {
        val didResolver = this.didResolver
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val crypto = this.crypto
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val activityLogger = this.activityLogger
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val kms =
            this.kms ?: throw java.lang.Exception("Local kms is not initialized, call initSDK()")
        return OpenID4CI(
            requestURI,
            crypto,
            didResolver,
            activityLogger,
            kms
        )
    }

    fun createOpenID4CIWalletInitiatedInteraction(issuerURI: String): WalletInitiatedOpenID4CI {
        val didResolver = this.didResolver
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val crypto = this.crypto
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        return WalletInitiatedOpenID4CI(
            issuerURI,
            crypto,
            didResolver,
        )
    }


    fun createOpenID4VPInteraction(): OpenID4VP {
        val crypto = this.crypto
            ?: throw java.lang.Exception("crypto is not initialized, call initSDK()")
        val didResolver = this.didResolver
            ?: throw java.lang.Exception("did resolver is not initialized, call initSDK()")

        val activityLogger = this.activityLogger
            ?: throw java.lang.Exception("activity logger is not initialized, call initSDK()")

        return OpenID4VP(crypto, didResolver, activityLogger)
    }
}