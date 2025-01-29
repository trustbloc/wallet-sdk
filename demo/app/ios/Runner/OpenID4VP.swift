/*
 Copyright Gen Digital Inc. All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
 */

import Foundation
import Walletsdk

enum OpenID4VPError: Error {
    case runtimeError(String)
}

public class OpenID4VP {
    private var didResolver: ApiDIDResolverProtocol
    private var crypto: ApiCryptoProtocol
    private var activityLogger: ApiActivityLoggerProtocol
    private var kms: LocalkmsKMS
    private var correlationID: String

    private var initiatedInteraction: Openid4vpInteraction?
    private var vpQueryContent: Data?
    private var submissionRequirement: CredentialSubmissionRequirementArray?

    init (didResolver: ApiDIDResolverProtocol, crypto: ApiCryptoProtocol, activityLogger: ApiActivityLoggerProtocol, kms: LocalkmsKMS, correlationID: String) {
        self.didResolver = didResolver
        self.crypto = crypto
        self.activityLogger = activityLogger
        self.kms = kms
        self.correlationID = correlationID
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
        opts!.add(ApiHeader("X-Correlation-Id", value: self.correlationID))
        opts!.enableAddingDIProofs(kms)

        let interaction = Openid4vpNewInteraction(args, opts, nil)

        vpQueryContent = try interaction!.getQuery()
        initiatedInteraction = interaction
    }

    func getMatchedSubmissionRequirements(storedCredentials: VerifiableCredentialsArray)
    throws -> CredentialSubmissionRequirementArray {
        guard let vpQueryContent = self.vpQueryContent else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }

        submissionRequirement = try CredentialNewInquirer(CredentialInquirerOpts()?.setDIDResolver(didResolver), nil)!.getSubmissionRequirements(vpQueryContent, credentials: storedCredentials)
        return submissionRequirement!
    }

    func checkWithTrustRegistry(evaluatePresentationURL: String) throws -> TrustregistryEvaluationResult {

        guard let initiatedInteraction = self.initiatedInteraction else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }

        guard let submissionRequirements = self.submissionRequirement else {
            throw OpenID4VPError.runtimeError("Before you can call checkWithTrustRegistry, you need call getMatchedSubmissionRequirements first")
        }

        let presentationRequest = TrustregistryPresentationRequest()

        for rInd in 0..<submissionRequirements.len() {
            let requirement = submissionRequirements.atIndex(rInd)!
            for dInd in 0..<requirement.descriptorLen() {
                let descriptor = requirement.descriptor(at: dInd)!
                for credInd in 0..<descriptor.matchedVCs!.length() {
                    let cred = descriptor.matchedVCs!.atIndex(credInd)!

                    let claimsToCheck = TrustregistryCredentialClaimsToCheck();
                    let credentialClaims = try presentedClaims(credential: cred);

                    claimsToCheck.credentialID = cred.id_()
                    claimsToCheck.issuerID = cred.issuerID()
                    claimsToCheck.credentialTypes = cred.types()
                    claimsToCheck.expirationDate = cred.expirationDate()
                    claimsToCheck.issuanceDate = cred.issuanceDate()
                    claimsToCheck.credentialClaimKeys = credentialClaims
                    presentationRequest.addCredentialClaims(claimsToCheck)
                }
            }
        }

        let trustInfo = try initiatedInteraction.trustInfo()
        presentationRequest.verifierDID = trustInfo.did
        presentationRequest.verifierDomain = trustInfo.domain

        let config = TrustregistryRegistryConfig()
        config.evaluatePresentationURL = evaluatePresentationURL
        config.add(ApiHeader("X-Correlation-Id", value: self.correlationID))

        return try TrustregistryRegistry(config)!.evaluatePresentation(presentationRequest)
    }

    /**
     * initiatedInteraction has PresentCredential method which presents credentials to redirect uri from request object.
     */
    func presentCredential(selectedCredentials: VerifiableCredentialsArray, customScopes: Dictionary<String, Any>,
        didVerificationMethod: ApiVerificationMethod?, attestationVC: String?, attestationVM: ApiVerificationMethod?) throws {
        //         guard let vpQueryContent = self.vpQueryContent else {
        //             throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        //         }
        guard let initiatedInteraction = self.initiatedInteraction else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }

        let opts = Openid4vpNewPresentCredentialOpts()

        if (attestationVC != nil && attestationVM != nil) {
          opts?.setAttestationVC(attestationVM, vc: attestationVC)
        }

        var test = [String : String] ()
        test["user"] = "123456"

        let data = try JSONEncoder().encode(test)
        let serializedInteractionDetails = String(data: data, encoding: .utf8)!

        opts?.setInteractionDetails(serializedInteractionDetails)

        for scope in customScopes {
            opts?.addScopeClaim(scope.key, claimJSON: scope.value as? String)
        }

        try initiatedInteraction.presentCredentialOpts(selectedCredentials, opts: opts)

    }


    func presentedClaims(credential: VerifiableCredential) throws -> Openid4vpCredentialClaimKeys {
       return try initiatedInteraction?.presentedClaims(credential) ?? Openid4vpCredentialClaimKeys()
    }

    // returns the presentation acknowledgment
    func noConsentAcknowledgement() throws -> Openid4vpAcknowledgment {
        guard self.initiatedInteraction != nil else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }
        return try initiatedInteraction?.acknowledgment() ?? Openid4vpAcknowledgment()
    }

    func getCustomScope() throws -> [String] {
        guard let initiatedInteraction = self.initiatedInteraction else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }

        let customScopes = initiatedInteraction.customScope()
        var customScopesList = [String]()

        if (customScopes?.length() != 0){
            for i in 0...((customScopes?.length() ?? 0)-1) {
                if (customScopes?.atIndex(i) != "openid"){
                    customScopesList.append(customScopes?.atIndex(i) ?? "")
                }
            }
        }

        // Otherwise return the default scope
        return customScopesList
    }

    func getVerifierDisplayData() throws -> Openid4vpVerifierDisplayData {
        guard self.initiatedInteraction != nil else {
            throw OpenID4VPError.runtimeError("OpenID4VP interaction not properly initialized, call processAuthorizationRequest first")
        }

        return initiatedInteraction?.verifierDisplayData() ?? Openid4vpVerifierDisplayData()
    }
}
