//
//  flutterPlugin.swift
//  Runner
//
//  Created by Talwinder Kaur on 2022-10-24.
//

import Flutter
import UIKit
import Walletsdk

public class SwiftHelloPlugin: NSObject, FlutterPlugin {
  public static func register(with registrar: FlutterPluginRegistrar) {
    let channel = FlutterMethodChannel(name: "HelloPlugin", binaryMessenger: registrar.messenger())
    let instance = SwiftHelloPlugin()
    registrar.addMethodCallDelegate(instance, channel: channel)
  }
 
  public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
      if (call.method == "storeCredentials"){
          
          let storageProvider = StorageNewProvider()
          do {
              let jsonVC = """
              {
                "id": "http://example.edu/credentials/1872"
              }
              """.data(using: .utf8)!
              try storageProvider!.add(jsonVC)
              print("credential is stored")
          } catch {
              result("error while storing verfiable credential")
          }
      }
      
      if (call.method == "sayHello"){
          guard let args = call.arguments as? [String : Any] else {return}
          let name = args["name"] as! String
          result("Welcome " + name)
      }

  }
}
