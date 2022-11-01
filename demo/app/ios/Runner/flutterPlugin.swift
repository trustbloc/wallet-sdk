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
 
  public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
      if (call.method == "createDID"){
          createDid(result: result)
      }

  }
    public func createDid(result: @escaping FlutterResult) {
        let didCreator = DidNewCreator()
        do {
            let apiCreate = initializeObject(fromType: ApiCreateDIDOpts.self)
            apiCreate.method = "key"
            let byteArray: [UInt8] = [19, 14, 81, 157, 226, 93, 94, 94, 61, 151, 246, 226, 20, 86, 41, 139, 164, 203, 99, 96 ,155 ,5, 156, 244, 198, 141, 99, 6, 19, 130, 72, 46]
            apiCreate.key = Data(byteArray)
            print("initializing api create did opts")
            
            let doc = try didCreator!.create(apiCreate)
            let docString = String(bytes: doc, encoding: .utf8)
            print("did create" , docString!)
        } catch {
            result(FlutterError.init(code: "NATIVE_ERR",
            message: "error while creating did",
            details: nil))
        }
    
    }
    
   public func initializeObject<T: ApiCreateDIDOpts>(fromType type: T.Type) -> T {
        return T.init() //No Error
    }
}
