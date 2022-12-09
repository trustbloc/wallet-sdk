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
    private var signerCreator: LocalkmsSignerCreator?
    private var didDocRes: ApiDIDDocResolution?
    private var didDocID: String?
    
    
    private var openID4VP: OpenID4VP?
    
    public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
        let arguments = call.arguments as? Dictionary<String, Any>
        
        switch call.method {
        case "createDID":
            createDid(result: result)
            
        case "authorize":
            let requestURI = fetchArgsKeyValue(call, key: "requestURI")
            qrCodeData.requestURI = requestURI!
            authorize(requestURI: requestURI!, result: result)
            
        case "requestCredential":
            let otp = fetchArgsKeyValue(call, key: "otp")
            requestCredential(otp: otp!, result: result)
            
        case "initSDK":
            initSDK(result:result)
            
        case "processAuthorizationRequest":
            processAuthorizationRequest(arguments: arguments!, result: result)
            
        case "presentCredential":
            presentCredential(arguments: arguments!, result: result)
            
        default:
            print("No call method is found")
        }
    }
    
    private func initSDK(result: @escaping FlutterResult) {
        kms = LocalkmsNewKMS(nil, nil)
        didResolver = DidresolverNewDIDResolver()
        crypto = kms?.getCrypto()
        documentLoader = LinkeddomainsNewDocumentLoader()
        signerCreator = LocalkmsCreateSignerCreator(kms, nil)
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
        
        return OpenID4VP(keyReader: kms, didResolver: didResolver, documentLoader: documentLoader, crypto: crypto)
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
            
            try openID4VP.processAuthorizationRequest(authorizationRequest: authorizationRequest, storedCredentials: storedCredentials)
            
            self.openID4VP = openID4VP
            
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
    
    public func presentCredential(arguments: Dictionary<String, Any>, result: @escaping FlutterResult) {
        do {
            
            guard let signingKeyId = arguments["signingKeyId"] as? String else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process present credential",
                                                 details: "parameter signingKeyId is missed"))
            }
            
            guard let openID4VP = self.openID4VP else{
                return  result(FlutterError.init(code: "NATIVE_ERR",
                                                 message: "error while process present credential",
                                                 details: "OpenID4VP interaction is not initialted"))
            }
            
            
            
            try openID4VP.presentCredential(signingKeyId: signingKeyId)
            
            
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
    
    public func createDid(result: @escaping FlutterResult){
        let didCreator = DidcreatorNewCreatorWithKeyWriter(self.kms, nil)
        do {
            let apiCreate = initializeObject(fromType: ApiCreateDIDOpts.self)
            let doc = try didCreator!.create("key", createDIDOpts: apiCreate)
            let docString = String(bytes: doc.content!, encoding: .utf8)
            didDocID = doc.id_(nil)
            result(docString)
        } catch {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while creating did",
                                     details: nil))
        }
    }
    
    public func authorize(requestURI: String, result: @escaping FlutterResult){
        let clientConfig =  Openid4ciClientConfig( didDocID!,  clientID: "ClientID", signerCreator: self.signerCreator, didRes: self.didResolver)
        let newOIDCInteraction = Openid4ciNewInteraction(qrCodeData.requestURI, clientConfig, nil)
        do {
            let authorizeResult  = try newOIDCInteraction?.authorize()
            let userPINRequired = authorizeResult?.userPINRequired;
            // Todo Issue-65 Pass the whole object for the future changes
            result(Bool(userPINRequired ?? false))
          } catch {
              result(FlutterError.init(code: "NATIVE_ERR",
                                       message: "error while creating new OIDC interaction",
                                       details: nil))
          }
    }
    
    public func requestCredential(otp: String, result: @escaping FlutterResult){
        let clientConfig =  Openid4ciClientConfig( didDocID!,  clientID: "ClientID", signerCreator: self.signerCreator, didRes: self.didResolver)
        let newOIDCInteraction = Openid4ciNewInteraction(qrCodeData.requestURI, clientConfig, nil)
        do {
            let credentialRequest = Openid4ciCredentialRequestOpts()
            credentialRequest.userPIN = otp
            let credResp  = try newOIDCInteraction?.requestCredential(credentialRequest)
            // TODO Checking the first credential in the array
            if (credResp!.length() > 0) {
                let resolvedDisplayData = try newOIDCInteraction?.resolveDisplay()
                let displayDataResp = String(bytes: (resolvedDisplayData?.data)!, encoding: .utf8)
                result(displayDataResp)
            }
          } catch {
              result(FlutterError.init(code: "NATIVE_ERR",
                                       message: "error while requesting credential",
                                       details: error))
          }
        
    }
    
    public func initializeObject<T: ApiCreateDIDOpts>(fromType type: T.Type) -> T {
        return T.init() //No Error
    }
    
    public func initializeCredentialRequest<T: Openid4ciCredentialRequestOpts>(fromType type: T.Type) -> T {
        return T.init() //No Error
    }
    
    public func initializeApiJSONObject<T: ApiJSONObject>(fromType type: T.Type) -> T {
        return T.init() //No Error
    }
    
    
    public func fetchArgsKeyValue(_ call: FlutterMethodCall, key: String) -> String? {
        guard let args = call.arguments else {
            return ""
        }
        let myArgs = args as? [String: Any];
        return myArgs?[key] as? String;
    }

    public func dataToJSON(data: Data) -> Any? {
       do {
           return try JSONSerialization.jsonObject(with: data, options: .mutableContainers)
       } catch let myJSONError {
           print(myJSONError)
       }
       return nil
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

