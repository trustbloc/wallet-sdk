/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4ci

import android.annotation.SuppressLint
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.display.*
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.stderr.MetricsLogger
import dev.trustbloc.wallet.sdk.trustregistry.CredentialOffer
import dev.trustbloc.wallet.sdk.trustregistry.EvaluationResult
import dev.trustbloc.wallet.sdk.trustregistry.IssuanceRequest
import dev.trustbloc.wallet.sdk.trustregistry.Registry
import dev.trustbloc.wallet.sdk.trustregistry.RegistryConfig
import walletsdk.flutter.converters.convertVerifiableCredentialsWithIdArray
import java.util.TreeMap
class OpenID4CI constructor(
        private val requestURI: String,
        private val crypto: Crypto,
        private val didResolver: DIDResolver,
        private val activityLogger: ActivityLogger,
        private val kms: KMS,
) {
    private var newInteraction: IssuerInitiatedInteraction

    init {
        val trace = Otel.newTrace()

        val args = IssuerInitiatedInteractionArgs(requestURI, crypto, didResolver)

        val opts = InteractionOpts()
        opts.addHeader(trace.traceHeader())
        opts.setActivityLogger(activityLogger)
        opts.setMetricsLogger(MetricsLogger())
        opts.enableDIProofChecks(kms)

        newInteraction = IssuerInitiatedInteraction(args, opts)
    }

    fun checkFlow(): String {
        if (newInteraction.authorizationCodeGrantTypeSupported()) {
            return "auth-code-flow"
        }
        if (newInteraction.preAuthorizedCodeGrantTypeSupported()) {
            return "preauth-code-flow"
        }
        return ""
    }

    fun createAuthorizationURL(
            clientID: String,
            redirectURI: String,
            oauthDiscoverableClientURI: String,
            scopes: StringArray
    ): String {
        val opts = CreateAuthorizationURLOpts()

        if (scopes.length() != 0.toLong()) {
            opts.setScopes(scopes)
        }

        if (oauthDiscoverableClientURI != "") {
            opts.useOAuthDiscoverableClientIDScheme()
        }

        return newInteraction.createAuthorizationURL(
                clientID,
                redirectURI,
                opts,
        )
    }

    fun pinRequired(): Boolean {
        if (!newInteraction.preAuthorizedCodeGrantTypeSupported()) {
            return false
        }
        return newInteraction.preAuthorizedCodeGrantParams().pinRequired()
    }

    fun issuerURI(): String {
        return newInteraction.issuerURI()
    }

    fun checkWithTrustRegistry(evaluateIssuanceURL: String): EvaluationResult {
        val issuanceRequest = IssuanceRequest()

        val trustInfo = newInteraction.issuerTrustInfo()
        issuanceRequest.issuerDID = trustInfo.did
        issuanceRequest.issuerDomain = trustInfo.domain

        for (rInd in 0 until trustInfo.offerLength() ) {
            val offer = trustInfo.offerAtIndex(rInd)

            val credentialOffer = CredentialOffer();
            credentialOffer.credentialFormat = offer.credentialFormat
            credentialOffer.credentialType =offer.credentialType
            credentialOffer.clientAttestationRequested = true

            issuanceRequest.addCredentialOffers(credentialOffer)
        }

        val config = RegistryConfig()
        config.evaluateIssuanceURL = evaluateIssuanceURL

        return Registry(config).evaluateIssuance(issuanceRequest)
    }

    fun getCredentialOfferDisplayData(): dev.trustbloc.wallet.sdk.display.Data {
        return Display.resolveCredentialOffer(
                newInteraction.issuerMetadata(),
                newInteraction.offeredCredentialsTypes(), ""
        )
    }

    fun requestCredentials(didVerificationMethod: VerificationMethod, otp: String?,
                           attestationVC: String?, attestationVM: VerificationMethod?): List<HashMap<String, String>> {
        val opts = RequestCredentialWithPreAuthOpts().setPIN(otp)

        if (attestationVC != null && attestationVM != null) {
            opts.setAttestationVC(attestationVM, attestationVC)
        }

        val credsArr = newInteraction.requestCredentialWithPreAuth(didVerificationMethod, opts)

        return convertVerifiableCredentialsWithIdArray(credsArr)
    }

    @SuppressLint("SuspiciousIndentation")
    fun requestCredentialWithAuth(
            didVerificationMethod: VerificationMethod,
            redirectURIWithParams: String
    ): String? {
        var credentials = newInteraction.requestCredentialWithAuth(
                didVerificationMethod,
                redirectURIWithParams,
                null
        )
        return credentials.atIndex(0).serialize();
    }

    fun dynamicRegistrationSupported(): Boolean {
        return newInteraction.dynamicClientRegistrationSupported()
    }

    fun dynamicRegistrationEndpoint(): String {
        return newInteraction.dynamicClientRegistrationEndpoint()
    }

    fun getAuthorizationCodeGrantParams(): AuthorizationCodeGrantParams {
        return newInteraction.authorizationCodeGrantParams()
    }

    fun getIssuerMetadata(): IssuerMetadata {
        return newInteraction.issuerMetadata()
    }

    fun requireAcknowledgment(): Boolean {
        return newInteraction.requireAcknowledgment()
    }

    fun acknowledgeSuccess() {
        val serializedStateResp = newInteraction.acknowledgment().serialize()
        val acknowledgement = Acknowledgment(serializedStateResp)

        val interactionDetailsData = """{"user": "123456"}"""
        acknowledgement.setInteractionDetails(interactionDetailsData)
        return acknowledgement.success()
    }

    fun acknowledgeReject() {
        return newInteraction.acknowledgment().reject()
    }

}