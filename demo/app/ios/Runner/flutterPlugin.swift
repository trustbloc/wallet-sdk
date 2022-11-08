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
        let localKMS = LocalkmsNewKMS(nil)
        let didCreator = DidcreatorNewCreatorWithKeyWriter(localKMS, nil)
        do {
            let apiCreate = initializeObject(fromType: ApiCreateDIDOpts.self)
            let doc = try didCreator!.create("key", createDIDOpts: apiCreate)
            let docString = String(bytes: doc, encoding: .utf8)
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
