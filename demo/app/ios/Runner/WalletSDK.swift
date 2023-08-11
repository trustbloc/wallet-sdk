//
//  WalletSDK.swift
//  Runner
//
//  Created by Volodymyr Kubiv on 28.02.2023.
//

import Foundation
import Walletsdk

enum WalletSDKError: Error {
    case runtimeError(String)
}

class WalletSDK {
    private var kms:LocalkmsKMS?
    private var didResolver: ApiDIDResolverProtocol?
    private var crypto: ApiCryptoProtocol?
    var activityLogger: MemActivityLogger?



    func initSDK(kmsStore: LocalkmsStoreProtocol, didResolverURI: String) {
        kms = LocalkmsNewKMS(kmsStore, nil)
        
        let opts = DidNewResolverOpts()
        opts!.setResolverServerURI(didResolverURI)
        
        didResolver = DidNewResolver(opts, nil)
        crypto = kms!.getCrypto()
        activityLogger = MemActivityLogger()
    }

    func createDID(didMethodType: String, didKeyType: String) throws -> ApiDIDDocResolution {
        let didCreator = DidNewCreator(self.kms, nil)
        let opts = DidNewCreateOpts()
        opts!.setKeyType(didKeyType)
    
        return try didCreator!.create(didMethodType, opts: opts)
         
    }

    func createOpenID4CIInteraction(requestURI: String) throws -> OpenID4CI {
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        
        activityLogger = MemNewActivityLogger()
 
        return OpenID4CI(requestURI: requestURI, didResolver: didResolver, crypto: crypto, activityLogger: activityLogger! )
    }
    
    func createOpenID4CIWalletInitiatedInteraction(issuerURI: String) throws -> WalletInitiatedOpenID4CI {
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
         
        return WalletInitiatedOpenID4CI(issuerURI: issuerURI, didResolver: didResolver, crypto: crypto)
    }

    func createOpenID4VPInteraction() throws -> OpenID4VP {
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("crypto is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("did resolver is not initialized, call initSDK()")
        }
        guard let activityLogger = self.activityLogger else {
            throw WalletSDKError.runtimeError("activity logger is not initialized, call initSDK()")
        }
        return OpenID4VP(didResolver: didResolver, crypto: crypto, activityLogger: activityLogger)
    }
}
