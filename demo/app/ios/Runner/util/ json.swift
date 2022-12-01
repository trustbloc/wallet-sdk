//
//   json.swift
//  Runner
//
//  Created by Volodymyr Kubiv on 29.11.2022.
//

import Foundation
import Walletsdk

func createJsonArray(objs: Array<String>) -> ApiJSONArray {
    var result = "["
    for i in 0...objs.count {
        if (i != 0) {
            result += ","
        }
        result += objs[i]
    }
    result += "]"
    
    let arr = ApiJSONArray()
    arr.data = result.data(using: .utf8)
    return arr
}
