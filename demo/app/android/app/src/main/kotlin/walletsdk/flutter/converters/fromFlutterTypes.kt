package walletsdk.flutter.converters

import dev.trustbloc.wallet.sdk.api.VerifiableCredentialsArray
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirementArray
import dev.trustbloc.wallet.sdk.vcparse.Opts
import dev.trustbloc.wallet.sdk.vcparse.Vcparse
import java.lang.Exception

fun convertToVerifiableCredentialsArray(credentials: List<String>): VerifiableCredentialsArray {
    val credArray = VerifiableCredentialsArray()
    for (cred in credentials) {
        val opts = Opts(true, null)

        val parsedCred = Vcparse.parse(cred, opts)
        credArray.add(parsedCred)
    }

    return credArray
}