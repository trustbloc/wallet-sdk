/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4vp

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.openid4vp.Interaction
import dev.trustbloc.wallet.sdk.openid4vp.ClientConfig
import dev.trustbloc.wallet.sdk.credential.*
import java.lang.Exception

class OpenID4VP constructor(
        private val keyReader: KeyReader,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val documentLoader: LDDocumentLoader,
        private val activityLogger: ActivityLogger,
) {

    private var initiatedInteraction: Interaction? = null
    private var vpQueryContent: ByteArray? = null

    /**
     * ClientConfig contains various parameters for an OpenID4VP Interaction. ActivityLogger is optional, but if provided then activities will be logged there.
      If not provided, then no activities will be logged.
     * interaction is local variable to intiate Interaction representing a single OpenID4VP interaction between a wallet and a verifier.
     * The methods defined on this object are used to help guide the calling code through the OpenID4VP flow.
     */
    fun startVPInteraction(authorizationRequest: String) {
        val cfg = ClientConfig(keyReader, crypto, didResolver, documentLoader, activityLogger)
        val interaction = Interaction(authorizationRequest, cfg)

        vpQueryContent = interaction.getQuery()
        initiatedInteraction = interaction
    }

    fun getMatchedSubmissionRequirements(storedCredentials: VerifiableCredentialsArray): SubmissionRequirementArray {
        val vpQueryContent = this.vpQueryContent
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")


        return Inquirer(documentLoader)
                .getSubmissionRequirements(vpQueryContent, CredentialsOpt(storedCredentials))
    }
    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    fun presentCredential(selectedCredentials: VerifiableCredentialsArray) {
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")

        initiatedInteraction.presentCredential(selectedCredentials)
    }

    private fun createVerifiablePresentation(selectedCredentials: VerifiableCredentialsArray): VerifiablePresentation {
        val vpQueryContent = this.vpQueryContent
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")


        return Inquirer(documentLoader).query(vpQueryContent, CredentialsOpt(selectedCredentials))
    }
}