/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.Crypto
import dev.trustbloc.wallet.sdk.api.DIDResolver
import dev.trustbloc.wallet.sdk.api.StringArray
import dev.trustbloc.wallet.sdk.api.VerificationMethod
import dev.trustbloc.wallet.sdk.openid4ci.CreateAuthorizationURLOpts
import dev.trustbloc.wallet.sdk.openid4ci.InteractionOpts
import dev.trustbloc.wallet.sdk.openid4ci.SupportedCredentials
import dev.trustbloc.wallet.sdk.openid4ci.WalletInitiatedInteraction
import dev.trustbloc.wallet.sdk.openid4ci.WalletInitiatedInteractionArgs
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.verifiable.Credential


class WalletInitiatedOpenID4CI constructor(
    private val issuerURI: String,
    private val crypto: Crypto,
    private val didResolver: DIDResolver,
) {
    private var walletInitiatedInteraction: WalletInitiatedInteraction

    init {
        val trace = Otel.newTrace()

        val args = WalletInitiatedInteractionArgs(issuerURI, crypto, didResolver)

        val opts = InteractionOpts()
        opts.addHeader(trace.traceHeader())
        opts.setMetricsLogger(dev.trustbloc.wallet.sdk.stderr.MetricsLogger())

        walletInitiatedInteraction = WalletInitiatedInteraction(args, opts)
    }

    fun getSupportedCredentials(): SupportedCredentials {
        return walletInitiatedInteraction.issuerMetadata().supportedCredentials()
    }

    fun requestCredentialWithWalletInitiatedFlow(didVerificationMethod: VerificationMethod, redirectURIWithParams: String): Credential {
        var credentials =  walletInitiatedInteraction.requestCredential(didVerificationMethod, redirectURIWithParams, null)
            return credentials.atIndex(0)
        }

    fun createAuthorizationURLWalletInitiatedFlow(scopes: StringArray, credentialFormat: String, credentialTypes: StringArray, clientID: String,
    redirectURI: String, issuerURI: String): String{
        var opts = CreateAuthorizationURLOpts().setScopes(scopes)
        opts.setIssuerState(issuerURI)

        var authorizationLink = walletInitiatedInteraction.createAuthorizationURL(clientID, redirectURI, credentialFormat, credentialTypes, opts)
        return authorizationLink
    }
}