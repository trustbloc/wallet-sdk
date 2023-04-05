package walletsdk.flutter.converters

import dev.trustbloc.wallet.sdk.api.VerifiableCredentialsArray
import dev.trustbloc.wallet.sdk.vcparse.*
import java.lang.Exception

fun convertToVerifiableCredentialsArray(credentials: List<String>): VerifiableCredentialsArray {
    val credArray = VerifiableCredentialsArray()
    for (cred in credentials) {
        val opts = Opts()
        opts.disableProofCheck()

        val parsedCred = Vcparse.parse(cred, opts)
        credArray.add(parsedCred)
    }

    return credArray
}