//
//  kmsStore.swift
//  Runner
//
//  Created by Avast.Inc on 2022-10-24.
//

import UIKit
import Walletsdk
import Foundation

public class kmsStore: NSObject, LocalkmsStoreProtocol{
    
    public func put(_ keysetID: String?, key: Data?) throws {
        UserDefaults.standard.set(key, forKey: keysetID!)
        return
    }

    public func get(_ keysetID: String?) throws -> LocalkmsResult {
        let localResult = LocalkmsResult.init()
        let keyVal = UserDefaults.standard.data(forKey: keysetID!)
        if (keyVal == nil){
            localResult.found = false
            return localResult
        }
        localResult.found = true
        localResult.key = keyVal
        return localResult
    }
    
    public func delete(_ keysetID: String?) throws {
        UserDefaults.standard.removeObject(forKey: keysetID!)
        return
    }

}
