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
    
    private var initiatedInteraction: Openid4vpInteraction?
    private var verifiablePresentation: CredentialVerifiablePresentation?
    
    init (keyReader:LocalkmsKMS, didResolver: ApiDIDResolverProtocol, documentLoader: ApiLDDocumentLoaderProtocol, crypto: ApiCryptoProtocol) {
        self.keyReader = keyReader
        self.didResolver = didResolver
        self.documentLoader = documentLoader
        self.crypto = crypto
    }
    
    func processAuthorizationRequest(authorizationRequest: String, storedCredentials: Array<String>) throws -> Array<String> {
        let interaction = Openid4vpInteraction(authorizationRequest, keyHandle: keyReader, crypto: crypto, didResolver: didResolver, ldDocumentLoader: documentLoader)
        
        let query = try? interaction?.getQuery()

        let creds = ApiVerifiableCredentialsArray()
        for cred in storedCredentials {
            creds?.add(ApiVerifiableCredential(cred))
        }
 
        verifiablePresentation = try CredentialNewInquirer(documentLoader)?.query(query, contents: CredentialCredentialsOpt(creds))
        
        let matchedCreds = try verifiablePresentation?.credentials()

        initiatedInteraction = interaction
        //Todo loop through matched credentials
        var credlist :[String] = [(matchedCreds?.atIndex(0)?.content)!]
        
        return credlist
    }
    
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
