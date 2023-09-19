/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

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
