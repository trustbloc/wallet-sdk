/*
 Copyright Gen Digital Inc. All Rights Reserved.
 
 SPDX-License-Identifier: Apache-2.0
 */

import XCTest
import Walletsdk
import UIKit

@testable import Runner

class IntegrationTest: XCTestCase {
    
    override func setUpWithError() throws {
        // Put setup code here. This method is called before the invocation of each test method in the class.
    }
    
    override func tearDownWithError() throws {
        // Put teardown code here. This method is called after the invocation of each test method in the class.
    }
    
    /// <#Description#>
    func testFullFlow() throws {
//        XCTAssertEqual(VersionGetVersion(), "testVer")
//        XCTAssertEqual(VersionGetGitRevision(), "testRev")
//        XCTAssertEqual(VersionGetBuildTime(), "testTime")
        
        let trace = OtelNewTrace(nil)
        
        let kms = LocalkmsNewKMS(kmsStore(), nil)!
        
        let resolverOpts = DidNewResolverOpts()
        resolverOpts!.setResolverServerURI("http://localhost:8072/1.0/identifiers")
        let didResolver = DidNewResolver(resolverOpts, nil)!
        
        let crypto = kms.getCrypto()
        
        let jwk = try kms.create(LocalkmsKeyTypeED25519)
        
        var error: NSError?
        
        let userDID = DidionCreateLongForm(jwk, &error)
        XCTAssertNil(error)
        
        // Issue VCs
        let requestURI = ProcessInfo.processInfo.environment["INITIATE_ISSUANCE_URL"]
        
        XCTAssertTrue(requestURI != "", "requestURI:" + requestURI!)
        
        let openID4CIInteractionArgs = Openid4ciNewIssuerInitiatedInteractionArgs(requestURI, crypto, didResolver)
        
        let ciOpts = Openid4ciNewInteractionOpts()
        ciOpts!.add(trace!.traceHeader())
        
        let ciInteraction = Openid4ciNewIssuerInitiatedInteraction(openID4CIInteractionArgs, ciOpts, nil)
        XCTAssertNotNil(ciInteraction)
        
        let pinRequired = try ciInteraction!.preAuthorizedCodeGrantParams().pinRequired()
        XCTAssertFalse(pinRequired)

        let attClient = AttestationNewClient(
          AttestationNewCreateClientArgs("https://localhost:8097/profiles/profileID/profileVersion/wallet/attestation/", crypto)?.disableHTTPClientTLSVerify()?.add(ApiHeader("Authorization", value: "Bearer token")),
          nil
        )!

        let attestationVC = try attClient.getAttestationVC(userDID!.assertionMethod(),
            attestationPayloadJSON: """
                                 {
                                    "type": "urn:attestation:application:trustbloc",
                                    "application": {
                                        "type":    "wallet-cli",
                                        "name":    "wallet-cli",
                                        "version": "1.0"
                                    },
                                    "compliance": {
                                        "type": "fcra"
                                    }
                                }
                                """)

        let preAuthOpts = Openid4ciRequestCredentialWithPreAuthOpts()!.setPIN("")!
        preAuthOpts.setAttestationVC(try userDID!.assertionMethod(), vc: try attestationVC.serialize(nil))

        let issuedCreds  = try ciInteraction!.requestCredential(withPreAuth: userDID!.assertionMethod(), opts: preAuthOpts)
        XCTAssertTrue(issuedCreds.length() > 0)
        
        let acknowledgmentData = try ciInteraction!.acknowledgment().serialize(nil)
        
        try Openid4ciAcknowledgment(acknowledgmentData)!.success();
        
        //Presenting VCs
        let authorizationRequestURI = ProcessInfo.processInfo.environment["INITIATE_VERIFICATION_URL"]
        XCTAssertTrue(authorizationRequestURI != "", "authorizationRequestURI:" + authorizationRequestURI!)
        
        let openID4VPArgs = Openid4vpNewArgs(authorizationRequestURI, crypto, didResolver)
        
        let opts = Openid4vpNewOpts()
        opts!.add(trace!.traceHeader())
        
        let vpInteraction = Openid4vpNewInteraction(openID4VPArgs, opts, nil)!
        
        let credentialsQuery = try vpInteraction.getQuery()
        
        var newInquirerError: NSError?
        let inquirer = CredentialNewInquirer(nil, &newInquirerError)!
        XCTAssertNil(newInquirerError)
        
        let submissionRequirements = try inquirer.getSubmissionRequirements(
            credentialsQuery, credentials: issuedCreds)
        
        XCTAssertTrue(submissionRequirements.len() > 0)
        let requirement = submissionRequirements.atIndex(0)!
        // rule "all" means that we need to satisfy all input descriptors.
        // In case of multiple input descriptors we need to send one credential per descriptor
        // that satisfy it.
        XCTAssertEqual(requirement.rule(), "all")
        
        // In current test case we have only one input descriptor. so we need send only one credential.
        XCTAssertEqual(submissionRequirements.atIndex(0)!.descriptorLen(), 1)
        
        let requirementDescriptor = requirement.descriptor(at:0)!
        // matchedVCs contains list of credentials that match given input descriptor.
        XCTAssertTrue(requirementDescriptor.matchedVCs!.length() > 0)
        
        let selectedCreds = VerifiableCredentialsArray()!
        // Pick first credential from matched creds
        selectedCreds.add(requirementDescriptor.matchedVCs!.atIndex(0))
        
        // Check trust info
        let trustInfo = try vpInteraction.trustInfo()
        XCTAssertTrue(trustInfo.did != "")
        XCTAssertTrue(trustInfo.domain != "")
        
        // Credential trust info
        let oneOfTheCreds = requirementDescriptor.matchedVCs!.atIndex(0)!
        XCTAssertTrue(oneOfTheCreds.id_() != "")
        XCTAssertTrue(oneOfTheCreds.types()!.length() > 0)
        XCTAssertTrue(oneOfTheCreds.issuerID() != "")
        XCTAssertTrue(oneOfTheCreds.issuanceDate() > 0)
        XCTAssertTrue(oneOfTheCreds.expirationDate() > 0)
        
        // Presenting from selected credentials.
        try vpInteraction.presentCredentialOpts(
            selectedCreds,
            opts: Openid4vpNewPresentCredentialOpts()!.addScopeClaim("registration", claimJSON:#"{"email":"test@example.com"}"#)!.addScopeClaim("testscope", claimJSON: #"{"data": "testdata"}"#)!.setAttestationVC(userDID!.assertionMethod(), vc: try attestationVC.serialize(nil))
        )
    }
    
    func testAuthFlow() throws {
        let trace = OtelNewTrace(nil)
        let kms = LocalkmsNewKMS(kmsStore(), nil)!
        
        let resolverOpts = DidNewResolverOpts()
        resolverOpts!.setResolverServerURI("http://localhost:8072/1.0/identifiers")
        let didResolver = DidNewResolver(resolverOpts, nil)!
        
        let crypto = kms.getCrypto()
        
        let jwk = try kms.create(LocalkmsKeyTypeED25519)
        
        var error: NSError?
        
        let userDID = DidionCreateLongForm(jwk, &error)
        XCTAssertNil(error)
        
        // Issue VCs in auth flow
        let requestAuthURI = ProcessInfo.processInfo.environment["INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW"]
        print(requestAuthURI)
        XCTAssertTrue(requestAuthURI != "", "requestAuthURI:" + requestAuthURI!)
        
        let openID4CIInteractionArgs = Openid4ciNewIssuerInitiatedInteractionArgs(requestAuthURI, crypto, didResolver)
        
        let ciOpts = Openid4ciNewInteractionOpts()
        ciOpts!.add(trace!.traceHeader())
        
        let ciInteraction = Openid4ciNewIssuerInitiatedInteraction(openID4CIInteractionArgs, ciOpts, nil)
        XCTAssertNotNil(ciInteraction)
        let redirectURI = "http://127.0.0.1/callback"
        var clientID = "oidc4vc_client"
        var scopes = ApiNewStringArray()!
        var scopesFromArgs = ["openid", "profile"] as! [String]
        
        for scope in scopesFromArgs {
            scopes.append(scope)
        }
        
        if (ciInteraction?.dynamicClientRegistrationSupported != nil) {
            let dynamicRegistrationEndpoint =  ciInteraction?.dynamicClientRegistrationEndpoint(nil)
            XCTAssertNotNil(dynamicRegistrationEndpoint)
            
            let clientMetadata = Oauth2ClientMetadata()
            let grantTypesArr = ApiStringArray()
            grantTypesArr?.append("authorization_code")
            clientMetadata?.setGrantTypes(grantTypesArr)
            XCTAssertNotNil(grantTypesArr)
            
            let redirectURIArr = ApiStringArray()
            redirectURIArr?.append(redirectURI)
            clientMetadata?.setRedirectURIs(redirectURIArr)
            XCTAssertNotNil(redirectURIArr)
            
            clientMetadata?.setScopes(scopes)
            clientMetadata?.setTokenEndpointAuthMethod("none")
            
            let authorizationCodeGrantParams = try ciInteraction?.authorizationCodeGrantParams()
            
            if ((authorizationCodeGrantParams?.hasIssuerState()) != nil) {
                let issuerState = authorizationCodeGrantParams!.issuerState(nil)
                clientMetadata?.setIssuerState(issuerState)
                XCTAssertNotNil(clientMetadata?.issuerState())
            }
            
            let registrationResp = Oauth2RegisterClient(dynamicRegistrationEndpoint, clientMetadata, nil, nil)
            clientID = (registrationResp?.clientID())!
            scopes = (registrationResp!.registeredMetadata())!.scopes()!
        }
        
        let authCodeGrant = ciInteraction!.authorizationCodeGrantTypeSupported()
        XCTAssertTrue(authCodeGrant)
        
        let opts = Openid4ciNewCreateAuthorizationURLOpts()!.setScopes(scopes)
        let authorizationLink = ciInteraction!.createAuthorizationURL(clientID, redirectURI: redirectURI, opts: opts!, error: nil)
        XCTAssertTrue(authorizationLink  != "", "authorizationLink:" + authorizationLink)
        
        var redirectURL = ""
        let r = Redirect()
        let url = URL(string: authorizationLink)!
        r.makeRequest(url: url, callback: { (location) in
            guard let locationURL = location else {return}
            var updatedLoc = locationURL.absoluteString
            redirectURL = updatedLoc.replacingOccurrences(of: "cognito-mock.trustbloc.local",with: "localhost")
            r.makeRequest(url: URL(string: redirectURL)!, callback: { (location) in
                guard let locationURL = location else {return}
                r.makeRequest(url: locationURL, callback: { (location) in
                    guard let locationURL = location else {return}
                    redirectURL = locationURL.absoluteString
                    do {
                        let issuedCreds = try ciInteraction!.requestCredential(withAuth:userDID!.assertionMethod(), redirectURIWithAuthCode: redirectURL, opts: nil)
                        XCTAssertTrue(issuedCreds.length() > 0)
                    } catch {
                        print("Error: \(error)")
                    }
                    return
                })
            })
        })
    }
}


// More efficient click-tracking with HTTP GET to obtain the "302" response, but not follow the redirect through to the Location.
// The callback is used to return the Location header back from the async task
class Redirect : NSObject {
    var session: URLSession?
    
    override init() {
        super.init()
        session = URLSession(configuration: .default, delegate: self, delegateQueue: nil)
    }
    
    func makeRequest(url: URL, callback: @escaping (URL?) -> ()) {
        let task = self.session?.dataTask(with: url) {(data, response, error) in
            guard response != nil else {
                return
            }
            if let response = response as? HTTPURLResponse {
                if let l = response.value(forHTTPHeaderField: "Location") {
                    callback(URL(string: l))
                }
            }
        }
        task?.resume()
    }
}

extension Redirect: URLSessionDelegate, URLSessionTaskDelegate {
    func urlSession(_ session: URLSession, task: URLSessionTask, willPerformHTTPRedirection response: HTTPURLResponse, newRequest request: URLRequest, completionHandler: @escaping (URLRequest?) -> Void) {
        completionHandler(nil)
    }
}
