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
    private var vpQueryContent: Data?
    
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
    func startVPInteraction(authorizationRequest: String) throws {
        let clientConfig = Openid4vpClientConfig(keyReader, crypto: crypto, didResolver: didResolver, ldDocumentLoader: documentLoader, activityLogger: activityLogger)

        let interaction = Openid4vpInteraction(authorizationRequest, config: clientConfig)
        
        vpQueryContent = try interaction!.getQuery()
        initiatedInteraction = interaction
    }
    
    func getMatchedSubmissionRequirements(storedCredentials: ApiVerifiableCredentialsArray)
        throws -> CredentialSubmissionRequirementArray {
        guard let vpQueryContent = self.vpQueryContent else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        
        return  try CredentialNewInquirer(documentLoader)!.getSubmissionRequirements(vpQueryContent, contents: CredentialCredentialsOpt(storedCredentials))
    }
    
    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    func presentCredential(didVerificationMethod: ApiVerificationMethod, selectedCredentials: ApiVerifiableCredentialsArray) throws {
        guard let vpQueryContent = self.vpQueryContent else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        guard let initiatedInteraction = self.initiatedInteraction else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        
        let  verifiablePresentation = try CredentialNewInquirer(documentLoader)!.query(vpQueryContent, contents: CredentialCredentialsOpt(selectedCredentials))
                       
        try initiatedInteraction.presentCredential(verifiablePresentation.content(), vm: didVerificationMethod)
    } 
    
}
