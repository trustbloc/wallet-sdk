/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4vp

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.vcparse.*
import dev.trustbloc.wallet.sdk.credential.CredentialsOpt
import dev.trustbloc.wallet.sdk.openid4vp.Interaction
import dev.trustbloc.wallet.sdk.openid4vp.ClientConfig
import dev.trustbloc.wallet.sdk.credential.Inquirer
import dev.trustbloc.wallet.sdk.credential.VerifiablePresentation
import dev.trustbloc.wallet.sdk.api.VerificationMethod
import java.lang.Exception
import  dev.trustbloc.wallet.sdk.credential.SubmissionRequirement

class OpenID4VP constructor(
        private val keyReader: KeyReader,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val documentLoader: LDDocumentLoader,
        private val activityLogger: ActivityLogger,
        ) {

    private var initiatedInteraction: Interaction? = null
    private var verifiablePresentation: VerifiablePresentation? = null

    /**
     * ClientConfig contains various parameters for an OpenID4VP Interaction. ActivityLogger is optional, but if provided then activities will be logged there.
      If not provided, then no activities will be logged.
     * interaction is local variable to intiate Interaction representing a single OpenID4VP interaction between a wallet and a verifier.
     * The methods defined on this object are used to help guide the calling code through the OpenID4VP flow.
     */
    fun processAuthorizationRequest(authorizationRequest: String, storedCredentials: List<String>): List<String> {
        val cfg = ClientConfig(keyReader, crypto, didResolver, documentLoader, activityLogger)
        val interaction = Interaction(authorizationRequest, cfg)

        val query = interaction.getQuery()

        val credArray = VerifiableCredentialsArray()
        for (cred in storedCredentials) {
            val opts = Opts(true, null)

            val parsedCred = Vcparse.parse(cred, opts)
            credArray.add(parsedCred)
        }

        val credentials = CredentialsOpt(credArray)

        val verifiablePresentation = Inquirer(documentLoader).query(query, credentials)

        val matchedCreds = verifiablePresentation.credentials()

        this.verifiablePresentation = verifiablePresentation
        initiatedInteraction = interaction

        return List<String>(matchedCreds.length().toInt()
        ) { i: Int -> matchedCreds.atIndex(i.toLong()).serialize() }
    }
    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    fun presentCredential(didVerificationMethod: VerificationMethod) {
        val verifiablePresentation = this.verifiablePresentation
                ?: throw Exception("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")

        initiatedInteraction.presentCredential(verifiablePresentation.content(), didVerificationMethod)
    }
}