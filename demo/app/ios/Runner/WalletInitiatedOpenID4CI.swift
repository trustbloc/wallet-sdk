//
//  WalletInitiatedOpenID4CI.swift
//  Runner
// Copyright Gen Digital Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
        
        self.walletInitiatedInteraction = Openid4ciNewWalletInitiatedInteraction(args, opts, nil)!
    }
    
    func getSupportedCredentials() throws -> Openid4ciSupportedCredentials{
        let supportedCredentials = try walletInitiatedInteraction.supportedCredentials()
        return supportedCredentials
    }
}
