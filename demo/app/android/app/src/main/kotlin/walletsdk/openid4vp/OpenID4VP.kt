/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4vp

import dev.trustbloc.wallet.sdk.api.Crypto
import dev.trustbloc.wallet.sdk.api.DIDResolver
import dev.trustbloc.wallet.sdk.api.KeyReader
import dev.trustbloc.wallet.sdk.api.LDDocumentLoader
import dev.trustbloc.wallet.sdk.credentialquery.Credentials
import dev.trustbloc.wallet.sdk.openid4vp.Interaction;
import dev.trustbloc.wallet.sdk.credentialquery.Query;
import walletsdk.util.createJsonArray
import java.lang.Exception

class OpenID4VP constructor(
        private val keyReader: KeyReader,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val documentLoader: LDDocumentLoader) {

    private var initiatedInteraction: Interaction? = null;
    private var verifiablePresentation: ByteArray? = null;

    fun processAuthorizationRequest(authorizationRequest: String, storedCredentials: List<String>) {
        val interaction = Interaction(authorizationRequest, keyReader, crypto, didResolver, documentLoader)

        val query = interaction.getQuery()

        val credentials = Credentials();
        credentials.vCs = createJsonArray(storedCredentials)

        verifiablePresentation = Query(documentLoader).query(query, credentials)

        initiatedInteraction = interaction

        // TODO: add go api to read credentials display information from verifiable.Presentation
    }

    fun presentCredential(signingKeyId: String) {
        val verifiablePresentation = this.verifiablePresentation
                ?: throw Exception("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")

        initiatedInteraction.presentCredential(verifiablePresentation, signingKeyId)
    }
}