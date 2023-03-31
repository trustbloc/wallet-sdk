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
        
        let args = Openid4ciNewArgs(requestURI, "ClientID", self.crypto, self.didResolver)
        
        let opts = Openid4ciNewOpts()
        opts!.setActivityLogger(activityLogger)
        
        self.initiatedInteraction = Openid4ciNewInteraction(args, opts, nil)!
    }
    
    func authorize() throws -> Openid4ciAuthorizeResult {
        return try initiatedInteraction.authorize()
    }

    func issuerURI()-> String {
        return initiatedInteraction.issuerURI()
    }
    
    func requestCredential(didVerificationMethod: ApiVerificationMethod, otp: String) throws -> ApiVerifiableCredential{
        let credentials  = try initiatedInteraction.requestCredential(withPIN: didVerificationMethod, pin:otp)
        return credentials.atIndex(0)!;
    }
    
    public func serializeDisplayData(issuerURI: String, vcCredentials: ApiVerifiableCredentialsArray) -> String{
       let resolvedDisplayData = DisplayResolve(vcCredentials, issuerURI, nil, nil)
        return resolvedDisplayData!.serialize(nil)
    }
}
