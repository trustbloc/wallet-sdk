/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4vp

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.credential.CredentialsOpt
import dev.trustbloc.wallet.sdk.openid4vp.Interaction
import dev.trustbloc.wallet.sdk.credential.Inquirer
import dev.trustbloc.wallet.sdk.credential.VerifiablePresentation
import dev.trustbloc.wallet.sdk.api.VerificationMethod
import java.lang.Exception

class OpenID4VP constructor(
        private val keyReader: KeyReader,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val documentLoader: LDDocumentLoader) {

    private var initiatedInteraction: Interaction? = null;
    private var verifiablePresentation: VerifiablePresentation? = null;

    fun processAuthorizationRequest(authorizationRequest: String, storedCredentials: List<String>): List<String> {
        val interaction = Interaction(authorizationRequest, keyReader, crypto, didResolver, documentLoader)

        val query = interaction.getQuery()

        val credArray = VerifiableCredentialsArray()
        for (cred in storedCredentials) {
            credArray.add(VerifiableCredential(cred))
        }

        val credentials = CredentialsOpt(credArray)

        val verifiablePresentation = Inquirer(documentLoader).query(query, credentials)

        val matchedCreds = verifiablePresentation.credentials()

        this.verifiablePresentation = verifiablePresentation
        initiatedInteraction = interaction

        return List<String>(matchedCreds.length().toInt()
        ) { i: Int -> matchedCreds.atIndex(i.toLong()).content }
    }

    fun presentCredential(didVerificationMethod: VerificationMethod) {
        val verifiablePresentation = this.verifiablePresentation
                ?: throw Exception("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")

        initiatedInteraction.presentCredential(verifiablePresentation.content(), didVerificationMethod)
    }
}