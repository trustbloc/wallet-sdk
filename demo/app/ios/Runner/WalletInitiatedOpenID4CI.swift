/*
 Copyright Gen Digital Inc. All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
 */

import Foundation
import Walletsdk


public class WalletInitiatedOpenID4CI {

    private var didResolver: ApiDIDResolverProtocol
    private var crypto: ApiCryptoProtocol

    private var walletInitiatedInteraction: Openid4ciWalletInitiatedInteraction

    init (issuerURI: String, didResolver: ApiDIDResolverProtocol, crypto: ApiCryptoProtocol) {
        self.didResolver = didResolver
        self.crypto = crypto

        let trace = OtelNewTrace(nil)

        let args = Openid4ciNewWalletInitiatedInteractionArgs(issuerURI, self.crypto, self.didResolver)

        let opts = Openid4ciNewInteractionOpts()
        opts!.add(trace!.traceHeader())
        // opts!.add(ApiHeader("X-Correlation-Id", value: "randomstr"))
        self.walletInitiatedInteraction = Openid4ciNewWalletInitiatedInteraction(args, opts, nil)!
    }

    func getSupportedCredentials() throws -> Openid4ciSupportedCredentials{
        return try walletInitiatedInteraction.issuerMetadata().supportedCredentials()!
    }

    func requestCredentialWithWalletInitiatedFlow(didVerificationMethod: ApiVerificationMethod, redirectURIWithParams: String) throws -> VerifiableCredential {
        let credentials = try walletInitiatedInteraction.requestCredential(didVerificationMethod, redirectURIWithAuthCode: redirectURIWithParams, opts: nil)
        return credentials.atIndex(0)!;
    }

    func createAuthorizationURLWalletInitiatedFlow(scopes: ApiStringArray, credentialFormat: String, credentialTypes: ApiStringArray, clientID: String,
                                                   redirectURI: String, issuerURI: String) throws -> String {
        var createAuthURLError: NSError?

        let opts = Openid4ciNewCreateAuthorizationURLOpts()!.setScopes(scopes)
        opts!.setIssuerState(issuerURI)

        let authorizationLink = walletInitiatedInteraction.createAuthorizationURL(clientID, redirectURI: redirectURI, credentialFormat: credentialFormat, credentialTypes: credentialTypes, opts: opts, error: &createAuthURLError)
        if let actualError = createAuthURLError {
            print("error from create authorization URL Wallet Initiated Flow",  actualError.localizedDescription)
            throw actualError
        }

        return authorizationLink
    }

}
