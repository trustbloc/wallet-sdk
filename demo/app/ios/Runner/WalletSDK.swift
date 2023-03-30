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
    private var documentLoader: ApiLDDocumentLoaderProtocol?
    private var crypto: ApiCryptoProtocol?
    var activityLogger: MemActivityLogger?



    func InitSDK(kmsStore: LocalkmsStoreProtocol) {
        kms = LocalkmsNewKMS(kmsStore, nil)
        didResolver = DidNewResolver("http://did-resolver.trustbloc.local:8072/1.0/identifiers", nil)
        crypto = kms?.getCrypto()
        documentLoader = LdNewDocLoader()
        activityLogger = MemNewActivityLogger()
    }

    func createDID(didMethodType: String) throws -> ApiDIDDocResolution {
        let didCreator = DidNewCreatorWithKeyWriter(self.kms, nil)
            let apiCreate = ApiCreateDIDOpts.init()
            if (didMethodType == "jwk"){
                apiCreate.keyType = "ECDSAP384IEEEP1363"
            }
          
            return try didCreator!.create(didMethodType, createDIDOpts: apiCreate)
         
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
        guard let documentLoader = self.documentLoader else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }        
        guard let activityLogger = self.activityLogger else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        
        return OpenID4VP(keyReader: kms, didResolver: didResolver, documentLoader: documentLoader, crypto: crypto, activityLogger: activityLogger)
    }
}
