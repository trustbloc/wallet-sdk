/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4vp

import dev.trustbloc.wallet.sdk.api.ActivityLogger
import dev.trustbloc.wallet.sdk.api.Crypto
import dev.trustbloc.wallet.sdk.api.DIDResolver
import dev.trustbloc.wallet.sdk.credential.Inquirer
import dev.trustbloc.wallet.sdk.credential.InquirerOpts
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirementArray
import dev.trustbloc.wallet.sdk.openid4vp.Args
import dev.trustbloc.wallet.sdk.openid4vp.Interaction
import dev.trustbloc.wallet.sdk.openid4vp.Opts
import dev.trustbloc.wallet.sdk.openid4vp.PresentCredentialOpts
import dev.trustbloc.wallet.sdk.openid4vp.VerifierDisplayData
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.stderr.MetricsLogger
import dev.trustbloc.wallet.sdk.trustregistry.CredentialClaimsToCheck
import dev.trustbloc.wallet.sdk.trustregistry.EvaluationResult
import dev.trustbloc.wallet.sdk.trustregistry.PresentationRequest
import dev.trustbloc.wallet.sdk.trustregistry.Registry
import dev.trustbloc.wallet.sdk.trustregistry.RegistryConfig
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray
import dev.trustbloc.wallet.sdk.api.VerificationMethod
import dev.trustbloc.wallet.sdk.openid4vp.Acknowledgment
import dev.trustbloc.wallet.sdk.openid4vp.CredentialClaimKeys
import dev.trustbloc.wallet.sdk.verifiable.Credential
import java.util.TreeMap
class OpenID4VP constructor(
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val activityLogger: ActivityLogger,
) {

    private var initiatedInteraction: Interaction? = null
    private var vpQueryContent: ByteArray? = null
    private var submissionRequirements: SubmissionRequirementArray? = null

    /**
     * ClientConfig contains various parameters for an OpenID4VP Interaction. ActivityLogger is optional, but if provided then activities will be logged there.
    If not provided, then no activities will be logged.
     * interaction is local variable to intiate Interaction representing a single OpenID4VP interaction between a wallet and a verifier.
     * The methods defined on this object are used to help guide the calling code through the OpenID4VP flow.
     */
    fun startVPInteraction(authorizationRequest: String) {
        val trace = Otel.newTrace()

        val args = Args(authorizationRequest, crypto, didResolver)

        val opts = Opts()
        opts.setActivityLogger(activityLogger)
        opts.addHeader(trace.traceHeader())
        opts.setMetricsLogger(MetricsLogger())

        val interaction = Interaction(args, opts)

        vpQueryContent = interaction.getQuery()
        initiatedInteraction = interaction
    }

    fun getMatchedSubmissionRequirements(storedCredentials: CredentialsArray): SubmissionRequirementArray {
        val vpQueryContent = this.vpQueryContent
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")


        val requirements = Inquirer(InquirerOpts().setDIDResolver(didResolver))
                .getSubmissionRequirements(vpQueryContent, storedCredentials)

        submissionRequirements = requirements
        return requirements
    }

    fun checkWithTrustRegistry(evaluatePresentationURL: String): EvaluationResult {
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")

        val submissionRequirements = this.submissionRequirements
                ?: throw Exception("Before you can call checkWithTrustRegistry, you need call getMatchedSubmissionRequirements first")

        val presentationRequest = PresentationRequest()

        for (rInd in 0 until submissionRequirements.len()) {
            val requirement = submissionRequirements.atIndex(rInd)
            for (dInd in 0 until requirement.descriptorLen()) {
                val descriptor = requirement.descriptorAtIndex(dInd)
                for (credInd in 0 until descriptor.matchedVCs.length()) {
                    val cred = descriptor.matchedVCs.atIndex(credInd)
                    val credentialClaims = presentedClaims(cred);

                    val claimsToCheck = CredentialClaimsToCheck();
                    claimsToCheck.credentialID = cred.id()
                    claimsToCheck.issuerID = cred.issuerID()
                    claimsToCheck.credentialTypes = cred.types()
                    claimsToCheck.expirationDate = cred.expirationDate()
                    claimsToCheck.issuanceDate = cred.issuanceDate()
                    claimsToCheck.credentialClaimKeys = credentialClaims

                    presentationRequest.addCredentialClaims(claimsToCheck)
                }
            }
        }

        val trustInfo = initiatedInteraction.trustInfo()
        presentationRequest.verifierDID = trustInfo.did
        presentationRequest.verifierDomain = trustInfo.domain

        val config = RegistryConfig()
        config.evaluatePresentationURL = evaluatePresentationURL

        return Registry(config).evaluatePresentation(presentationRequest)
    }

    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    fun presentCredential(selectedCredentials: CredentialsArray,
                          customScopes: MutableMap<String, Any>?,
                          didVerificationMethod: VerificationMethod?,
                          attestationVC: String?
                          ) {
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")

        val opts = PresentCredentialOpts()
        if (customScopes != null) {
            for (scope in customScopes) {
                opts?.addScopeClaim(scope.key, scope.value.toString())
            }
        }

        if (attestationVC != null && didVerificationMethod != null) {
            opts.setAttestationVC(didVerificationMethod, attestationVC)
        }

       // var map = TreeMap<String, Any>()
        val interactionDetailsData = """{"user": "123456"}"""
        opts.setInteractionDetails(interactionDetailsData)
        initiatedInteraction.presentCredentialOpts(selectedCredentials, opts)
    }

    fun presentedClaims(credential: Credential): CredentialClaimKeys? {
        return initiatedInteraction?.presentedClaims(credential)
        }
    fun getCustomScope(): ArrayList<String> {
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")

        val customScopes = initiatedInteraction.customScope()
        val customScopesList = ArrayList<String>()
        for (i in 0 until (customScopes.length())) {
            if (customScopes.atIndex(i) != "openid") {
                customScopesList.add(customScopes.atIndex(i))
            }
        }

        return customScopesList
    }

    fun noConsentAcknowledgement(): Acknowledgment {
        val initiatedInteraction = this.initiatedInteraction
            ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")

        return initiatedInteraction.acknowledgment()
    }
    fun getVerifierDisplayData(): VerifierDisplayData {
        val initiatedInteraction = this.initiatedInteraction
                ?: throw Exception("OpenID4VP interaction not properly initialized, call startVPInteraction first")

        return initiatedInteraction.verifierDisplayData()
    }
}