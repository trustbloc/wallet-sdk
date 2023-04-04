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



    func InitSDK(kmsStore: LocalkmsStoreProtocol) {
        kms = LocalkmsNewKMS(kmsStore, nil)
        
        let opts = DidNewResolverOpts()
        opts!.setResolverServerURI("http://did-resolver.trustbloc.local:8072/1.0/identifiers")
        
        didResolver = DidNewResolver(opts, nil)
        crypto = kms!.getCrypto()
        activityLogger = MemNewActivityLogger()
    }

    func createDID(didMethodType: String, didKeyType: String) throws -> ApiDIDDocResolution {
        let didCreator = DidNewCreator(self.kms, nil)
        let opts = DidNewCreateOpts()
        opts!.setKeyType(didKeyType)
    
        return try didCreator!.create(didMethodType, opts: opts)
         
    }
    
    func wellKnownConfig(did: String) {
        
    }

    func createOpenID4CIInteraction(requestURI: String) throws -> OpenID4CI {
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }

        guard let activityLogger = self.activityLogger else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
 
        return OpenID4CI(requestURI: requestURI, didResolver: didResolver, crypto: crypto, activityLogger: activityLogger )
    }

    func createOpenID4VPInteraction() throws -> OpenID4VP {
        guard let kms = self.kms else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let activityLogger = self.activityLogger else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        
        return OpenID4VP(keyReader: kms, didResolver: didResolver, crypto: crypto, activityLogger: activityLogger)
    }
}
