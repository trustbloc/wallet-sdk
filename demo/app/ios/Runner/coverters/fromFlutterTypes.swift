//
//  fromFlutterTypes.swift
//  Runner
//
//  Created by Volodymyr Kubiv on 28.02.2023.
//

import Foundation
import Walletsdk

func convertToVerifiableCredentialsArray(credentials: Array<String>) -> ApiVerifiableCredentialsArray {
    let opts = VcparseNewOpts()
    opts!.disableProofCheck()
    
    let parsedCredentials = ApiVerifiableCredentialsArray()!
    
    for cred in credentials{
        let parsedVC = VcparseParse(cred, opts, nil)!
        parsedCredentials.add(parsedVC)
    }

    return parsedCredentials
}
