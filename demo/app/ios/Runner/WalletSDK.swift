/*
 Copyright Gen Digital Inc. All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
 */

import Foundation
import Walletsdk

enum WalletSDKError: Error {
    case runtimeError(String)
}

class WalletSDK {
    private var kms:LocalkmsKMS?
    public var didResolver: ApiDIDResolverProtocol?
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
        let jwk = try self.kms!.create(didKeyType)

        print("Created a new key. The key ID is \(jwk.id_())")

        var didCreateError: NSError?
        var doc: ApiDIDDocResolution?

        if didMethodType == "key" {
            doc = DidkeyCreate(jwk, &didCreateError)
        } else if didMethodType == "jwk" {
            doc = DidjwkCreate(jwk, &didCreateError)
        } else if didMethodType == "ion" {
            doc = DidionCreateLongForm(jwk, &didCreateError)
        } else {
            throw WalletSDKError.runtimeError("DID method type \(didMethodType) not supported")
        }

        if let didCreateError = didCreateError {
            throw didCreateError
        }

        var docIDError: NSError?

        let docID = doc!.id_(&docIDError)

        if let docIDError = docIDError {
            throw docIDError
        }

        print("Successfully created a new did:\(didMethodType) DID. The DID is \(docID)")

        return doc!
    }

    func createOpenID4CIInteraction(requestURI: String, correlationID: String) throws -> OpenID4CI {
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("SDK is not initialized, call initSDK()")
        }

        guard let localKMS = self.kms else {
            throw WalletSDKError.runtimeError("Local kms is not initialized, call initSDK()")
        }

        activityLogger = MemNewActivityLogger()


        return try OpenID4CI(requestURI: requestURI, didResolver: didResolver, crypto: crypto, activityLogger: activityLogger!, kms: localKMS, correlationID: correlationID)
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

    func createOpenID4VPInteraction(correlationID: String) throws -> OpenID4VP {
        guard let crypto = self.crypto else {
            throw WalletSDKError.runtimeError("crypto is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw WalletSDKError.runtimeError("did resolver is not initialized, call initSDK()")
        }
        guard let activityLogger = self.activityLogger else {
            throw WalletSDKError.runtimeError("activity logger is not initialized, call initSDK()")
        }
        guard let kms = self.kms else {
            throw WalletSDKError.runtimeError("kms is not initialized, call initSDK()")
        }
        return OpenID4VP(didResolver: didResolver, crypto: crypto, activityLogger: activityLogger, kms: kms, correlationID: correlationID)
    }

    func getAttestationVC(
        didVerificationMethod: ApiVerificationMethod,
        attestationURL: String,
        disableTLSVerify: Bool,
        attestationPayload: String,
        attestationToken: String?
   ) throws -> String {
        let opts = AttestationNewCreateClientArgs(attestationURL, crypto)!
        if (disableTLSVerify) {
            opts.disableHTTPClientTLSVerify()
        }
        if (attestationToken != nil) {
            opts.add(ApiHeader("Authorization", value: "Bearer "+attestationToken!))
        }

        // opts.add(ApiHeader("X-Correlation-Id", value: "ramdomstr"))

        let attestClient = AttestationNewClient(opts, nil)!


        let attestationVC = try attestClient.getAttestationVC(didVerificationMethod, attestationPayloadJSON: attestationPayload)

        return try attestationVC.serialize(nil)
    }
}

