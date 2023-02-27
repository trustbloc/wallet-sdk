package dev.trustbloc.wallet

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.walleterror.Walleterror
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import walletsdk.WalletSDK
import walletsdk.flutter.converters.convertSubmissionRequirementArray
import walletsdk.flutter.converters.convertToVerifiableCredentialsArray
import walletsdk.flutter.converters.convertVerifiableCredentialsArray
import walletsdk.kmsStorage.KmsStore
import walletsdk.openid4ci.OpenID4CI
import walletsdk.openid4vp.OpenID4VP
import kotlin.collections.ArrayList


class MainActivity : FlutterActivity() {
    private var walletSDK: WalletSDK? = null
    private var openID4CI: OpenID4CI? = null
    private var openID4VP: OpenID4VP? = null


    // TODO: remove next three variables after refactoring finished.
    private var processAuthorizationRequestVCs: VerifiableCredentialsArray? = null
    private var didDocResolution: DIDDocResolution? = null

    @Override
    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
                .setMethodCallHandler { call, result ->
                    when (call.method) {
                        "initSDK" -> {
                            try {
                                initSDK()
                                result.success(null)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while creating basic sdk services", e)
                            }
                        }

                        "createDID" -> {
                            try {
                                val didCreated = createDID(call)
                                result.success(didCreated)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while creating did creator", e)
                            }
                        }
                        "authorize" -> {
                            try {
                                val userPinRequired = authorize(call)
                                result.success(userPinRequired)

                            } catch (e: Exception) {
                                val err = Walleterror.parse(e.message)
                                // Add custom error handling logic here basing on code and error properties
                                println("code: ${err.code}")
                                println("error: ${err.category}")
                                println("details: ${err.details}")

                                result.error(
                                        "Exception",
                                        "Error while authorizing the oidc vc flow",
                                        "code: ${err.code}, error: ${err.category}, details: ${err.details}"
                                )

                            }
                        }
                        "requestCredential" -> {
                            try {
                                val credentialCreated = requestCredential(call)
                                val serializedCredential = credentialCreated!!.serialize()

                                result.success(serializedCredential)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while requesting credential", e)
                            }
                        }

                        "fetchDID" -> {
                            try {
                                val didID = call.argument<String>("didID")
                            } catch (e: Exception) {
                                result.error("Exception", "Error while setting fetched DID", e)
                            }
                        }

                        "resolveCredentialDisplay" -> {
                            try {
                                val credentialDisplay = resolveCredentialDisplay(call)
                                result.success(credentialDisplay)

                            } catch (e: Exception) {
                                result.error("Exception", "Error while resolving credential display", e)
                            }
                        }

                        "processAuthorizationRequest" -> {
                            try {
                                val creds = processAuthorizationRequest(call)

                                result.success(creds)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }

                        "issuerURI" -> {
                            try {
                                val issuerURIResp = issuerURI()
                                result.success(issuerURIResp)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while getting issuerURI", e)
                            }
                        }

                        "activityLogger" -> {
                            try {
                                val activityLoggerResp = storeActivityLogger()
                                result.success(activityLoggerResp)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while storing activity logger request", e)
                            }
                        }

                        "getCredID" -> {
                            try {
                                val credID = getCredID(call)
                                result.success(credID)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }

                        "parseActivities" -> {
                            try {
                                val credID = parseActivities(call)
                                result.success(credID)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while parsing activities", e)
                            }
                        }

                        "presentCredential" -> {
                            try {
                                presentCredential(call)
                                result.success(null)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }

                        "getMatchedSubmissionRequirements" -> {
                            try {
                                result.success(getMatchedSubmissionRequirements(call))
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }


                    }
                }
    }

    private fun initSDK() {
        val walletSDK = WalletSDK()

        walletSDK.InitSDK(KmsStore(context))
        this.walletSDK = walletSDK;
    }

    /**
    Create method of Creator (dev.trustbloc.wallet.sdk.did.Creator) creates a DID document using the given DID method.
    The usage of CreateDIDOpts(dev.trustbloc.wallet.sdk.did.api) depends on the DID method you're using.
    In the app when user logins we invoke sdk Creator create method to create new did per user.
     */
    private fun createDID(call: MethodCall): String {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val didMethodType = call.argument<String>("didMethodType")
                ?: throw java.lang.Exception("didMethodType params is missed")

        val doc = walletSDK.createDID(didMethodType)

        didDocResolution = doc
        return doc.content
    }

    /**
     *Authorize method of Interaction(dev.trustbloc.wallet.sdk.openid4ci.Interaction) is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
    After initializing the Interaction object with an Issuance Request, this should be the first method you call in
    order to continue with the flow.

    AuthorizeResult is the object returned from the OpenID4CI.authorize method.
    userPINRequired method available on authorize result returns boolean value to differentiate pin is required or not.
     */
    private fun authorize(call: MethodCall): Boolean {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val requestURI = call.argument<String>("requestURI")
                ?: throw java.lang.Exception("requestURI params is missed")

        val openID4CI = walletSDK.createOpenID4CIInteraction(requestURI)

        val authRes = openID4CI.authorize()

        this.openID4CI = openID4CI

        return authRes.userPINRequired
    }

    private fun issuerURI(): String {
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.issuerURI()
    }

    /**
     * RequestCredential method of Interaction(dev.trustbloc.wallet.sdk.openid4ci.Interaction) is the final step (or second last step,
     * if the ResolveDisplay method isn't needed) in the interaction. This is called after the wallet is authorized and is ready to receive credential(s).
    Here if the pin required is true in the authorize method, then user need to enter OTP which is intercepted to create CredentialRequest Object using
    CredentialRequestOpts.If flow doesnt not require pin than Credential Request Opts will have empty string otp and sdk will return credential Data based on empty otp.
     */
    private fun requestCredential(call: MethodCall): VerifiableCredential? {
        val otp = call.argument<String>("otp") ?: throw java.lang.Exception("otp params is missed")

        val didDocResolution = this.didDocResolution
                ?: throw java.lang.Exception("DID should be created first")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.requestCredential(otp, didDocResolution.assertionMethod())
    }
    /**
     * ResolveDisplay resolves display information for issued credentials based on an issuer's metadata, which is fetched
    using the issuer's (base) URI. The CredentialDisplays returns DisplayData object correspond to the VCs passed in and are in the
    same order. This method requires one or more VCs and the issuer's base URI.
    IssuerURI and array of credentials  are parsed using VcParse to be passed to resolveDisplay which returns the resolved Display Data
     */
    private fun resolveCredentialDisplay(call: MethodCall): String? {
        val issuerURI = call.argument<String>("uri")
                ?: throw java.lang.Exception("issuerURI params is missed")
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
                ?: throw java.lang.Exception("vcCredentials params is missed")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.resolveCredentialDisplay(issuerURI, convertToVerifiableCredentialsArray(vcCredentials))
    }

    /**
     Local function  to get the credential IDs of the requested credentials
     */
    private fun getCredID(call: MethodCall): String? {
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
                ?: throw java.lang.Exception("vcCredentials params is missed")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.getCredID(vcCredentials)
    }

    /**
    This method invoke processAuthorizationRequest defined in OpenID4Vp.kt file.
     */
    private fun processAuthorizationRequest(call: MethodCall): List<String> {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")
        val authorizationRequest = call.argument<String>("authorizationRequest")
                ?: throw java.lang.Exception("authorizationRequest params is missed")
        val storedCredentials = call.argument<ArrayList<String>>("storedCredentials")

        val openID4VP = walletSDK.createOpenID4VPInteraction()

        this.openID4VP = openID4VP

        openID4VP.startVPInteraction(authorizationRequest)

        if (storedCredentials != null) {
            //TODO remove this block after refactoring finished.
            processAuthorizationRequestVCs = convertToVerifiableCredentialsArray(storedCredentials)
            val matchedReq = openID4VP.getMatchedSubmissionRequirements(convertToVerifiableCredentialsArray(storedCredentials))
            return convertVerifiableCredentialsArray(matchedReq.atIndex(0).descriptorAtIndex(0).matchedVCs)
        }
        return listOf()
    }

    private fun getMatchedSubmissionRequirements(call: MethodCall): List<Any> {
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")
        val storedCredentials = call.argument<ArrayList<String>>("storedCredentials")
                ?: throw java.lang.Exception("storedCredentials params is missed")

        return convertSubmissionRequirementArray(
                openID4VP.getMatchedSubmissionRequirements(convertToVerifiableCredentialsArray(storedCredentials)))
    }


    private fun presentCredential(call: MethodCall) {
        val selectedCredentials = call.argument<ArrayList<String>>("selectedCredentials")
        val selectedCredentialsArray = if (selectedCredentials != null) {
            convertToVerifiableCredentialsArray(selectedCredentials)
        } else {
            //TODO: remove this after refactoring will be finished
            this.processAuthorizationRequestVCs
                    ?: throw java.lang.Exception("processAuthorizationRequest should be called first.")
        }

        val didDocResolution = this.didDocResolution
                ?: throw java.lang.Exception("DID should be created first")
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")

        openID4VP.presentCredential(selectedCredentialsArray, didDocResolution.assertionMethod())
        this.openID4VP = null
    }
    /**
    Local function to fetch all activities and send the serializedData response to the app to be stored in the flutter secure storage.
     */
    private fun storeActivityLogger(): MutableList<Any> {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val activityLogger = walletSDK.activityLogger

        var activityList = mutableListOf<Any>()
        var aryLength = activityLogger?.length()
        for (i in 0..aryLength!!) {
            val serializedData = activityLogger?.atIndex(i)?.serialize()
            val activityDicResp = mutableListOf<Any>()
            if (serializedData != null) {
                activityDicResp.add(serializedData)
            }

            activityList.addAll(activityDicResp)
        }
        return activityList
    }

    /**
    ParseActivity is invoked to parse the list of activities which are stored in the app when we issue and present credential,
     */
    private fun parseActivities(call: MethodCall): MutableList<Any> {
        val arrayList = mutableListOf<Any>()

        val activities = call.argument<ArrayList<String>>("activities")
                ?: throw java.lang.Exception("parameter activities is missing")

        for (activity in activities) {
            val activityObj = Api.parseActivity(activity)
            val status = activityObj.status()
            val client = activityObj.client()
            val activityType = activityObj.type()
            val timestampDate = activityObj.unixTimestamp()

            val activityDicResp = mutableListOf<Any>()
            if (status != null) {
                activityDicResp.add(status)
            }
            if (client != null) {
                activityDicResp.add(client)
            }
            if (activityType != null) {
                activityDicResp.add(activityType)
            }
            if (timestampDate != null) {
                activityDicResp.add(timestampDate)
            }
            arrayList.addAll(activityDicResp)
        }

        return arrayList
    }


    companion object {
        private const val CHANNEL = "WalletSDKPlugin"
    }
}
