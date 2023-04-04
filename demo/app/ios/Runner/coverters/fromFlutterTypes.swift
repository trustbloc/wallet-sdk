//
//  fromFlutterTypes.swift
//  Runner
//
//  Created by Volodymyr Kubiv on 28.02.2023.
//

import Foundation
import Walletsdk

func convertToVerifiableCredentialsArray(credentials: Array<String>) -> VerifiableCredentialsArray {
    let opts = VerifiableNewOpts()
    opts!.disableProofCheck()
    
    let parsedCredentials = VerifiableCredentialsArray()!
    
    for cred in credentials{
        let parsedVC = VerifiableParseCredential(cred, opts, nil)!
        parsedCredentials.add(parsedVC)
    }

    return parsedCredentials
}
