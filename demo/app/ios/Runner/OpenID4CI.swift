/*
 Copyright Gen Digital Inc. All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
 */

import Foundation
import Walletsdk

public class OpenID4CI {
    private var didResolver: ApiDIDResolverProtocol
    private var crypto: ApiCryptoProtocol
    private var activityLogger: ApiActivityLoggerProtocol
    private var kms: LocalkmsKMS
    private var correlationID: String

    private var initiatedInteraction: Openid4ciIssuerInitiatedInteraction

    init (requestURI: String, didResolver: ApiDIDResolverProtocol, crypto: ApiCryptoProtocol, activityLogger: ApiActivityLoggerProtocol, kms: LocalkmsKMS, correlationID: String) throws {
        self.didResolver = didResolver
        self.crypto = crypto
        self.activityLogger = activityLogger
        self.kms = kms
        self.correlationID = correlationID

        let trace = OtelNewTrace(nil)

        let args = Openid4ciNewIssuerInitiatedInteractionArgs(requestURI, crypto, didResolver)

        let opts = Openid4ciNewInteractionOpts()
        opts!.setActivityLogger(activityLogger)
        opts!.add(trace!.traceHeader())
        opts!.add(ApiHeader("X-Correlation-Id", value: self.correlationID))
        opts!.enableDIProofChecks(kms)


        var error: NSError?
        let interaction = Openid4ciNewIssuerInitiatedInteraction(args, opts, &error)
        if let actualError = error {
            throw actualError
        }

        self.initiatedInteraction = interaction!
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

    func createAuthorizationURL(clientID: String, redirectURI: String, oauthDiscoverableClientURI: String,  scopes:ApiStringArray) throws  -> String {
        var error: NSError?
        let opts = Openid4ciNewCreateAuthorizationURLOpts()
        if (scopes.length() != 0) {
            opts!.setScopes(scopes)
        }

        if (oauthDiscoverableClientURI != "") {
            opts!.useOAuthDiscoverableClientIDScheme()
        }


        let authorizationLink =  initiatedInteraction.createAuthorizationURL(clientID, redirectURI: redirectURI, opts: opts, error: &error)
        if let actualError = error {
            print("error while creating authorization link", error!.localizedDescription)
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

    func getCredentialOfferDisplayData() throws -> DisplayData {
        let issuerMetadata = try initiatedInteraction.issuerMetadata()

        return DisplayResolveCredentialOffer(
            issuerMetadata,
            initiatedInteraction.offeredCredentialsTypes(), ""
        )!
    }

    func requestCredentialWithAuth(didVerificationMethod: ApiVerificationMethod, redirectURIWithParams: String) throws -> VerifiableCredential {
        let credentials = try initiatedInteraction.requestCredential(withAuth: didVerificationMethod, redirectURIWithAuthCode: redirectURIWithParams, opts: nil)
        return credentials.atIndex(0)!;
    }

    func requestCredentials(didVerificationMethod: ApiVerificationMethod, otp: String,
        attestationVC: String?, attestationVM: ApiVerificationMethod?) throws -> Array<Dictionary<String, Any>> {
        let opts = Openid4ciRequestCredentialWithPreAuthOpts()!.setPIN(otp)!

        if(attestationVC != nil) {
            opts.setAttestationVC(attestationVM!, vc: attestationVC)
        }

        let credentials  = try initiatedInteraction.requestCredential(withPreAuth: didVerificationMethod, opts: opts)
        return convertVerifiableCredentialsWithIdArray(arr:credentials);
    }

    func requireAcknowledgment() throws -> ObjCBool{
        var ackResp: ObjCBool = false
        try initiatedInteraction.requireAcknowledgment(&ackResp)
        return ackResp
    }

    func acknowledgeSuccess() throws {
        var error: NSError?
        let serializedStateResp = try initiatedInteraction.acknowledgment().serialize(&error)
        if let actualError = error {
            print("error from acknowledge success",  actualError.localizedDescription)
            throw actualError
        }

        let acknowledgement = try Openid4ciNewAcknowledgment(serializedStateResp, &error)
        if let actualError = error {
            print("error from new acknowledgement",  actualError.localizedDescription)
            throw actualError
        }


        var test = [String : String] ()
        test["user"] = "123456"

        let data = try JSONEncoder().encode(test)
        let serializedInteractionDetails = String(data: data, encoding: .utf8)!

        try acknowledgement?.setInteractionDetails(serializedInteractionDetails)

        try acknowledgement?.success()
    }


    func acknowledgeReject() throws {
        return try initiatedInteraction.acknowledgment().reject()
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

    public func checkWithTrustRegistry(evaluateIssuanceURL: String) throws -> TrustregistryEvaluationResult {
        let issuanceRequest = TrustregistryIssuanceRequest()

        let trustInfo = try initiatedInteraction.issuerTrustInfo()
        issuanceRequest.issuerDID = trustInfo.did
        issuanceRequest.issuerDomain = trustInfo.domain

        for rInd in 0..<trustInfo.offerLength() {
            let offer = trustInfo.offer(at:rInd)!

            let credentialOffer = TrustregistryCredentialOffer();
            credentialOffer.credentialFormat = offer.credentialFormat
            credentialOffer.credentialType = offer.credentialType

            issuanceRequest.addCredentialOffers(credentialOffer)
        }

        let config = TrustregistryRegistryConfig()
        config.evaluateIssuanceURL = evaluateIssuanceURL
        config.add(ApiHeader("X-Correlation-Id", value: self.correlationID))

        return try TrustregistryRegistry(config)!.evaluateIssuance(issuanceRequest)
    }

    func getAuthorizationCodeGrantParams() throws -> Openid4ciAuthorizationCodeGrantParams {
        return  try initiatedInteraction.authorizationCodeGrantParams()
    }

    func getIssuerMetadata() throws -> Openid4ciIssuerMetadata {
        return try initiatedInteraction.issuerMetadata()
    }

    func verifyIssuer() throws -> String {
        var error: NSError?
        let issuerServiceURL = initiatedInteraction.verifyIssuer(&error)
        if let actualError = error {
            print("error from verify issuer",  actualError.localizedDescription)
            throw actualError
        }
        return issuerServiceURL
    }


}
