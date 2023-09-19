/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.flutter.converters

import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray
import dev.trustbloc.wallet.sdk.verifiable.Opts
import dev.trustbloc.wallet.sdk.verifiable.Verifiable

fun convertToVerifiableCredentialsArray(credentials: List<String>): CredentialsArray {
    val credArray = CredentialsArray()
    for (cred in credentials) {
        val opts = Opts()
        opts.disableProofCheck()

        val parsedCred = Verifiable.parseCredential(cred, opts)
        credArray.add(parsedCred)
    }

    return credArray
}