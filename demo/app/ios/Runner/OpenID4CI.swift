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

        let trace = OtelNewTrace(nil)

        let args = Openid4ciNewArgs(requestURI, self.crypto, self.didResolver)
        
        let opts = Openid4ciNewOpts()
        opts!.setActivityLogger(activityLogger)
        opts!.add(trace!.traceHeader())
        
        self.initiatedInteraction = Openid4ciNewInteraction(args, opts, nil)!
    }
    
    func checkFlow() throws -> String {
        let issuerCapabilities = initiatedInteraction.issuerCapabilities()
        if ((issuerCapabilities!.authorizationCodeGrantTypeSupported())){
            return "auth-code-flow"
        }
        if ((issuerCapabilities!.preAuthorizedCodeGrantTypeSupported())){
           return "preauth-code-flow"
        }
        return ""
    }
    
    func getAuthorizationLink(scope1: String, scope2: String, clientID: String, redirectURI: String) throws -> String {
        let issuerCapabilities = initiatedInteraction.issuerCapabilities()
        if !(issuerCapabilities?.authorizationCodeGrantTypeSupported())! {
            return "Not implemented"
        }
        
        let scopes = ApiStringArray()
        scopes!.append(scope1)!.append(scope2)
        // TODO #423 Read withScopes and redirect uri from flutter enviornment. Replace these with approriate values as of now.
        // TODO #426 error handling
        let authorizationLink = initiatedInteraction.createAuthorizationURL(withScopes: clientID, redirectURI: redirectURI, scopes: scopes, error: nil)
        
        return authorizationLink
    }

    
    
    func pinRequired() throws -> Bool {
        let issuerCapabilities = initiatedInteraction.issuerCapabilities()
        if  !issuerCapabilities!.preAuthorizedCodeGrantTypeSupported() {
            return false
        }
        return try initiatedInteraction.issuerCapabilities()!.preAuthorizedCodeGrantParams().pinRequired()
    }

    func issuerURI()-> String {
        return initiatedInteraction.issuerURI()
    }
    
    func requestCredentialWithAuth(didVerificationMethod: ApiVerificationMethod, redirectURIWithParams: String) throws -> VerifiableCredential {
        let credentials = try initiatedInteraction.requestCredential(withAuth: didVerificationMethod, redirectURIWithAuthCode: redirectURIWithParams)
        return credentials.atIndex(0)!;
    }
    
    func requestCredential(didVerificationMethod: ApiVerificationMethod, otp: String) throws -> VerifiableCredential{
        let credentials  = try initiatedInteraction.requestCredential(withPIN: didVerificationMethod, pin:otp)
        return credentials.atIndex(0)!;
    }
    
    public func serializeDisplayData(issuerURI: String, vcCredentials: VerifiableCredentialsArray) -> String{
       let resolvedDisplayData = DisplayResolve(vcCredentials, issuerURI, nil, nil)
        return resolvedDisplayData!.serialize(nil)
    }
}
