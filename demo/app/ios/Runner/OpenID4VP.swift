//
//  OpenID4VP.swift
//  Runner
//
//  Created by Volodymyr Kubiv on 29.11.2022.
//

import Foundation
import Walletsdk

enum OpenID4VPError: Error {
    case runtimeError(String)
}

public class OpenID4VP {
    private var keyReader:LocalkmsKMS
    private var didResolver: ApiDIDResolverProtocol
    private var documentLoader: ApiLDDocumentLoaderProtocol
    private var crypto: ApiCryptoProtocol
    private var activityLogger: ApiActivityLoggerProtocol
    
    private var initiatedInteraction: Openid4vpInteraction?
    private var verifiablePresentation: CredentialVerifiablePresentation?
    
    init (keyReader:LocalkmsKMS, didResolver: ApiDIDResolverProtocol, documentLoader: ApiLDDocumentLoaderProtocol, crypto: ApiCryptoProtocol, activityLogger: ApiActivityLoggerProtocol) {
        self.keyReader = keyReader
        self.didResolver = didResolver
        self.documentLoader = documentLoader
        self.crypto = crypto
        self.activityLogger = activityLogger
    }
    
    /**
     * Openid4vpClientConfig contains various parameters for an OpenID4VP Interaction. ActivityLogger is optional, but if provided then activities will be logged there.
       If not provided, then no activities will be logged.
     * InitiatedInteraction is local variable to intiate  Openid4vpInteraction representing a single OpenID4VP interaction between a wallet and a verifier.
     * The methods defined on this object are used to help guide the calling code through the OpenID4VP flow.
     */
    func processAuthorizationRequest(authorizationRequest: String, credentials: Array<ApiVerifiableCredential>) throws -> Array<String> {
        let clientConfig = Openid4vpClientConfig(keyReader, crypto: crypto, didResolver: didResolver, ldDocumentLoader: documentLoader, activityLogger: activityLogger)

        let interaction = Openid4vpInteraction(authorizationRequest, config: clientConfig)
        
        let query = try? interaction?.getQuery()

        let creds = ApiVerifiableCredentialsArray()
        for cred in credentials {
            creds?.add(cred)
        }
 
        let  verifiablePresentation = try CredentialNewInquirer(documentLoader)?.query(query, contents: CredentialCredentialsOpt(creds))
       
        self.verifiablePresentation = verifiablePresentation
        
        let matchedCreds = try verifiablePresentation!.credentials()

        initiatedInteraction = interaction
        
        var credList: [String] = []
        
        for i in 0...(matchedCreds.length()-1) {
            credList.append((matchedCreds.atIndex(i)?.serialize(nil))!)
        }
        
        
        return credList
    }
    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    func presentCredential(didVerificationMethod: ApiVerificationMethod) throws {
        guard let verifiablePresentation = self.verifiablePresentation else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        
        
        guard let initiatedInteraction = self.initiatedInteraction else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
                
        try initiatedInteraction.presentCredential(verifiablePresentation.content(), vm: didVerificationMethod)
    }
}
