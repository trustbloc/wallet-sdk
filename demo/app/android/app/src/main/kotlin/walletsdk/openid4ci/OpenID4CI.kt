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
import dev.trustbloc.wallet.sdk.verifiable.Credential
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

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

    fun requestCredential(didVerificationMethod: VerificationMethod, otp: String?): String? {
        val opts = RequestCredentialWithPreAuthOpts().setPIN(otp)
        val credsArr = newInteraction.requestCredentialWithPreAuth(didVerificationMethod, opts)

        if (credsArr.length() != 0L) {
            return credsArr.atIndex(0).serialize()
        }

        return null
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
    fun acknowledgeSuccess(): Boolean {
        return newInteraction.acknowledgeSuccess().equals(true)
    }
    fun acknowledgeReject(): Boolean {
        return newInteraction.acknowledgeReject().equals(true)
    }

}