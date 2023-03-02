//
//  OpenID4CI.swift
//  Runner
//
//  Created by Volodymyr Kubiv on 28.02.2023.
//

import Foundation
import Walletsdk

public class OpenID4CI {
    private var didResolver: ApiDIDResolverProtocol
    private var crypto: ApiCryptoProtocol
    private var activityLogger: ApiActivityLoggerProtocol
    
    private var initiatedInteraction: Openid4ciInteraction
    
    init (requestURI: String, didResolver: ApiDIDResolverProtocol, crypto: ApiCryptoProtocol, activityLogger: ApiActivityLoggerProtocol) {
        self.didResolver = didResolver
        self.crypto = crypto
        self.activityLogger = activityLogger
        
        let clientConfig =  Openid4ciClientConfig("ClientID", crypto: self.crypto, didRes: self.didResolver, activityLogger: activityLogger)
        self.initiatedInteraction = Openid4ciNewInteraction(requestURI, clientConfig, nil)!
    }
    
    func authorize() throws -> Openid4ciAuthorizeResult {
        return try initiatedInteraction.authorize()
    }

    func issuerURI()-> String {
        return initiatedInteraction.issuerURI()
    }
    
    func requestCredential(otp: String, didVerificationMethod: ApiVerificationMethod) throws -> ApiVerifiableCredential{
        let credentialRequest = Openid4ciNewCredentialRequestOpts( otp )
        let credResp  = try initiatedInteraction.requestCredential(credentialRequest, vm: didVerificationMethod)
        return credResp.atIndex(0)!;
    }
    
    public func resolveCredentialDisplay(issuerURI: String, vcCredentials: ApiVerifiableCredentialsArray) -> String{
        let resolveOpts = DisplayNewResolveOpts(vcCredentials, issuerURI)

        let resolvedDisplayData = DisplayResolve(resolveOpts, nil)
        return resolvedDisplayData!.serialize(nil)
    }
}
