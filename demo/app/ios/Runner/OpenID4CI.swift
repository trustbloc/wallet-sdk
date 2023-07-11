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

        let args = Openid4ciNewInteractionArgs(requestURI, self.crypto, self.didResolver)
        
        let opts = Openid4ciNewInteractionOpts()
        opts!.setActivityLogger(activityLogger)
        opts!.add(trace!.traceHeader())
        
        self.initiatedInteraction = Openid4ciNewInteraction(args, opts, nil)!
    }
    
    func checkFlow() throws -> String {
        if ((initiatedInteraction.authorizationCodeGrantTypeSupported())){
            return "auth-code-flow"
        }
        if ((initiatedInteraction.preAuthorizedCodeGrantTypeSupported())){
           return "preauth-code-flow"
        }
        return ""
    }
    
    func createAuthorizationURLWithScopes(scopes: [String], clientID: String, redirectURI: String) throws  -> String {
     let scopesArr = ApiStringArray()
        for scope in scopes {
            scopesArr!.append(scope)!
        }

      var error: NSError?
        
      let opts = Openid4ciNewCreateAuthorizationURLOpts()!.setScopes(scopesArr)
    
       let authorizationLink =  initiatedInteraction.createAuthorizationURL(clientID, redirectURI: redirectURI, opts: opts, error: &error)
        if let actualError = error {
            print("error in authorizations", error!.localizedDescription)
            throw actualError
       }
        
      return authorizationLink
    }
    
    func createAuthorizationURL(clientID: String, redirectURI: String) throws  -> String {
      var error: NSError?
    
        let authorizationLink =  initiatedInteraction.createAuthorizationURL(clientID, redirectURI: redirectURI, opts: nil, error: &error)
        if let actualError = error {
            print("error in authorizations", error!.localizedDescription)
            throw actualError
       }
    
      return authorizationLink
    }
    
    func pinRequired() throws -> Bool {
       return try initiatedInteraction.preAuthorizedCodeGrantParams().pinRequired()
    }

    func issuerURI()-> String {
        return initiatedInteraction.issuerURI()
    }
    
    func requestCredentialWithAuth(didVerificationMethod: ApiVerificationMethod, redirectURIWithParams: String) throws -> VerifiableCredential {
        let credentials = try initiatedInteraction.requestCredential(withAuth: didVerificationMethod, redirectURIWithAuthCode: redirectURIWithParams, opts: nil)
        return credentials.atIndex(0)!;
    }
    
    func requestCredential(didVerificationMethod: ApiVerificationMethod, otp: String) throws -> VerifiableCredential{
        let opts = Openid4ciRequestCredentialWithPreAuthOpts()!.setPIN(otp)
        let credentials  = try initiatedInteraction.requestCredential(withPreAuth: didVerificationMethod, opts: opts)
        return credentials.atIndex(0)!;
    }
    
    public func serializeDisplayData(issuerURI: String, vcCredentials: VerifiableCredentialsArray) -> String{
       let resolvedDisplayData = DisplayResolve(vcCredentials, issuerURI, nil, nil)
        return resolvedDisplayData!.serialize(nil)
    }
    
    func dynamicRegistrationSupported() throws -> ObjCBool {
        var dynamicRegistrationSupported: ObjCBool = false
        try initiatedInteraction.dynamicClientRegistrationSupported(&dynamicRegistrationSupported)
        return dynamicRegistrationSupported
    }
    
    func dynamicRegistrationEndpoint() throws -> String {
        var error: NSError?
        let endpoint = initiatedInteraction.dynamicClientRegistrationEndpoint(&error)
        if let actualError = error {
            print("error from dynamic registration endpoint",  actualError.localizedDescription)
            throw actualError
       }
        return endpoint
    }
    
    func getAuthorizationCodeGrantParams() throws -> Openid4ciAuthorizationCodeGrantParams {
        return  try initiatedInteraction.authorizationCodeGrantParams()
    }
}
