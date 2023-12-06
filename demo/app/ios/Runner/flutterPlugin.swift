/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Flutter
import UIKit
import Walletsdk

public class SwiftWalletSDKPlugin: NSObject, FlutterPlugin {
    
    public static func register(with registrar: FlutterPluginRegistrar) {
        let channel = FlutterMethodChannel(name: "WalletSDKPlugin", binaryMessenger: registrar.messenger())
        let instance = SwiftWalletSDKPlugin()
        registrar.addMethodCallDelegate(instance, channel: channel)
    }
   
    
    private var walletSDK: WalletSDK?
    private var openID4CI: OpenID4CI?
    private var openID4VP: OpenID4VP?
    private var walletInitiatedOpenID4CI: WalletInitiatedOpenID4CI?

    // TODO: remove next three variables after refactoring finished.
    private var processAuthorizationRequestVCs: VerifiableCredentialsArray?
    private var didDocResolution: ApiDIDDocResolution?
   
    public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
        let arguments = call.arguments as? Dictionary<String, Any>
        
        switch call.method {
        case "createDID":
            createDid(arguments: arguments!, result: result)
            
        case "initialize":
            initialize(arguments: arguments!, result: result)
            
        case "initializeWalletInitiatedFlow":
            initializeWalletInitiatedFlow(arguments: arguments!, result: result)
            
        case "requestCredential":
            let otp = fetchArgsKeyValue(call, key: "otp")
            requestCredential(otp: otp!, result: result)
            
        case "parseWalletSDKError":
            let localizedErrorMessage = fetchArgsKeyValue(call, key: "localizedErrorMessage")
            parseWalletSDKError(localizedErrorMessage: localizedErrorMessage!, result: result)
            
        case "requestCredentialWithAuth":
            let redirectURIWithParams = fetchArgsKeyValue(call, key: "redirectURIWithParams")
            requestCredentialWithAuth(redirectURIWithParams: redirectURIWithParams!, result: result)
            
        case "requestCredentialWithWalletInitiatedFlow":
            let redirectURIWithParams = fetchArgsKeyValue(call, key: "redirectURIWithParams")
            requestCredentialWithWalletInitiatedFlow(redirectURIWithParams: redirectURIWithParams!, result: result)

        case "fetchDID":
            let didID = fetchArgsKeyValue(call, key: "didID")
            
        case "credentialStatusVerifier":
            credentialStatusVerifier(arguments: arguments!,  result: result)

        case "serializeDisplayData":
            serializeDisplayData(arguments: arguments!,  result: result)

        case "resolveCredentialDisplay":
            resolveCredentialDisplay(arguments: arguments!, result: result)

        case "getVersionDetails":
           getVersionDetails(result:result)

        case "getCredID":
            getCredID(arguments: arguments!,  result: result)

        case "parseActivities":
            parseActivities(arguments: arguments!,  result: result)

        case "initSDK":
            initSDK(arguments: arguments!, result:result)

        case "issuerURI":
            issuerURI(result:result)
            
        case "getIssuerID":
             getIssuerID(arguments: arguments!, result:result)
            
        case "getIssuerMetaData":
            getIssuerMetaData(arguments: arguments!, result:result)
            
        case "activityLogger":
            storeActivityLogger(result:result)
            
        case "verifyIssuer":
            verifyIssuer(result:result)

        case "getVerifierDisplayData":
             getVerifierDisplayData(result:result)
            
        case "processAuthorizationRequest":
            processAuthorizationRequest(arguments: arguments!, result: result)

        case "getMatchedSubmissionRequirements":
            getMatchedSubmissionRequirements(arguments: arguments!, result: result)
            
        case "getCustomScope":
            getCustomScope(result: result)

        case "presentCredential":
            presentCredential(arguments: arguments!, result: result)
            
        case "wellKnownDidConfig":
            wellKnownDidConfig(arguments: arguments!, result: result)
        
        case "createAuthorizationURLWalletInitiatedFlow":
            createAuthorizationURLWalletInitiatedFlow(arguments: arguments!, result: result)
            
        case "requireAcknowledgment":
            requireAcknowledgment(result: result)
        case "acknowledgeSuccess":
            acknowledgeSuccess(result: result)
        case "acknowledgeReject":
            acknowledgeReject(result: result)

        default:
            print("No call method is found")
        }
    }

    private func initSDK(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        guard let didResolverURI = arguments["didResolverURI"] as? String else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while initiating SDK",
                                             details: "parameter didResolverURI is missed"))
        }
        
        let walletSDK = WalletSDK();
        walletSDK.initSDK(kmsStore: kmsStore(), didResolverURI: didResolverURI)

        self.walletSDK = walletSDK
        result(true)
    }
   /**
    This method gets the version detail if we build sdk using the env variable
    For Example: NEW_VERSION=testVer GIT_REV=testRev BUILD_TIME=testTime make generate-ios-bindings copy-ios-bindings
    */
     public func getVersionDetails(result: @escaping FlutterResult) {
      var versionResp : [String: Any] = [:]
      versionResp["walletSDKVersion"] = VersionGetVersion()
      versionResp["gitRevision"] = VersionGetGitRevision()
      versionResp["buildTimeRev"] = VersionGetBuildTime()
      result(versionResp)
    }

    /**
     This method  invoke processAuthorizationRequest defined in OpenID4Vp file.
     */
    public func processAuthorizationRequest(arguments: Dictionary<String, Any> , result: @escaping FlutterResult) {
        do {
            guard let walletSDK = self.walletSDK else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process authorization request",
                                                 details: "WalletSDK interaction is not initialized, call initSDK()"))
            }
            
            guard let authorizationRequest = arguments["authorizationRequest"] as? String else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process authorization request",
                                                 details: "parameter authorizationRequest is missed"))
            }
            

            let storedCredentials = arguments["storedCredentials"] as? Array<String>
            
            let openID4VP = try walletSDK.createOpenID4VPInteraction()

            self.openID4VP = openID4VP

            try openID4VP.startVPInteraction(authorizationRequest: authorizationRequest)
            
            if (storedCredentials != nil) {
                processAuthorizationRequestVCs = convertToVerifiableCredentialsArray(credentials: storedCredentials!)
              let matchedReq = try openID4VP.getMatchedSubmissionRequirements(
                    storedCredentials:convertToVerifiableCredentialsArray(
                    credentials: storedCredentials!))
                var resp = convertVerifiableCredentialsArray(arr: matchedReq.atIndex(0)!.descriptor(at:0)!.matchedVCs!)
                if (resp.isEmpty) {
                    var typeConstraint = matchedReq.atIndex(0)!.descriptor(at:0)!.typeConstraint()
                    if typeConstraint == "" {
                      var schemas =  matchedReq.atIndex(0)!.descriptor(at:0)!.schemas()
                        var schemaList:[Any] = []
                        if let schemas = schemas {
                            var schemasResp : [String: Any] = [:]
                            for index in 0..<schemas.length() {
                                if let schema = schemas.atIndex(index) {
                                    schemasResp["required"] = schema.required()
                                    schemasResp["uri"] = schema.uri()
                                }
                                
                                schemaList.append(schemasResp)
                            }
                            
                            return result(FlutterError.init(code: "NATIVE_ERR",
                                                            message: "No credentials conforming to the following schemas were found",
                                                            details: "\(schemaList)"))
                            
                        }
                    }
                    return result(FlutterError.init(code: "NATIVE_ERR",
                                                    message: "No credentials of type \(typeConstraint) were found",
                                                    details: "Required credential \(typeConstraint) is missing from the wallet"))
                    
                }

                result(resp)
            }
            return result(Array<String>())
            
        } catch OpenID4VPError.runtimeError(let errorMsg){
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while process authorization request",
                                     details: errorMsg))
        } catch let error as NSError {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while process authorization request",
                                     details: error.description))
        }
    }
    
    public func getMatchedSubmissionRequirements(arguments: Dictionary<String, Any> , result: @escaping FlutterResult) {
        do {
            guard let openID4VP = self.openID4VP else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while getting matched submission requirements",
                                                 details: "OpenID4VP interaction is not initialted"))
            }

            guard let storedCredentials = arguments["storedCredentials"] as? Array<String> else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while getting matched submission requirements",
                                                 details: "parameter storedCredentials is missed"))
            }
            
 
         
            let matchResult = try convertSubmissionRequirementArray(
                requirements: try openID4VP.getMatchedSubmissionRequirements(storedCredentials:  convertToVerifiableCredentialsArray(credentials:storedCredentials)))
            
            return result(matchResult)
            
        } catch OpenID4VPError.runtimeError(let errorMsg){
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while getting matched submission requirements",
                                     details: errorMsg))
        } catch let error as NSError {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while getting matched submission requirements",
                                     details: error.description))
        }
    }
    
    public func getCustomScope(result: @escaping FlutterResult){
        do {
            guard let openID4VP = self.openID4VP else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process present credential",
                                                 details: "OpenID4VP interaction is not initialted"))
            }
            
            let customScopeList = try openID4VP.getCustomScope()
            result(customScopeList)
        } catch let error as NSError{
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while getting custom scope ",
                                     details: error.localizedDescription))
        }
    }
    
    /**
     This method invokes presentCredentialt defined in OpenID4Vp file.
     */
    public func presentCredential(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        do {
            guard let openID4VP = self.openID4VP else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process present credential",
                                                 details: "OpenID4VP interaction is not initialted"))
            }
            
            let selectedCredentials = arguments["selectedCredentials"] as? Array<String>
            let customScopeList = arguments["customScopeList"] as? Dictionary<String, Any> ?? [String: Any]()
            
            let selectedCredentialsArray: VerifiableCredentialsArray?
            if (selectedCredentials != nil) {
                selectedCredentialsArray = convertToVerifiableCredentialsArray(credentials: selectedCredentials!)
            } else {
                guard let processAuthorizationRequestVCs = self.processAuthorizationRequestVCs else {
                    return  result(FlutterError.init(code: "NATIVE_ERR",
                                                     message: "error while process present credential",
                                                     details: "OpenID4VP interaction is not initiated"))
                }
                
                selectedCredentialsArray = processAuthorizationRequestVCs
            }

            try openID4VP.presentCredential(
                selectedCredentials: selectedCredentialsArray!, customScopes: customScopeList)
            result(true);
            
        }  catch let error as NSError{
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while processing present credential",
                                     details: error.localizedDescription))
        }
    }
    
    /**
     Create method of DidNewCreator creates a DID document using the given DID method.
     The usage of ApiCreateDIDOpts depends on the DID method you're using.
     In the app when user logins we invoke sdk DidNewCreator create method to create new did per user.
     */
    public func createDid(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        do {
            guard let walletSDK = self.walletSDK else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while creating did",
                                                 details: "WalletSDK interaction is not initialized, call initSDK()"))
            }
            
            guard let didMethodType = arguments["didMethodType"] as? String else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while create did operation",
                                                 details: "parameter didMethodType is missed"))
            }

            guard let didKeyType = arguments["didKeyType"] as? String else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while create did operation",
                                                 details: "parameter didKeyType is missed"))
            }
            
            let doc = try walletSDK.createDID(didMethodType: didMethodType, didKeyType: didKeyType)
            didDocResolution = doc
            var docResolution : [String: Any] = [:]
            docResolution["did"] = doc.id_(nil)
            docResolution["didDoc"] = doc.content
            result(docResolution)
        } catch let error as NSError {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while creating did",
                                     details: error.localizedDescription))
        }
    }
    
    /**
     *Initialize method of Openid4ciNewInteraction is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
     After initializing the Interaction object with an Issuance Request, this should be the first method you call in
     order to continue with the flow.
     */
    public func initialize(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        guard let walletSDK = self.walletSDK else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while creating new OIDC interaction",
                                             details: "WalletSDK interaction is not initialized, call initSDK()"))
        }
        
        guard let requestURI = arguments["requestURI"] as? String else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while reading requestURI",
                                             details: "parameter requestURI is missed"))
        }

        do {
            
            let openID4CI = try walletSDK.createOpenID4CIInteraction(requestURI:requestURI)
            var pinRequired = false
            var authorizationLink = ""
            
            let flowType = try openID4CI.checkFlow()
            if (flowType == "preauth-code-flow") {
                pinRequired = try openID4CI.pinRequired()
            }
            
            if (flowType == "auth-code-flow"){
                guard let authCodeArgs = arguments["authCodeArgs"] as? Dictionary<String, Any> else {
                    return  result(FlutterError.init(code: "NATIVE_ERR",
                                                     message: "error while reading auth code arguments",
                                                     details: "Pass scopes, clientID and redirectURI as the arguments"))
                }
                
                let dynamicRegistrationSupported = try openID4CI.dynamicRegistrationSupported()
                var clientID = authCodeArgs["clientID"]
                let redirectURI = authCodeArgs["redirectURI"]! as? String
                var scopesFromArgs = authCodeArgs["scopes"]! as! [String]
                var oauthDiscoverableClientURI = authCodeArgs["oauthDiscoverableClientURI"] as? String
                var scopes = ApiNewStringArray()!
                
                for scope in scopesFromArgs {
                    scopes.append(scope)
                }
                
                if (dynamicRegistrationSupported.boolValue == true) {
                    let dynamicRegistrationEndpoint = try openID4CI.dynamicRegistrationEndpoint()
                    let clientMetadata = Oauth2ClientMetadata()
                    let grantTypesArr = ApiStringArray()
                    grantTypesArr?.append("authorization_code")
                    clientMetadata?.setGrantTypes(grantTypesArr)
                
                    
                    let redirectURIArr = ApiStringArray()
                    redirectURIArr?.append(redirectURI)
                    clientMetadata?.setRedirectURIs(redirectURIArr)
            
                    clientMetadata?.setScopes(scopes)
                    clientMetadata?.setTokenEndpointAuthMethod("none")
                    
                    let authorizationCodeGrantParams = try openID4CI.getAuthorizationCodeGrantParams()
                    
                    if authorizationCodeGrantParams.hasIssuerState() {
                        let issuerState = authorizationCodeGrantParams.issuerState(nil)
                        clientMetadata?.setIssuerState(issuerState)
                    }
                    
                    let registrationResp = Oauth2RegisterClient(dynamicRegistrationEndpoint, clientMetadata, nil, nil)!
                    clientID = (registrationResp.clientID())
                    
                    // Use the actual scopes registered by the authorization server,
                    // which may differ from the scopes we specified in the metadata in our request.
                    scopes = registrationResp.registeredMetadata()!.scopes()!
                }
                
                authorizationLink = try openID4CI.createAuthorizationURL(clientID: clientID! as! String, redirectURI: redirectURI!, oauthDiscoverableClientURI: oauthDiscoverableClientURI ?? "", scopes: scopes)
      
            }
            
            let flowTypeData :[String:Any] = [
                "pinRequired": pinRequired,
                "authorizationURLLink":  authorizationLink
                
            ]
            
            self.openID4CI = openID4CI
            
            result(flowTypeData)
            
          } catch let error as NSError {
              result(FlutterError.init(code: "Exception",
                                       message: "error while initializing issuance flow",
                                       details: error.localizedDescription))
          }
    }
    
    public func verifyIssuer(result: @escaping FlutterResult) {
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while getting issuer meta data",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        do {
            let issuerVerified = try openID4CI.verifyIssuer()
            print(issuerVerified)
            result(issuerVerified)
        } catch let error as NSError {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while verifying the issuer",
                                     details: error.localizedDescription))
        }

        
    }
    
    public func getIssuerMetaData(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while getting issuer meta data",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        
        guard let credentialTypes = arguments["credentialTypes"] as? Array<String> else{
                  return  result(FlutterError.init(code: "NATIVE_ERR",
                                                   message: "error while reading credentialTypes",
                                                   details: "parameter credential types is missed"))
              }
        
        do {
            
            var issuerMetaData = try openID4CI.getIssuerMetadata()
            // credential issuer
            var credIssuer =  issuerMetaData.credentialIssuer()
            
            // supported Credentials list
            var supportedCredentials = issuerMetaData.supportedCredentials()!
            var supportedCredentialsList = getSupportedCredentialsList(supportedCredentials: supportedCredentials, credentialTypes: credentialTypes)
            
            // localized issuer displays data
            var localizedIssuerDisplays =  issuerMetaData.localizedIssuerDisplays()!
            var localizedIssuerDisplayList :[Any] = []
            for index in 0..<localizedIssuerDisplays.length() {
                let localizedIssuerDisplay :[String:Any] = [
                    "name":  localizedIssuerDisplays.atIndex(index)!.name(),
                    "locale": localizedIssuerDisplays.atIndex(index)!.locale(),
                    "url":   localizedIssuerDisplays.atIndex(index)!.url(),
                    "logo":  localizedIssuerDisplays.atIndex(index)?.logo()?.url() ?? "",
                    "textColor": localizedIssuerDisplays.atIndex(index)!.textColor(),
                    "backgroundColor": localizedIssuerDisplays.atIndex(index)!.backgroundColor()
                ]
                
                localizedIssuerDisplayList.append(localizedIssuerDisplay)
            }
           
        
            var issuerMetaDataRespList: [Any] = []
            let issuerMetaDataResp:[String:Any] = [
                "credentialIssuer": credIssuer,
                "supportedCredentials" : supportedCredentialsList,
                "localizedIssuerDisplays":localizedIssuerDisplayList
            ]
            
            issuerMetaDataRespList.append(issuerMetaDataResp)
            print(issuerMetaDataRespList)
            result(issuerMetaDataRespList)
            
        } catch let error as NSError {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while getting issuer meta data",
                                     details: error.localizedDescription))
        }
        
    }
    
    public func getSupportedCredentialsList(supportedCredentials: Openid4ciSupportedCredentials, credentialTypes: [String]) -> [Any] {
        var supportedCredentialsList: [Any] = []
        for index in 0..<supportedCredentials.length() {
            var typeStrArray = [String]()
            for i in 0..<(supportedCredentials.atIndex(index)!.types()?.length())!{
                let type = supportedCredentials.atIndex(index)!.types()?.atIndex(i)
                for credType in credentialTypes {
                    if (credType == type){
                        typeStrArray.append(type!)
                        typeStrArray.append("VerifiableCredential")
                        
                        var localizedCredentialsDisplayRespList: [Any] = []
                        for i in 0..<(supportedCredentials.atIndex(index)!.localizedDisplays()?.length())!{
                            let localizedCredentialsDisplayResp :[String:Any] = [
                                "name":  supportedCredentials.atIndex(index)!.localizedDisplays()!.atIndex(i)!.name(),
                                "locale": supportedCredentials.atIndex(index)!.localizedDisplays()!.atIndex(i)!.locale(),
                                "logo": supportedCredentials.atIndex(index)!.localizedDisplays()!.atIndex(i)!.logo()!.url(),
                                "textColor": supportedCredentials.atIndex(index)!.localizedDisplays()!.atIndex(i)!.textColor(),
                                "backgroundColor": supportedCredentials.atIndex(index)!.localizedDisplays()!.atIndex(i)!.backgroundColor()
                            ]
                            localizedCredentialsDisplayRespList.append(localizedCredentialsDisplayResp)
                        }
                        
                        let supportedCredentialResp:[String:Any] = [
                            "format":   supportedCredentials.atIndex(index)!.format(),
                            "types":    typeStrArray,
                            "display":   localizedCredentialsDisplayRespList
                        ]
                        
                        supportedCredentialsList.append(supportedCredentialResp)
                    }
                }
            }
        }
        
        return supportedCredentialsList
    }
    
    public func initializeWalletInitiatedFlow(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        guard let issuerURI = arguments["issuerURI"] as? String else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while reading initializeWalletInitiatedFlow issuer URI",
                                             details: "Pass issuerURI as the arguments"))
        }
        
        guard let credentialTypes = arguments["credentialTypes"] as? Array<String> else{
                  return  result(FlutterError.init(code: "NATIVE_ERR",
                                                   message: "error while reading credentialTypes",
                                                   details: "parameter credential types is missed"))
              }
                
        guard let walletSDK = self.walletSDK else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while creating new OIDC interaction",
                                             details: "WalletSDK interaction is not initialized, call initSDK()"))
        }
        
        do {
            let walletInitiatedOpenID4CI = try walletSDK.createOpenID4CIWalletInitiatedInteraction(issuerURI: issuerURI)
            let supportedCredentials = try walletInitiatedOpenID4CI.getSupportedCredentials()
            
            var supportedCredentialsList = getSupportedCredentialsList(supportedCredentials: supportedCredentials, credentialTypes: credentialTypes)
       
            self.walletInitiatedOpenID4CI = walletInitiatedOpenID4CI
            result(supportedCredentialsList)
        
        }  catch let error as NSError {
            result(FlutterError.init(code: "Exception",
                                     message: "error while initializing wallet initiated issuance flow",
                                     details: error.localizedDescription))
        }
      
        
    }
     
    public func createAuthorizationURLWalletInitiatedFlow(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        guard let walletInitiatedOpenID4CI = self.walletInitiatedOpenID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "walletInitiatedOpenID4CI interaction is not initialized",
                                             details: ""))
        }
        
        guard let scopesFromArgs = arguments["scopes"] as? Array<String> else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "Missing scopes argument",
                                             details: ""))
        }
        
        guard let credentialTypesFromArgs = arguments["credentialTypes"] as? Array<String> else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "Missing credentialTypes argument",
                                             details: ""))
        }
        
        guard let clientID = arguments["clientID"] as? String else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "Missing clientID argument", details: ""))
        }
        
        guard let redirectURI = arguments["redirectURI"] as? String else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "Missing redirectURI argument",
                                             details: ""))
        }
        
        
        guard let credentialFormat = arguments["credentialFormat"] as? String else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "Missing credentialFormat argument",
                                             details: ""))
        }
        
        guard let issuerURI = arguments["issuerURI"] as? String else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "Missing issuerURI argument",
                                             details: ""))
        }
        
        var scopes = ApiNewStringArray()!
        
        for scope in scopesFromArgs {
            scopes.append(scope)
        }
        
        
        var credentialTypes = ApiNewStringArray()!
        
        for credentialType in credentialTypesFromArgs {
            credentialTypes.append(credentialType)
        }

        do {
            let authorizationURL = try walletInitiatedOpenID4CI.createAuthorizationURLWalletInitiatedFlow(scopes: scopes, credentialFormat: credentialFormat, credentialTypes: credentialTypes, clientID: clientID, redirectURI: redirectURI, issuerURI: issuerURI)
            
             result(authorizationURL)
        } catch let error as NSError {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while creating authorization URL in wallet initiated issuance flow",
                                     details: error.localizedDescription))
        }
    }
    
    public func getVerifierDisplayData(result: @escaping FlutterResult){
        guard let openID4VP = self.openID4VP else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process present credential",
                                             details: "OpenID4VP interaction is not initialted"))
        }
        do {
           let verifierData =  try openID4VP.getVerifierDisplayData()
            let verifierDisplayData:[String:Any] = [
                "did":   verifierData.did(),
                "logoUri": verifierData.logoURI(),
                "name":  verifierData.name(),
                "purpose" : verifierData.purpose()
            ]

           result(verifierDisplayData)
    
        } catch let error as NSError {
        result(FlutterError.init(code: "NATIVE_ERR",
                                 message: "error while getting verifier display data",
                                 details: error.localizedDescription))
       }
        
    }
    
    /**
    * RequestCredential method of Openid4ciNewInteraction is the final step in the
    interaction. This is called after the wallet is authorized and is ready to receive credential(s).
    
    Here if the pin required is true in the authorize method, then user need to enter OTP which is intercepted to create CredentialRequest Object using
    Openid4ciNewCredentialRequestOpt.
     If flow doesnt not require pin than Credential Request Opts will have empty string otp and sdk will return credential Data based on empty otp.
    */
    public func requestCredential(otp: String, result: @escaping FlutterResult){
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process requestCredential credential",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        
        guard let didDocResolution = self.didDocResolution else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process requestCredential credential",
                                             details: "Did document not initialized"))
        }
                
         do {
            let credentialCreated = try openID4CI.requestCredential(didVerificationMethod: didDocResolution.assertionMethod(), otp: otp)
            result(credentialCreated.serialize(nil))
          } catch let error as NSError{
             return result(FlutterError.init(code: "Exception",
                                       message: "error while requesting credential",
                                       details: error.localizedDescription
                                      ))
          }
        
    }
    
    public func requireAcknowledgment(result: @escaping FlutterResult){
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error in require acknowlwdgement",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        
        do {
            let ackResp = try openID4CI.requireAcknowledgment()
            result(ackResp.boolValue)
        } catch let error as NSError {
            return result(FlutterError.init(code: "Exception",
                                      message: "error in require acknowlwdgement",
                                      details: error.localizedDescription
                                     ))
        }
        
    }
    
    public func acknowledgeSuccess(result: @escaping FlutterResult) {
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error in acknowledge Success",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        do {
            let ackSuccessResp = try openID4CI.acknowledgeSuccess()
            result(ackSuccessResp.boolValue)
        } catch let error as NSError {
            return result(FlutterError.init(code: "Exception",
                                      message: "error in acknowledge Success",
                                      details: error.localizedDescription
                                     ))
        }
    }
    
    
    public func acknowledgeReject(result: @escaping FlutterResult) {
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error in acknowledge reject",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        do {
            let ackRejectResp = try openID4CI.acknowledgeReject()
            result(ackRejectResp.boolValue)
        } catch let error as NSError {
            return result(FlutterError.init(code: "Exception",
                                      message: "error in acknowledge reject",
                                      details: error.localizedDescription
                                     ))
        }
    }
    
    public func requestCredentialWithAuth(redirectURIWithParams: String, result: @escaping FlutterResult){
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process requestCredential auth",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        
        guard let didDocResolution = self.didDocResolution else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process requestCredential credential",
                                             details: "Did document not initialized"))
        }
         do {
             let credentialCreated = try openID4CI.requestCredentialWithAuth(didVerificationMethod: didDocResolution.assertionMethod(), redirectURIWithParams: redirectURIWithParams)
            result(credentialCreated.serialize(nil))
          } catch let error as NSError{
             return result(FlutterError.init(code: "Exception",
                                       message: "error while requesting credential with auth",
                                             details: error.localizedDescription
                                      ))
          }
        
    }
    
    
    
    public func requestCredentialWithWalletInitiatedFlow(redirectURIWithParams: String, result: @escaping FlutterResult){
        guard let walletInitiatedOpenID4CI = self.walletInitiatedOpenID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process requestCredential with wallet initiated flow",
                                             details: "walletInitiatedOpenID4CI not initiated"))
        }
        
        guard let didDocResolution = self.didDocResolution else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "DID document not initialized",
                                             details: ""))
        }
                
         do {
             let credentialCreated = try walletInitiatedOpenID4CI.requestCredentialWithWalletInitiatedFlow(didVerificationMethod: didDocResolution.assertionMethod(), redirectURIWithParams: redirectURIWithParams)
            result(credentialCreated.serialize(nil))
          } catch let error as NSError{
             return result(FlutterError.init(code: "Exception",
                                       message: "error while requesting credential with auth - wallet initiated flow",
                                       details: error.localizedDescription
                                      ))
          }
        
    }
    
    /**
     * ResolveDisplay resolves display information for issued credentials based on an issuer's metadata, which is fetched
       using the issuer's (base) URI. The CredentialDisplays returns DisplayData object correspond to the VCs passed in and are in the
       same order. This method requires one or more VCs and the issuer's base URI.
       IssuerURI and array of credentials  are parsed using VcparseParse to be passed to Openid4ciResolveDisplay which returns the resolved Display Data
     */
    
    public func serializeDisplayData(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        do {
            guard let issuerURI = arguments["uri"] as? String else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while resolve credential display",
                                                 details: "parameter issuerURI is missed"))
            }

            guard let vcCredentials = arguments["vcCredentials"] as? Array<String> else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while resolve credential display",
                                                 details: "parameter storedcredentials is missed"))
            }
            
            let resolvedDisplayData = DisplayResolve(convertToVerifiableCredentialsArray(credentials: vcCredentials), issuerURI, nil, nil)
            let resolvedDisplayDataString = resolvedDisplayData!.serialize(nil)
        
            result(resolvedDisplayDataString)
          } catch let error as NSError {
              let parsedError = WalleterrorParse(error.localizedDescription)
               return result(FlutterError.init(code: "Exception",
                                         message: "error while resolving credential",
                                               details: parsedError))
              
            }
    }

    
    public func parseWalletSDKError(localizedErrorMessage: String, result: @escaping FlutterResult){
        let parsedError = WalleterrorParse(localizedErrorMessage)!
        
        var parsedErrorResult :[String: Any] = [
            "category":   parsedError.category,
            "details":  parsedError.details,
            "code":  parsedError.code,
            "traceID":  parsedError.traceID
        ]
        result(parsedErrorResult)
    }
    
    
    public func resolveCredentialDisplay(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){

        
        guard let resolvedCredentialDisplayData = arguments["resolvedCredentialDisplayData"] as? String else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while resolveCredentialDisplay",
                                             details: "parameter resolvedCredentialDisplayData is missed"))
        }
        let displayData = DisplayParseData(resolvedCredentialDisplayData, nil)
        let issuerDisplayData = displayData?.issuerDisplay()
        
        var resolvedCredDisplayList : [Any] = []
        var claimList:[Any] = []
        
        if(displayData!.credentialDisplaysLength() != 0){
            for i in 0...((displayData!.credentialDisplaysLength())-1){
                let credentialDisplay = displayData!.credentialDisplay(at: i)!
                if(credentialDisplay.claimsLength() != 0){
                    for i in 0...(credentialDisplay.claimsLength())-1{
                        let claim = credentialDisplay.claim(at: i)!
                        var claims : [String: Any] = [:]
                        if claim.isMasked(){
                            claims["value"] = claim.value()
                            claims["rawValue"] = claim.rawValue()
                        }
                        var order: Int = -1
                        if claim.hasOrder() {
                            do {
                                try claim.order(&order)
                                claims["order"] = order
                            } catch let err as NSError {
                                print("Error: \(err)")
                            }
                        }
                        claims["rawValue"] = claim.rawValue()
                        claims["valueType"] = claim.valueType()
                        claims["label"] = claim.label()
                        claimList.append(claims)
                    }
                }
                
                let overview = credentialDisplay.overview()
                let logo = overview?.logo()
                
                var resolveDisplayResp : [String: Any] = [:]
                resolveDisplayResp["claims"] = claimList
                resolveDisplayResp["overviewName"] = overview?.name()
                resolveDisplayResp["logo"] = logo?.url()
                resolveDisplayResp["textColor"] = overview?.textColor()
                resolveDisplayResp["backgroundColor"] = overview?.backgroundColor()
                resolveDisplayResp["issuerName"] =  issuerDisplayData?.name()

                
                resolvedCredDisplayList.append(resolveDisplayResp)
            }
        }
        
       
        result(resolvedCredDisplayList)
    }
    
    public func credentialStatusVerifier(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        
        guard let credentials = arguments["credentials"] as? Array<String> else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while getting get credential status verifier",
                                             details: "parameter credentials is missed"))
        }
        do {
            let statusVerifier = CredentialNewStatusVerifier(nil, nil);
            let credentialArray = convertToVerifiableCredentialsArray(credentials: credentials)
            try statusVerifier?.verify(credentialArray.atIndex(0))
            result(true)
         } catch let error as NSError{
             result(FlutterError.init(code: "Exception",
                                      message: "error while getting get credential status verifier",
                                      details: error.localizedDescription))
         }

        
    }
    
    public func wellKnownDidConfig(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        guard let issuerID = arguments["issuerID"] as? String else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while getting well known did config",
                                             details: "issuer id is missing"))
            }
        
            let didResolver = DidNewResolver(nil, nil)
            var error: NSError?

            var didValidateResult = DidValidateLinkedDomains(issuerID, didResolver, nil, &error)
           
            if let actualError = error {
               return result(FlutterError.init(code: "Exception",
                                         message: "error while validating linked domains",
                                         details: actualError.localizedDescription))
              }

            var didValidateResultResp :[String:Any] = [
                "isValid":   didValidateResult!.isValid,
                "serviceURL":  didValidateResult!.serviceURL,
            ]

           return result(didValidateResultResp)
    
    }

    /**
     ApiParseActivity is invoked to parse the list of activities which are stored in the app when we issue and present credential,
     */
    public func parseActivities(arguments: Dictionary<String, Any>,result: @escaping FlutterResult){
        var activityList: [Any] = []
        guard let activities = arguments["activities"] as? Array<String> else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while parsing activities",
                                             details: "parameter activities is missing"))
        }
                
        for activity in activities {
            let activityObj = ApiParseActivity(activity, nil)
            var status = activityObj!.status()
            var date = NSDate(timeIntervalSince1970: TimeInterval(activityObj!.unixTimestamp()))
            var utcDateFormatter = DateFormatter()
            utcDateFormatter.dateStyle = .long
            utcDateFormatter.timeStyle = .short
            let updatedDate = date
            var activityDicResp:[String:Any] = [
                "Status":  status,
                "Issued By": activityObj?.client(),
                "Operation": activityObj?.operation(),
                "Activity Type": activityObj?.type(),
                "Date": utcDateFormatter.string(from: updatedDate as Date),
            ]
            activityList.append(activityDicResp)
        }
    

        result(activityList)
    }
    
    /**
     Local function to fetch all activities and send the serialized response to the app to be stored in the flutter secure storage.
     */
    public func storeActivityLogger(result: @escaping FlutterResult){
        guard let walletSDK = self.walletSDK else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while authorize ci",
                                             details: "WalletSDK interaction is not initialized, call initSDK()"))
        }
        
        var activityList: [Any] = []
        let aryLength = walletSDK.activityLogger!.length()
        for index in 0..<aryLength {
            activityList.append(walletSDK.activityLogger!.atIndex(index)!.serialize(nil))
        }

        result(activityList)
    }
    /**
     Local function  to get the credential IDs of the requested credentials.
     */
    public func getCredID(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        
        guard let vcCredentials = arguments["vcCredentials"] as? Array<String> else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while fetching credential ID",
                                             details: "parameter storedcredentials is missed"))
        }
        let opts = VerifiableNewOpts()
        opts!.disableProofCheck()
        
        var credIDs: [Any] = []

        for cred in vcCredentials{
            let parsedVC = VerifiableParseCredential(cred, opts, nil)!
            let credID = parsedVC.id_()
            credIDs.append(credID)
            
        }
        result(credIDs[0])
    }
    
    public func getIssuerID(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        
        guard let vcCredentials = arguments["credentials"] as? Array<String> else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while fetching issuer ID",
                                             details: "parameter credentials are missing"))
        }
        let opts = VerifiableNewOpts()
        opts!.disableProofCheck()

        for cred in vcCredentials{
            let parsedVC = VerifiableParseCredential(cred, opts, nil)!
            let issuerID = parsedVC.issuerID()
            return result(issuerID)
        }
    }
    /**
     * IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
        there's a later need to refresh credential display data using the latest display information from the issuer.
     */
    public func issuerURI( result: @escaping FlutterResult){
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while calling issuerURI",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
        
        let issuerURIResp = openID4CI.issuerURI();
        result(issuerURIResp)
    }

    
    public func fetchArgsKeyValue(_ call: FlutterMethodCall, key: String) -> String? {
        guard let args = call.arguments else {
            return ""
        }
        let myArgs = args as? [String: Any];
        return myArgs?[key] as? String;
    }
  
}

