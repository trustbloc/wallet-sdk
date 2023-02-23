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
    
    struct qrCodeData{
        static var requestURI = ""
    }
    
    private var kms:LocalkmsKMS?
    private var didResolver: ApiDIDResolverProtocol?
    private var documentLoader: ApiLDDocumentLoaderProtocol?
    private var crypto: ApiCryptoProtocol?
    private var didDocRes: ApiDIDDocResolution?
    private var didDocID: String?
    private var newOIDCInteraction: Openid4ciInteraction?
    private var didVerificationMethod: ApiVerificationMethod?
    private var activityLogger: MemActivityLogger?

    private var openID4VP: OpenID4VP?
    
    public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
        let arguments = call.arguments as? Dictionary<String, Any>
        
        switch call.method {
        case "createDID":
            let didMethodType = fetchArgsKeyValue(call, key: "didMethodType")
            createDid(didMethodType: didMethodType!, result: result)
            
        case "authorize":
            let requestURI = fetchArgsKeyValue(call, key: "requestURI")
            qrCodeData.requestURI = requestURI!
            authorize(requestURI: requestURI!, result: result)
            
        case "requestCredential":
            let otp = fetchArgsKeyValue(call, key: "otp")
            requestCredential(otp: otp!, result: result)

        case "fetchDID":
            let didID = fetchArgsKeyValue(call, key: "didID")
            if didDocID == nil {
                didDocID = didID
            }

        case "resolveCredentialDisplay":
            resolveCredentialDisplay(arguments: arguments!,  result: result)
            
        case "getCredID":
            getCredID(arguments: arguments!,  result: result)
            
        case "parseActivities":
            parseActivities(arguments: arguments!,  result: result)
            
        case "initSDK":
            initSDK(result:result)
            
        case "issuerURI":
            issuerURI(result:result)

        case "activityLogger":
            storeActivityLogger(result:result)

        case "processAuthorizationRequest":
            processAuthorizationRequest(arguments: arguments!, result: result)
            
        case "presentCredential":
            presentCredential(result: result)
            
        default:
            print("No call method is found")
        }
    }
    
    private func initSDK(result: @escaping FlutterResult) {
        let kmsstore = kmsStore()
        kms = LocalkmsNewKMS(kmsstore, nil)
        didResolver = DidNewResolver("http://did-resolver.trustbloc.local:8072/1.0/identifiers", nil)
        crypto = kms?.getCrypto()
        documentLoader = LdNewDocLoader()
        activityLogger = MemNewActivityLogger()
        result(true)
    }
    
    private func createOpenID4VP() throws -> OpenID4VP {
        guard let kms = self.kms else {
            throw OpenID4VPError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let crypto = self.crypto else {
            throw OpenID4VPError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let didResolver = self.didResolver else {
            throw OpenID4VPError.runtimeError("SDK is not initialized, call initSDK()")
        }
        guard let documentLoader = self.documentLoader else {
            throw OpenID4VPError.runtimeError("SDK is not initialized, call initSDK()")
        }
        
        return OpenID4VP(keyReader: kms, didResolver: didResolver, documentLoader: documentLoader, crypto: crypto, activityLogger: activityLogger!)
    }
    
    public func processAuthorizationRequest(arguments: Dictionary<String, Any> , result: @escaping FlutterResult) {
        do {
            
            guard let authorizationRequest = arguments["authorizationRequest"] as? String else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process authorization request",
                                                 details: "parameter authorizationRequest is missed"))
            }
            
            guard let storedCredentials = arguments["storedCredentials"] as? Array<String> else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process authorization request",
                                                 details: "parameter storedCredentials is missed"))
            }
            
            
            let openID4VP = try createOpenID4VP()
            self.openID4VP = openID4VP
            
            let opts = VcparseNewOpts(true, nil)
            var parsedCredentials: Array<ApiVerifiableCredential> = Array()
            
            for cred in storedCredentials{
                let parsedVC = VcparseParse(cred, opts, nil)!
                parsedCredentials.append(parsedVC)
            }
            
            let matchedCredentials = try openID4VP.processAuthorizationRequest(authorizationRequest: authorizationRequest, credentials: parsedCredentials)
            result(matchedCredentials)
            
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
    
    public func presentCredential(result: @escaping FlutterResult) {
        do {
            guard let openID4VP = self.openID4VP else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process present credential",
                                                 details: "OpenID4VP interaction is not initialted"))
            }

            try openID4VP.presentCredential(didVerificationMethod: didVerificationMethod!)
            result(true);
            
        } catch OpenID4VPError.runtimeError(let errorMsg){
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while process authorization request",
                                     details: errorMsg))
        } catch let error as NSError{
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while process authorization request",
                                     details: error.description))
        }
    }
    
    public func createDid(didMethodType: String, result: @escaping FlutterResult){
        let didCreator = DidNewCreatorWithKeyWriter(self.kms, nil)
        do {
            let apiCreate = initializeObject(fromType: ApiCreateDIDOpts.self)
            let doc = try didCreator!.create(didMethodType, createDIDOpts: apiCreate)
            let _ = String(bytes: doc.content!, encoding: .utf8)
            didDocID = doc.id_(nil)
            didVerificationMethod = try doc.assertionMethod()
            result(didDocID)
        } catch {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while creating did",
                                     details: error))
        }
    }
    
    public func authorize(requestURI: String, result: @escaping FlutterResult){
        let clientConfig =  Openid4ciClientConfig("ClientID", crypto: self.crypto, didRes: self.didResolver, activityLogger: activityLogger)
        newOIDCInteraction = Openid4ciNewInteraction(qrCodeData.requestURI, clientConfig, nil)
        do {
            let authorizeResult  = try newOIDCInteraction?.authorize()
            let userPINRequired = authorizeResult?.userPINRequired;
            // Todo Issue-65 Pass the whole object for the future changes
            result(Bool(userPINRequired ?? false))
          } catch {
              result(FlutterError.init(code: "NATIVE_ERR",
                                       message: "error while creating new OIDC interaction",
                                       details: error))
          }
    }
    

    public func requestCredential(otp: String, result: @escaping FlutterResult){
        let clientConfig =  Openid4ciClientConfig("ClientID", crypto: self.crypto, didRes: self.didResolver, activityLogger: activityLogger)
        newOIDCInteraction = Openid4ciNewInteraction(qrCodeData.requestURI, clientConfig, nil)
        do {
            let credentialRequest = Openid4ciNewCredentialRequestOpts( otp )
            let credResp  = try newOIDCInteraction?.requestCredential(credentialRequest, vm: didVerificationMethod)
            let credentialData = credResp?.atIndex(0)!;
            result(credentialData?.serialize(nil))
          } catch let error as NSError{
              result(FlutterError.init(code: "Exception",
                                       message: "error while requesting credential",
                                       details: error.description))
          }
        
    }
    
    public func resolveCredentialDisplay(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
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

            let opts = VcparseNewOpts(true, nil)
            var parsedCredentials: ApiVerifiableCredentialsArray = ApiVerifiableCredentialsArray()!

            for cred in vcCredentials{
                let parsedVC = VcparseParse(cred, opts, nil)!
                parsedCredentials.add(parsedVC)
            }
            let resolvedDisplayData = Openid4ciResolveDisplay(parsedCredentials, issuerURI, nil, nil)
            let displayDataResp = resolvedDisplayData?.serialize(nil)
            result(displayDataResp)
          } catch let error as NSError {
                result(FlutterError.init(code: "Exception",
                                         message: "error while resolving credential",
                                         details: error.description))
            }
    }

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
            var type = activityObj?.type()
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
    
    public func storeActivityLogger(result: @escaping FlutterResult){
        var activityList: [Any] = []
        var aryLength = activityLogger!.length()
        for index in 0..<aryLength {
            activityList.append(activityLogger!.atIndex(index)!.serialize(nil))
        }

        result(activityList)
    }
    
    public func getCredID(arguments: Dictionary<String, Any>, result: @escaping FlutterResult){
        
        guard let vcCredentials = arguments["vcCredentials"] as? Array<String> else{
            return  result(FlutterError.init(code: "NATIVE_ERR",
                                             message: "error while fetching credential ID",
                                             details: "parameter storedcredentials is missed"))
        }
        let opts = VcparseNewOpts(true, nil)
        var credIDs: [Any] = []

        for cred in vcCredentials{
            let parsedVC = VcparseParse(cred, opts, nil)!
            let credID = parsedVC.id_()
            print("credid -->", credID)
            credIDs.append(credID)
            
        }
        print("first credid -->", credIDs[0])
        result(credIDs[0])
    }


    public func issuerURI( result: @escaping FlutterResult){
        let issuerURIResp = newOIDCInteraction?.issuerURI();
        result(issuerURIResp)
    }



    public func initializeObject<T: ApiCreateDIDOpts>(fromType type: T.Type) -> T {
        return T.init() //No Error
    }
    
    
    public func fetchArgsKeyValue(_ call: FlutterMethodCall, key: String) -> String? {
        guard let args = call.arguments else {
            return ""
        }
        let myArgs = args as? [String: Any];
        return myArgs?[key] as? String;
    }

    //Define type method to access the new interaction further in the flow
    class OpenID
    {
        class func NewInteraction(requestURI: String, clientConfig: Openid4ciClientConfig) -> Openid4ciInteraction?
          {
              return Openid4ciNewInteraction(requestURI, clientConfig, nil)
          }
        

    }
}

