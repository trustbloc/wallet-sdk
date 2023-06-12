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
    private var didResolver: ApiDIDResolverProtocol
    private var crypto: ApiCryptoProtocol
    private var activityLogger: ApiActivityLoggerProtocol
    
    private var initiatedInteraction: Openid4vpInteraction?
    private var vpQueryContent: Data?
    
    init (didResolver: ApiDIDResolverProtocol, crypto: ApiCryptoProtocol, activityLogger: ApiActivityLoggerProtocol) {
        self.didResolver = didResolver
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
        let trace = OtelNewTrace(nil)

        let args = Openid4vpNewArgs(authorizationRequest, crypto, didResolver)

        let opts = Openid4vpNewOpts()
        opts!.setActivityLogger(activityLogger)
        opts!.add(trace!.traceHeader())
        
        let interaction = Openid4vpNewInteraction(args, opts, nil)
        
        vpQueryContent = try interaction!.getQuery()
        initiatedInteraction = interaction
    }
    
    func getMatchedSubmissionRequirements(storedCredentials: VerifiableCredentialsArray)
        throws -> CredentialSubmissionRequirementArray {
        guard let vpQueryContent = self.vpQueryContent else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        
            return try CredentialNewInquirer(CredentialInquirerOpts()?.setDIDResolver(didResolver))!.getSubmissionRequirements(vpQueryContent, credentials: storedCredentials)
    }
    
    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    func presentCredential(selectedCredentials: VerifiableCredentialsArray) throws {
//         guard let vpQueryContent = self.vpQueryContent else {
//             throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
//         }
        guard let initiatedInteraction = self.initiatedInteraction else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        
//         let  verifiablePresentation = try CredentialNewInquirer(documentLoader)!.query(vpQueryContent, credentials: selectedCredentials)
                       
        try initiatedInteraction.presentCredential(selectedCredentials)
    }
    
    func getVerifierDisplayData() throws -> Openid4vpVerifierDisplayData {
        guard self.initiatedInteraction != nil else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        
      let verifierDisplayData =  try initiatedInteraction?.verifierDisplayData()
      return verifierDisplayData!
         
    }
}
