package dev.trustbloc.wallet

import android.annotation.SuppressLint
import dev.trustbloc.wallet.sdk.api.Api
import dev.trustbloc.wallet.sdk.api.DIDDocResolution
import dev.trustbloc.wallet.sdk.api.StringArray
import dev.trustbloc.wallet.sdk.credential.StatusVerifier
import dev.trustbloc.wallet.sdk.did.Did
import dev.trustbloc.wallet.sdk.display.Display
import dev.trustbloc.wallet.sdk.oauth2.Oauth2
import dev.trustbloc.wallet.sdk.verifiable.Credential
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray
import dev.trustbloc.wallet.sdk.verifiable.Opts
import dev.trustbloc.wallet.sdk.verifiable.Verifiable
import dev.trustbloc.wallet.sdk.version.Version
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
import java.text.SimpleDateFormat


class MainActivity : FlutterActivity() {
    private var walletSDK: WalletSDK? = null
    private var openID4CI: OpenID4CI? = null
    private var openID4VP: OpenID4VP? = null


    // TODO: remove next three variables after refactoring finished.
    private var processAuthorizationRequestVCs: CredentialsArray? = null
    private var didDocResolution: DIDDocResolution? = null

    @Override
    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
                .setMethodCallHandler { call, result ->
                    when (call.method) {
                        "initSDK" -> {
                            try {
                                initSDK(call)
                                result.success(null)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while creating basic sdk services", e)
                            }
                        }
                        "getVersionDetails" -> {
                            try {
                                val walletSDKVersion = getVersionDetails()
                                result.success(walletSDKVersion)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while get wallet sdk version", e)
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
                        "initialize" -> {
                            try {
                                val userPinRequired = initialize(call)
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
                        "credentialStatusVerifier" -> {
                            try {
                                val credentialStatus = credentialStatusVerifier(call)
                                result.success(credentialStatus)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while credential status verifier", e.localizedMessage)
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

                        "requestCredentialWithAuth" -> {
                            try {
                                val credentialCreated =  requestCredentialAuth(call)
                                val serializedCredential = credentialCreated!!.serialize()

                                result.success(serializedCredential)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while requesting credential auth", e)
                            }
                        }

                        "fetchDID" -> {
                            try {
                                val didID = call.argument<String>("didID")
                            } catch (e: Exception) {
                                result.error("Exception", "Error while setting fetched DID", e)
                            }
                        }

                        "serializeDisplayData" -> {
                            try {
                                val credentialDisplay = serializeDisplayData(call)
                                result.success(credentialDisplay)

                            } catch (e: Exception) {
                                result.error("Exception", "Error while resolving credential display", e)
                            }
                        }

                        "resolveCredentialDisplay" -> {
                            try{
                                val credentialDisplay = resolveCredentialDisplay(call)
                                result.success(credentialDisplay)
                            } catch (e: Exception)  {
                                result.error("Exception", "Error while resolving credential display 2", e)
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

                        "getIssuerID" -> {
                            try {
                                val issuerID = getIssuerID(call)
                                result.success(issuerID)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while getting issuer ID", e)
                            }
                        }

                        "wellKnownDidConfig" -> {
                            try {
                                val didValidateResultResp = wellKnownDidConfig(call)
                                result.success(didValidateResultResp)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while getting well known did config", e)
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

                       "getVerifierDisplayData" -> {
                           try {
                               val verifierDisplayData = getVerifierDisplayData()
                               result.success(verifierDisplayData)
                           } catch (e: Exception) {
                               result.error("Exception", "Error while getting verifier display data", e)
                           }
                       }

                    }
                }
    }

    private fun initSDK(call: MethodCall) {
        val walletSDK = WalletSDK()
        val didResolverURI = call.argument<String>("didResolverURI")
            ?: throw java.lang.Exception("didResolverURI params is missed")
        walletSDK.InitSDK(KmsStore(context), didResolverURI)
        this.walletSDK = walletSDK;
    }

    private fun getVersionDetails(): MutableMap<String, Any> {
        var versionResp: MutableMap<String, Any> = mutableMapOf()
        versionResp["walletSDKVersion"] = Version.getVersion()
        versionResp["gitRevision"] = Version.getGitRevision()
        versionResp["buildTimeRev"] = Version.getBuildTime()
        return  versionResp
    }

    /**
    Create method of Creator (dev.trustbloc.wallet.sdk.did.Creator) creates a DID document using the given DID method.
    The usage of CreateDIDOpts(dev.trustbloc.wallet.sdk.did.api) depends on the DID method you're using.
    In the app when user logins we invoke sdk Creator create method to create new did per user.
     */
    private fun createDID(call: MethodCall): MutableMap<String, Any> {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val didMethodType = call.argument<String>("didMethodType")
                ?: throw java.lang.Exception("didMethodType params is missed")

        val didKeyType = call.argument<String>("didKeyType")
            ?: throw java.lang.Exception("didKeyType params is missed")

        val doc = walletSDK.createDID(didMethodType, didKeyType)
        didDocResolution = doc
        val docResolution: MutableMap<String, Any> = mutableMapOf()
        docResolution["did"] = doc.id()
        docResolution["didDoc"] = doc.content
        return docResolution
    }

    /**
     *Initialize method of Interaction(dev.trustbloc.wallet.sdk.openid4ci.Interaction) is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
    After initializing the Interaction object with an Issuance Request, this should be the first method you call in
    order to continue with the flow.
     */
    private fun initialize(call: MethodCall): MutableMap<String, Any> {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val requestURI = call.argument<String>("requestURI")
                ?: throw java.lang.Exception("requestURI params is missed")

        val openID4CI = walletSDK.createOpenID4CIInteraction(requestURI)

        this.openID4CI = openID4CI

        var pinRequired = false
        var authorizationLink = ""

        var flowType = openID4CI.checkFlow()
            if (flowType == "preauth-code-flow") {
                pinRequired = openID4CI.pinRequired()
            }
            if (flowType == "auth-code-flow"){
                val authCodeArgs = call.argument<MutableMap<String, String>>("authCodeArgs")
                    ?: throw java.lang.Exception("authCodeArgs params is missed, Pass scopes, clientID and redirectURI as the arguments")

                var dynamicRegistrationSupported =  openID4CI.dynamicRegistrationSupported()
                var clientID = authCodeArgs["clientID"].toString()
                val redirectURI = authCodeArgs["redirectURI"].toString()
                var scopes = authCodeArgs["scopes"] as ArrayList<String>

                if (dynamicRegistrationSupported)  {
                    var dynamicRegistrationEndpoint = openID4CI.dynamicRegistrationEndpoint()

                    var clientMetadata = Oauth2.newClientMetadata()
                    var grantTypesArr = StringArray()
                    grantTypesArr.append("authorization_code")
                    clientMetadata.setGrantTypes(grantTypesArr)

                    var redirectUri = StringArray()
                    redirectUri.append(redirectURI)
                    clientMetadata.setRedirectURIs(redirectUri)

                    var spaceSeparatedScopes = scopes.joinToString(" ")
                    clientMetadata.setScope(spaceSeparatedScopes)

                    clientMetadata.setTokenEndpointAuthMethod("none")

                    var authorizationCodeGrantParams = openID4CI.getAuthorizationCodeGrantParams()
                    if (authorizationCodeGrantParams.hasIssuerState()){
                         var issuerState = authorizationCodeGrantParams.issuerState()
                         clientMetadata.setIssuerState(issuerState)
                    }

                    var registrationResp = Oauth2.registerClient(dynamicRegistrationEndpoint, clientMetadata, null)
                    clientID = registrationResp.clientID()
                }

                if (!authCodeArgs.keys.contains("scopes")) {
                    authorizationLink = openID4CI.createAuthorizationURL(
                        clientID,
                        redirectURI)
                } else {
                    authorizationLink = openID4CI.createAuthorizationURLWithScopes(scopes,
                        clientID,
                        redirectURI)
                }
            }

        val flowTypeData: MutableMap<String, Any> = mutableMapOf()
        flowTypeData["pinRequired"] = pinRequired
        flowTypeData["authorizationURLLink"] = authorizationLink

        return flowTypeData
    }

    private fun issuerURI(): String {
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.issuerURI()
    }

    /**
     * RequestCredential method of Interaction(dev.trustbloc.wallet.sdk.openid4ci.Interaction) is the final step,
     * in the interaction. This is called after the wallet is authorized and is ready to receive credential(s).
    Here if the pin required is true in the authorize method, then user need to enter OTP which is intercepted to create CredentialRequest Object using
    CredentialRequestOpts.If flow doesnt not require pin than Credential Request Opts will have empty string otp and sdk will return credential Data based on empty otp.
     */
    private fun requestCredential(call: MethodCall): Credential? {
        val otp = call.argument<String>("otp") ?: throw java.lang.Exception("otp params is missed")

        val didDocResolution = this.didDocResolution
                ?: throw java.lang.Exception("DID should be created first")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.requestCredential(didDocResolution.assertionMethod(), otp)
    }


    private fun requestCredentialAuth(call: MethodCall): Credential? {
        val redirectURI = call.argument<String>("redirectURIWithParams") ?: throw java.lang.Exception("redirectURIWithParams is missed")

        val didDocResolution = this.didDocResolution
            ?: throw java.lang.Exception("DID should be created first")

        val openID4CI = this.openID4CI
            ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.requestCredentialWithAuth(didDocResolution.assertionMethod(), redirectURI)
    }



    /**
     * ResolveDisplay resolves display information for issued credentials based on an issuer's metadata, which is fetched
    using the issuer's (base) URI. The CredentialDisplays returns DisplayData object correspond to the VCs passed in and are in the
    same order. This method requires one or more VCs and the issuer's base URI.
    IssuerURI and array of credentials  are parsed using VcParse to be passed to resolveDisplay which returns the resolved Display Data
     */
    private fun serializeDisplayData(call: MethodCall): String? {
        val issuerURI = call.argument<String>("uri")
                ?: throw java.lang.Exception("issuerURI params is missed")
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
                ?: throw java.lang.Exception("vcCredentials params is missed")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.serializeDisplayData(issuerURI, convertToVerifiableCredentialsArray(vcCredentials))
    }

    private fun resolveCredentialDisplay(call: MethodCall): MutableList<Any> {
        val resolvedCredentialDisplayData = call.argument<String>("resolvedCredentialDisplayData")
            ?: throw java.lang.Exception("resolvedCredentialDisplayData params is missed")

        val displayData = Display.parseData(resolvedCredentialDisplayData)
        val issuerDisplayData = displayData.issuerDisplay()
        val resolvedCredDisplayList = mutableListOf<Any>()
        val claimList = mutableListOf<Any>()

        for (i in 0 until (displayData.credentialDisplaysLength())) {
            val credentialDisplay = displayData.credentialDisplayAtIndex(i)
            for (i in 0 until credentialDisplay.claimsLength()){
                val claim = credentialDisplay.claimAtIndex(i)
                val claims: MutableMap<String, Any> = mutableMapOf()
                if (claim.isMasked){
                    claims["value"] = claim.value()
                    claims["rawValue"] = claim.rawValue()
                }
                if (claim.hasOrder()) {
                    val  order = claim.order()
                    claims["order"] = order
                }
                claims["rawValue"] = claim.rawValue()
                claims["valueType"] = claim.valueType()
                claims["label"] = claim.label()
                claimList.addAll(listOf(claims))
            }
            var overview = credentialDisplay.overview()
            var resolveDisplayResp : MutableMap<String, Any> = mutableMapOf()
            resolveDisplayResp["claims"] = claimList
            resolveDisplayResp["overviewName"] = overview.name()
            resolveDisplayResp["logo"] = overview.logo().url()
            resolveDisplayResp["textColor"] = overview.textColor()
            resolveDisplayResp["backgroundColor"] = overview.backgroundColor()
            resolveDisplayResp["issuerName"] = issuerDisplayData.name()

            resolvedCredDisplayList.addAll(listOf(resolveDisplayResp))
        }
        return resolvedCredDisplayList
    }

    /**
     Local function  to get the credential IDs of the requested credentials
     */
    private fun getCredID(call: MethodCall): String {
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
                ?: throw java.lang.Exception("vcCredentials params is missed")

        val opts = Opts()
        opts.disableProofCheck()

        val credIds = ArrayList<String>()
        for (cred in vcCredentials) {
            val parsedVC = Verifiable.parseCredential(cred, opts)
            var credID = parsedVC.id()
            credIds.add(credID)
        }
        return credIds[0]
    }

    private fun getIssuerID(call: MethodCall): String {
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
            ?: throw java.lang.Exception("vcCredentials params is missed")

        val opts = Opts()
        opts.disableProofCheck()

        for (cred in vcCredentials) {
            val parsedVC = Verifiable.parseCredential(cred, opts)
            var issuerID = parsedVC.issuerID()
            return issuerID
        }
        return  ""
    }

    private fun wellKnownDidConfig(call: MethodCall): MutableMap<String, Any> {
        val issuerID = call.argument<String>("issuerID")
            ?: throw java.lang.Exception("issuer id is missing")

        val walletSDK = this.walletSDK
            ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val validationResult = try {
            Did.validateLinkedDomains(issuerID, walletSDK.didResolver, null)
        } catch (e: Exception) {
            println("error received while getting well known didConfig $e")
            val didValidateResultResp: MutableMap<String, Any> = mutableMapOf()
            didValidateResultResp["isValid"] = false
            didValidateResultResp["serviceURL"] = ""
            return didValidateResultResp
        }

        val didValidateResultResp: MutableMap<String, Any> = mutableMapOf()
        didValidateResultResp["isValid"] = validationResult.isValid
        didValidateResultResp["serviceURL"] = validationResult.serviceURL
        println(didValidateResultResp)

       return didValidateResultResp
    }

    private fun getVerifierDisplayData() : MutableMap<String, Any> {
        val openID4VP = this.openID4VP
            ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")

        val verifierDisplayDataResp = openID4VP.getVerifierDisplayData()
        val verifierDisplayData: MutableMap<String, Any> = mutableMapOf()
        verifierDisplayData["did"] = verifierDisplayDataResp.did()
        verifierDisplayData["logoUri"] = verifierDisplayDataResp.logoURI()
        verifierDisplayData["name"] = verifierDisplayDataResp.name()
        verifierDisplayData["purpose"] = verifierDisplayDataResp.purpose()
        return verifierDisplayData
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

    private fun credentialStatusVerifier(call: MethodCall): Boolean {
        val credentials = call.argument<List<String>>("credentials")
            ?: throw java.lang.Exception("credentials params is missed")

            val statusVerifier = StatusVerifier(null)
            val credentialArray = convertToVerifiableCredentialsArray(credentials)
        return try {
            statusVerifier.verify(credentialArray.atIndex(0))
            true
        } catch (e: Exception) {
            throw e
        }
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

        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")

        openID4VP.presentCredential(selectedCredentialsArray)
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
    private fun parseActivities(call: MethodCall): ArrayList<HashMap<String, String>> {
        val parseActivityList = HashMap<String, String>()

        val activities = call.argument<ArrayList<String>>("activities")
                ?: throw java.lang.Exception("parameter activities is missing")

        for (activity in activities) {
            val activityObj = Api.parseActivity(activity)
            val status = activityObj.status()
            val client = activityObj.client()
            val activityType = activityObj.type()
            val operation = activityObj.operation()
            val timestampDate = activityObj.unixTimestamp()

            val simpleDate = SimpleDateFormat("MMM-dd-yyyy hh:mm:ss")
            val date = simpleDate.format(timestampDate)

            val activityDicResp = HashMap<String, String>()
            activityDicResp.put("Status",status)
            activityDicResp.put("Issued By", client)
            activityDicResp.put("Operation",operation);
            activityDicResp.put("Activity Type",activityType);
            activityDicResp.put("Date",date);

            parseActivityList.putAll(activityDicResp)
          }

        return arrayListOf(parseActivityList)
    }


    companion object {
        private const val CHANNEL = "WalletSDKPlugin"
    }
}
