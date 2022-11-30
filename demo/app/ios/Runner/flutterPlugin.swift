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
    
    public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
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

            
        default:
           print("No call method is found")
        }
    }
    
    public func createDid(result: @escaping FlutterResult){
        let localKMS = LocalkmsNewKMS(nil, nil)
        let didCreator = DidcreatorNewCreatorWithKeyWriter(localKMS, nil)
        do {
            let apiCreate = initializeObject(fromType: ApiCreateDIDOpts.self)
            let doc = try didCreator!.create("key", createDIDOpts: apiCreate)
            let docString = String(bytes: doc, encoding: .utf8)
            result(docString)
        } catch {
            result(FlutterError.init(code: "NATIVE_ERR",
                                     message: "error while creating did",
                                     details: nil))
        }
    }
    
    public func authorize(requestURI: String, result: @escaping FlutterResult){
        let newOIDCInteraction = OpenID.NewInteraction(requestURI: requestURI)
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
        let newOIDCInteraction = OpenID.NewInteraction(requestURI:  qrCodeData.requestURI)
        do {
            let credentialRequest = initializeCredentialRequest(fromType: Openid4ciCredentialRequestOpts.self)
            credentialRequest.userPIN = otp
            let credResp  = try newOIDCInteraction?.requestCredential(credentialRequest)
            if (credResp != nil ) {
                let resolvedDisplayData = try newOIDCInteraction?.resolveDisplay()
                let rsdResp = initializeApiJSONObject(fromType: ApiJSONObject.self)
                rsdResp.data = resolvedDisplayData?.data
                let displayDataResp = String(bytes: rsdResp.data!, encoding: .utf8)
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
          class func NewInteraction(requestURI: String) -> Openid4ciInteraction?
          {
              return Openid4ciNewInteraction(requestURI, nil)
          }
        
         
    }

}

