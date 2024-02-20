/*
 Copyright Gen Digital Inc. All Rights Reserved.
 
 SPDX-License-Identifier: Apache-2.0
 */

import Foundation
import Walletsdk

public class localActivityLogger: NSObject, ApiActivityLoggerProtocol {
    public func log(_ activity: ApiActivity?) throws {
        // TODO - Add the custom activity logger implementation here
        print("")
    }
    
}
