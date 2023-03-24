//
//  IntegrationTest.swift
//  WalletTests
//
//  Created by Volodymyr Kubiv on 06.03.2023.
//

import XCTest
import Walletsdk

@testable import Runner

class IntegrationTest: XCTestCase {

    override func setUpWithError() throws {
        // Put setup code here. This method is called before the invocation of each test method in the class.
    }

    override func tearDownWithError() throws {
        // Put teardown code here. This method is called after the invocation of each test method in the class.
    }

    func testFullFlow() throws {
        XCTAssertEqual(VersionGetVersion(), "testVer")
        XCTAssertEqual(VersionGetGitRevision(), "testRev")
        XCTAssertEqual(VersionGetBuildTime(), "testTime")

        let kms = LocalkmsNewKMS(kmsStore(), nil)!
        
        let resolverOpts = DidNewResolverOpts()
        resolverOpts!.setResolverServerURI("http://localhost:8072/1.0/identifiers")
        let didResolver = DidNewResolver(resolverOpts, nil)!
        
        let crypto = kms.getCrypto()

        let didCreator = DidNewCreator(kms, nil)!
        let userDID = try didCreator.create("ion", opts: nil)

        // Issue VCs
        let requestURI = ProcessInfo.processInfo.environment["INITIATE_ISSUANCE_URL"]
        
        XCTAssertTrue(requestURI != "", "requestURI:" + requestURI!)
        
        let openID4CIInteractionArgs = Openid4ciNewArgs(requestURI, "ClientID", crypto, didResolver)

        let ciInteraction = Openid4ciNewInteraction(openID4CIInteractionArgs, nil, nil)
        XCTAssertNotNil(ciInteraction)

        let authorizeResult = try ciInteraction!.authorize()
        XCTAssertFalse(authorizeResult.userPINRequired)
        
        let issuedCreds = try ciInteraction!.requestCredential(userDID.assertionMethod())
        XCTAssertTrue(issuedCreds.length() > 0)

        //Presenting VCs
        let authorizationRequestURI = ProcessInfo.processInfo.environment["INITIATE_VERIFICATION_URL"]
        XCTAssertTrue(authorizationRequestURI != "", "authorizationRequestURI:" + authorizationRequestURI!)
        
        let openID4VPArgs = Openid4vpNewArgs(authorizationRequestURI, kms, crypto, didResolver)
        
        let vpInteraction = Openid4vpInteraction(openID4VPArgs, opts: nil)!

        let credentialsQuery = try vpInteraction.getQuery()
        let inquirer = CredentialNewInquirer(nil)!

        let submissionRequirements = try inquirer.getSubmissionRequirements(
            credentialsQuery, contents: CredentialCredentialsArg(fromVCArray: issuedCreds))

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

        let selectedCreds = ApiVerifiableCredentialsArray()!
        // Pick first credential from matched creds
        selectedCreds.add(requirementDescriptor.matchedVCs!.atIndex(0))

        // Presenting from selected credentials.
        try vpInteraction.presentCredential(selectedCreds)
    }

}
