//
//  flutterPlugin.swift
//  Runner
//
//  Created by Avast.Inc on 2022-10-24.
//

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
            
        case "requestCredentialWithAuth":
            let redirectURIWithParams = fetchArgsKeyValue(call, key: "redirectURIWithParams")
            requestCredentialWithAuth(redirectURIWithParams: redirectURIWithParams!, result: result)

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
            
        case "activityLogger":
            storeActivityLogger(result:result)

        case "getVerifierDisplayData":
             getVerifierDisplayData(result:result)
            
        case "processAuthorizationRequest":
            processAuthorizationRequest(arguments: arguments!, result: result)

        case "getMatchedSubmissionRequirements":
            getMatchedSubmissionRequirements(arguments: arguments!, result: result)

        case "presentCredential":
            presentCredential(arguments: arguments!, result: result)
            
        case "wellKnownDidConfig":
            wellKnownDidConfig(arguments: arguments!, result: result)

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
        walletSDK.InitSDK(kmsStore: kmsStore(), didResolverURI: didResolverURI)

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
                    return result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while process authorization request",
                                             details: "no matching submission requirement is found"))
                
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
            
            let selectedCredentialsArray: VerifiableCredentialsArray?
            if (selectedCredentials != nil) {
                selectedCredentialsArray = convertToVerifiableCredentialsArray(credentials: selectedCredentials!)
            } else {
                guard let processAuthorizationRequestVCs = self.processAuthorizationRequestVCs else {
                    return  result(FlutterError.init(code: "NATIVE_ERR",
                                                     message: "error while process present credential",
                                                     details: "OpenID4VP interaction is not initialted"))
                }
                
                selectedCredentialsArray = processAuthorizationRequestVCs
            }

            try openID4VP.presentCredential(
                selectedCredentials: selectedCredentialsArray!)
            result(true);
            
        } catch OpenID4VPError.runtimeError(let errorMsg as NSError){
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while processing present credential",
                                     details: errorMsg.localizedDescription))
        } catch let error as NSError{
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
        } catch {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while creating did",
                                     details: error))
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
                
                if (scopes.length() == 0) {
                    authorizationLink = try openID4CI.createAuthorizationURL(clientID: clientID as! String, redirectURI: redirectURI as! String)
                } else {
                    authorizationLink = try openID4CI.createAuthorizationURLWithScopes(scopes: scopes, clientID: clientID as! String, redirectURI: redirectURI as! String)
                }
      
            }
            
            let flowTypeData :[String:Any] = [
                "pinRequired": pinRequired,
                "authorizationURLLink":  authorizationLink
                
            ]
            
            self.openID4CI = openID4CI
            
            result(flowTypeData)
            
          } catch {
              result(FlutterError.init(code: "NATIVE_ERR",
                                       message: "error while initializing issuance flow",
                                       details: error))
          }
    }
    
    public func initializeWalletInitiatedFlow(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        guard let issuerURI = arguments["issuerURI"] as? String else {
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while reading initializeWalletInitiatedFlow issuer URI",
                                             details: "Pass issuerURI as the arguments"))
        }
                
    
        do {
            let walletInitiatedOpenID4CI = try walletSDK?.createOpenID4CIWalletIntiatedInteraction(issuerURI: issuerURI)
            //TODO- VCS part is not implemented yet, this logic needs to be updated to get the create authorization URL
            try walletInitiatedOpenID4CI?.getSupportedCredentials()

        } catch {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while initializing wallet initiated issuance flow",
                                     details: error))
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
            var verifierDisplayData:[String:Any] = [
                "did":   verifierData.did(),
                "logoUri": verifierData.logoURI(),
                "name":  verifierData.name(),
                "purpose" : verifierData.purpose()
            ]

           result(verifierDisplayData)
    
        } catch {
        result(FlutterError.init(code: "NATIVE_ERR",
                                 message: "error while getting verifier display data",
                                 details: error))
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
    
    /**
     * ResolveDisplay resolves display information for issued credentials based on an issuer's metadata, which is fetched
       using the issuer's (base) URI. The CredentialDisplays returns DisplayData object correspond to the VCs passed in and are in the
       same order. This method requires one or more VCs and the issuer's base URI.
       IssuerURI and array of credentials  are parsed using VcparseParse to be passed to Openid4ciResolveDisplay which returns the resolved Display Data
     */
    
    public func serializeDisplayData(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        guard let openID4CI = self.openID4CI else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while resolve credential display",
                                             details: "openID4CI not initiated. Call authorize before this."))
        }
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
      
            let displayDataResp = openID4CI.serializeDisplayData(issuerURI: issuerURI,
                                                                     vcCredentials: convertToVerifiableCredentialsArray(credentials: vcCredentials))
            result(displayDataResp)
          } catch let error as NSError {
               return result(FlutterError.init(code: "Exception",
                                         message: "error while resolving credential",
                                         details: error.localizedDescription))
            }
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
        
        
        for i in 0...((displayData!.credentialDisplaysLength())-1){
            let credentialDisplay = displayData!.credentialDisplay(at: i)!
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
     
            let overview = credentialDisplay.overview()
            let logo = overview?.logo()
            
            var resolveDisplayResp : [String: Any] = [:]
            resolveDisplayResp["claims"] = claimList
            resolveDisplayResp["overviewName"] = overview?.name()
            resolveDisplayResp["logo"] = logo?.url()
            resolveDisplayResp["textColor"] = overview?.textColor()
            resolveDisplayResp["backgroundColor"] = overview?.backgroundColor()
            resolveDisplayResp["issuerName"] = issuerDisplayData!.name()
            
            resolvedCredDisplayList.append(resolveDisplayResp)
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
               print("error in validate result", error!.localizedDescription)
                var didValidateResultResp :[String:Any] = [
                    "isValid": false,
                    "serviceURL":  "",
                ]
                return result(didValidateResultResp)
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
        
        guard let vcCredentials = arguments["vcCredentials"] as? Array<String> else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while fetching issuer ID",
                                             details: "parameter storedcredentials is missed"))
        }
        let opts = VerifiableNewOpts()
        opts!.disableProofCheck()

        for cred in vcCredentials{
            let parsedVC = VerifiableParseCredential(cred, opts, nil)!
            let issuerID = parsedVC.issuerID()
            print("issuerid - function", issuerID)
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

