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
    
    func testFullFlow() throws {
        // 1. Initialize with proper error handling
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
          AttestationNewCreateClientArgs("https://localhost:8097/profiles/profileID/profileVersion/wallet/attestation/", crypto)?
          .disableHTTPClientTLSVerify()?
          .add(ApiHeader("Authorization", value: "Bearer token")),
          nil
        )!

      guard let attestationVC = try? attClient.getAttestationVC(
               userDID!.assertionMethod(),
               attestationPayloadJSON: """
               {
                   "type": "urn:attestation:application:trustbloc",
                   "application": {
                       "type": "wallet-cli",
                       "name": "wallet-cli",
                       "version": "1.0"
                   },
                   "compliance": {
                       "type": "fcra"
                   }
               }
               """) else {
               XCTFail("Attestation VC creation failed")
               return
           }

    // 7. Serialize VC
   // Fixed serialization with proper error handling
    var serializationError: NSError?
    let serializedVC = attestationVC.serialize(&serializationError)
    if let error = serializationError {
        XCTFail("VC serialization failed: \(error.localizedDescription)")
        return
    }

     // 8. Request credentials
          guard let preAuthOpts = Openid4ciRequestCredentialWithPreAuthOpts()!
              .setPIN("")!
              .setAttestationVC(try userDID!.assertionMethod(), vc: serializedVC) else {
              XCTFail("Pre-auth options setup failed")
              return
          }

        let issuedCreds  = try ciInteraction!.requestCredential(withPreAuth: userDID!.assertionMethod(), opts: preAuthOpts)
        XCTAssertTrue(issuedCreds.length() > 0)
        
        let acknowledgmentData = try ciInteraction!.acknowledgment().serialize(nil)
        try Openid4ciAcknowledgment(acknowledgmentData)!.success();
        
        // Presenting VCs
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
        XCTAssertEqual(requirement.rule(), "all")
        XCTAssertEqual(submissionRequirements.atIndex(0)!.descriptorLen(), 1)
        
        let requirementDescriptor = requirement.descriptor(at:0)!
        XCTAssertTrue(requirementDescriptor.matchedVCs!.length() > 0)
        
        let selectedCreds = VerifiableCredentialsArray()!
        selectedCreds.add(requirementDescriptor.matchedVCs!.atIndex(0))
        
        // Check trust info
        let trustInfo = try vpInteraction.trustInfo()
        XCTAssertTrue(trustInfo.did != "")
        XCTAssertTrue(trustInfo.domain != "")
        
        // Present credentials
        try vpInteraction.presentCredentialOpts(
            selectedCreds,
            opts: Openid4vpNewPresentCredentialOpts()!
                .addScopeClaim("registration", claimJSON:#"{"email":"test@example.com"}"#)!
                .addScopeClaim("testscope", claimJSON: #"{"data": "testdata"}"#)!
                .setAttestationVC(userDID!.assertionMethod(), vc: serializedVC)
        )
    }
}
