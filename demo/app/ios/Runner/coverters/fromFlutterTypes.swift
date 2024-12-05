/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Foundation
import Walletsdk

/* Need to be removed */
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

func convertToVerifiableCredentialsArrayV2(credentials: Array<String>, configIds: Array<String>) -> VerifiableCredentialsArrayV2 {
    let opts = VerifiableNewOpts()
    opts!.disableProofCheck()
    
    let parsedCredentials = VerifiableCredentialsArrayV2()!
        
    for i in 0..<credentials.count {
        let parsedVC = VerifiableParseCredential(credentials[i], opts, nil)!
        parsedCredentials.add(parsedVC, configID: configIds[0])
    }
    
    return parsedCredentials
}
