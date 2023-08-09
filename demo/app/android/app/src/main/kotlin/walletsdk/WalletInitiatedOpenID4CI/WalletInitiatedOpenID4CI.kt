package walletsdk.openid4ci

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.CreateOpts
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.did.ResolverOpts
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.Store
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.verifiable.Credential
import walletsdk.openid4vp.OpenID4VP


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