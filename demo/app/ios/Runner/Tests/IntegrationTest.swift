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
        let kms = LocalkmsNewKMS(kmsStore(), nil)!
        let didResolver = DidNewResolver("http://localhost:8072/1.0/identifiers", nil)!
        let crypto = kms.getCrypto()
        let documentLoader = LdNewDocLoader()!
        let activityLogger = MemNewActivityLogger()!

        let didCreator = DidNewCreatorWithKeyWriter(kms, nil)!
        let userDID = try didCreator.create("ion", createDIDOpts: ApiCreateDIDOpts())

        // Issue VCs
        let cfg = Openid4ciClientConfig("ClientID", crypto: crypto, didRes:didResolver, activityLogger:	activityLogger)
        let requestURI = ProcessInfo.processInfo.environment["INITIATE_ISSUANCE_URL"]
        
        XCTAssertTrue(requestURI != "", "requestURI:" + requestURI!)

        let ciInteraction = Openid4ciNewInteraction(requestURI, cfg, nil)
        XCTAssertNotNil(ciInteraction)

        let authorizeResult = try ciInteraction!.authorize()
        XCTAssertTrue(!authorizeResult.userPINRequired)

        let otp = ""
        let issuedCreds = try ciInteraction!.requestCredential(
            Openid4ciNewCredentialRequestOpts(otp), vm: userDID.assertionMethod())
        XCTAssertTrue(issuedCreds.length() > 0)

        //Presenting VCs
        let vpConfig = Openid4vpClientConfig(kms,
                                             crypto: crypto,
                                             didResolver: didResolver,
                                             ldDocumentLoader: documentLoader,
                                             activityLogger: activityLogger)
        let authorizationRequestURI = ProcessInfo.processInfo.environment["INITIATE_VERIFICATION_URL"]
        XCTAssertTrue(authorizationRequestURI != "", "authorizationRequestURI:" + authorizationRequestURI!)
        let vpInteraction = Openid4vpInteraction(authorizationRequestURI, config: vpConfig)!

        let credentialsQuery = try vpInteraction.getQuery()
        let inquirer = CredentialNewInquirer(documentLoader)!

        let submissionRequirements = try inquirer.getSubmissionRequirements(
            credentialsQuery, contents: CredentialCredentialsOpt(issuedCreds))

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
