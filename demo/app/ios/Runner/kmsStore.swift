//
//  kmsStore.swift
//  Runner
//
//  Created by Talwinder Kaur on 2023-01-26.
//

import UIKit
import Walletsdk
import Foundation

enum KmsStoreEroor: Error {
    case runtimeError(String)
}

public class kmsStore: NSObject, LocalkmsStoreProtocol{
    public func delete(_ keysetID: String?) throws {
        UserDefaults.standard.removeObject(forKey: keysetID!)
        return
    }
    
    public func get(_ keysetID: String?) throws -> Data {
        print("inside get keysetid getting called", keysetID)
        let key = UserDefaults.standard.data(forKey: keysetID!)
        if let data = key {
        return data
        } else {
            throw  KmsStoreEroor.runtimeError("key doesnt exit")
        }
    }
    
    public func put(_ keysetID: String?, key: Data?) throws {
        print("inside put is keysetid ", keysetID)
        UserDefaults.standard.set(key, forKey: keysetID!)
        return
    }

}
