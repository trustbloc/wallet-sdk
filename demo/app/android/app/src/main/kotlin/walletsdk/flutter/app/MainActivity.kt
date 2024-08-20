/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dev.trustbloc.wallet

import dev.trustbloc.wallet.sdk.api.Api
import dev.trustbloc.wallet.sdk.api.DIDDocResolution
import dev.trustbloc.wallet.sdk.api.StringArray
import dev.trustbloc.wallet.sdk.credential.StatusVerifier
import dev.trustbloc.wallet.sdk.did.Did
import dev.trustbloc.wallet.sdk.display.Display
import dev.trustbloc.wallet.sdk.oauth2.Oauth2
import dev.trustbloc.wallet.sdk.openid4ci.SupportedCredentials
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
import walletsdk.flutter.converters.*
import walletsdk.kmsStorage.KmsStore
import walletsdk.openid4ci.OpenID4CI
import walletsdk.openid4ci.WalletInitiatedOpenID4CI
import walletsdk.openid4vp.OpenID4VP
import java.text.SimpleDateFormat


class MainActivity : FlutterActivity() {
    private var walletSDK: WalletSDK? = null
    private var openID4CI: OpenID4CI? = null
    private var openID4VP: OpenID4VP? = null
    private var walletInitiatedOpenID4CI: WalletInitiatedOpenID4CI? = null

    // TODO: remove next three variables after refactoring finished.
    private var processAuthorizationRequestVCs: CredentialsArray? = null
    private var didDocResolution: DIDDocResolution? = null
    private var attestationDID: DIDDocResolution? = null

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
                                result.error(
                                        "Exception",
                                        "Error while initializing the interaction",
                                        "code: ${err.code}, error: ${err.category}, details: ${err.details}"
                                )

                            }
                        }

                        "initializeWalletInitiatedFlow" -> {
                            try {
                                val supportedCredentials = initializeWalletInitiatedFlow(call)
                                result.success(supportedCredentials)
                            } catch (e: Exception) {
                                val err = Walleterror.parse(e.message)
                                result.error(
                                        "Exception",
                                        "Error while initializing wallet initiated flow",
                                        "code: ${err.code}, error: ${err.category}, details: ${err.details}"
                                )

                            }
                        }

                        "createAuthorizationURLWalletInitiatedFlow" -> {
                            try {
                                val authorizationLink = createAuthorizationURLWalletInitiatedFlow(call)
                                result.success(authorizationLink)
                            } catch (e: Exception) {
                                val err = Walleterror.parse(e.message)
                                result.error(
                                        "Exception",
                                        "Error while creating authorization link",
                                        "code: ${err.code}, error: ${err.category}, details: ${err.details}"
                                )

                            }
                        }

                        "credentialStatusVerifier" -> {
                            try {
                                val credentialStatus = credentialStatusVerifier(call)
                                result.success(credentialStatus)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while credential status verifier",
                                        e.localizedMessage
                                )
                            }
                        }

                        "evaluateIssuanceTrustInfo" -> {
                            try {
                                result.success(evaluateIssuanceTrustInfo(call))
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while checkWithTrustRegistry",
                                        e.localizedMessage
                                )
                            }
                        }

                        "evaluatePresentationTrustInfo" -> {
                            try {
                                result.success(evaluatePresentationTrustInfo(call))
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while checkWithTrustRegistry",
                                        e.localizedMessage
                                )
                            }
                        }

                        "noConsentAcknowledgement" -> {
                            try {
                                result.success(noConsentAcknowledgement())
                            } catch (e: Exception) {
                                result.error(
                                    "Exception",
                                    "Error while calling no consent acknowledgement\"",
                                    e.localizedMessage
                                )
                            }
                        }

                        "requestCredentials" -> {
                            try {
                                val credentialsCreated = requestCredentials(call)
                                result.success(credentialsCreated)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while requesting credential",
                                        e.localizedMessage
                                )
                            }
                        }

                        "requestCredentialWithAuth" -> {
                            try {
                                val credentialCreated = requestCredentialAuth(call)
                                result.success(credentialCreated)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while requesting credential auth",
                                        e.localizedMessage
                                )
                            }
                        }

                        "requestCredentialWithWalletInitiatedFlow" -> {
                            try {
                                val credentialCreated = requestCredentialWithWalletInitiatedFlow(call)
                                result.success(credentialCreated)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while requesting credential in wallet initiated flow",
                                        e.localizedMessage
                                )
                            }
                        }

                        "resolveDisplayData" -> {
                            try {
                                val credentialsDisplay = resolveDisplayData(call)
                                result.success(credentialsDisplay)

                            } catch (e: Exception) {
                                result.error("Exception", "Error while serializing display data", e)
                            }
                        }

                        "parseCredentialDisplay" -> {
                            try {
                                val credentialDisplay = parseCredentialDisplay(call)
                                result.success(credentialDisplay)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while resolving credential display", e)
                            }
                        }

                        "parseIssuerDisplay" -> {
                            try {
                                val issuerDisplay = parseIssuerDisplay(call)
                                result.success(issuerDisplay)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while resolving issuer display", e)
                            }
                        }

                        "processAuthorizationRequest" -> {
                            try {
                                val creds = processAuthorizationRequest(call, result)

                                result.success(creds)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while processing authorization request",
                                        e
                                )
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

                        "getCredentialOfferDisplayData" -> {
                            try {
                                result.success(getCredentialOfferDisplayData())
                            } catch (e: Exception) {
                                result.error("Exception", "Error while getting issuer Metadata", e)
                            }
                        }

                        "parseWalletSDKError" -> {
                            try {
                                val parsedWalletError = parseWalletSDKError(call)
                                result.success(parsedWalletError)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while parsing wallet sdk error", e)
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
                                result.error(
                                        "Exception",
                                        "Error while getting well known did config",
                                        e
                                )
                            }
                        }

                        "activityLogger" -> {
                            try {
                                val activityLoggerResp = storeActivityLogger()
                                result.success(activityLoggerResp)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while storing activity logger request",
                                        e
                                )
                            }
                        }

                        "getCredID" -> {
                            try {
                                val credID = getCredID(call)
                                result.success(credID)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while processing authorization request",
                                        e
                                )
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
                                result.error(
                                        "Exception",
                                        "Error while processing authorization request",
                                        e
                                )
                            }
                        }

                        "getCustomScope" -> {
                            try {
                                val customScopeList = getCustomScope()
                                result.success(customScopeList)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while getting custom scope",
                                        e
                                )
                            }
                        }


                        "getMatchedSubmissionRequirements" -> {
                            try {
                                result.success(getMatchedSubmissionRequirements(call))
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while processing authorization request",
                                        e
                                )
                            }
                        }

                        "getVerifierDisplayData" -> {
                            try {
                                val verifierDisplayData = getVerifierDisplayData()
                                result.success(verifierDisplayData)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while getting verifier display data",
                                        e
                                )
                            }
                        }

                        "requireAcknowledgment" -> {
                            try {
                                val ackResp = requireAcknowledgment()
                                result.success(ackResp)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "error in require acknowlwdgement",
                                        e
                                )
                            }
                        }

                        "acknowledgeSuccess" -> {
                            try {
                                acknowledgeSuccess()
                                result.success(null)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "error in acknowledge Success",
                                        e
                                )
                            }
                        }

                        "acknowledgeReject" -> {
                            try {
                                acknowledgeReject()
                                result.success(null)
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "error in acknowledge reject",
                                        e
                                )
                            }
                        }

                        "getAttestationVC" -> {
                            try {
                                val cred = getAttestationVC(call)
                                result.success(cred)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while getting attestation vc", e)
                            }

                        }

                    }
                }
    }

    private fun initSDK(call: MethodCall) {
        val walletSDK = WalletSDK()
        val didResolverURI = call.argument<String>("didResolverURI")
                ?: throw java.lang.Exception("didResolverURI params is missed")

        walletSDK.initSDK(KmsStore(context), didResolverURI)
        this.walletSDK = walletSDK;
    }

    private fun getVersionDetails(): MutableMap<String, Any> {
        var versionResp: MutableMap<String, Any> = mutableMapOf()
        versionResp["walletSDKVersion"] = Version.getVersion()
        versionResp["gitRevision"] = Version.getGitRevision()
        versionResp["buildTimeRev"] = Version.getBuildTime()
        return versionResp
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
        if (flowType == "auth-code-flow") {
            val authCodeArgs = call.argument<MutableMap<String, String>>("authCodeArgs")
                    ?: throw java.lang.Exception("authCodeArgs params is missed, Pass scopes, clientID and redirectURI as the arguments")

            var dynamicRegistrationSupported = openID4CI.dynamicRegistrationSupported()
            var clientID = authCodeArgs["clientID"].toString()
            val redirectURI = authCodeArgs["redirectURI"].toString()
            val oauthDiscoverableClientURI = authCodeArgs["oauthDiscoverableClientURI"]
            var scopesFromArgs = authCodeArgs["scopes"] as ArrayList<String>

            var scopes = StringArray()
            for (scope in scopesFromArgs) scopes.append(scope)

            if (dynamicRegistrationSupported) {
                var dynamicRegistrationEndpoint = openID4CI.dynamicRegistrationEndpoint()

                var clientMetadata = Oauth2.newClientMetadata()
                var grantTypesArr = StringArray()
                grantTypesArr.append("authorization_code")
                clientMetadata.setGrantTypes(grantTypesArr)

                var redirectUri = StringArray()
                redirectUri.append(redirectURI)
                clientMetadata.setRedirectURIs(redirectUri)

                clientMetadata.setScopes(scopes)

                clientMetadata.setTokenEndpointAuthMethod("none")

                var authorizationCodeGrantParams = openID4CI.getAuthorizationCodeGrantParams()
                if (authorizationCodeGrantParams.hasIssuerState()) {
                    var issuerState = authorizationCodeGrantParams.issuerState()
                    clientMetadata.setIssuerState(issuerState)
                }

                var registrationResp =
                        Oauth2.registerClient(dynamicRegistrationEndpoint, clientMetadata, null)
                clientID = registrationResp.clientID()

                // Use the actual scopes registered by the authorization server,
                // which may differ from the scopes we specified in the metadata in our request.
                scopes = registrationResp.registeredMetadata().scopes()
            }
            authorizationLink = openID4CI.createAuthorizationURL(
                    clientID,
                    redirectURI,
                    oauthDiscoverableClientURI ?: "",
                    scopes,
            )

        }

        val flowTypeData: MutableMap<String, Any> = mutableMapOf()
        flowTypeData["pinRequired"] = pinRequired
        flowTypeData["authorizationURLLink"] = authorizationLink

        return flowTypeData
    }


    private fun initializeWalletInitiatedFlow(call: MethodCall): MutableList<Any> {
        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val issuerURI = call.argument<String>("issuerURI")
                ?: throw java.lang.Exception("Missing issuerURI argument")

        val walletInitiatedOpenID4CI =
                walletSDK.createOpenID4CIWalletInitiatedInteraction(issuerURI)
        this.walletInitiatedOpenID4CI = walletInitiatedOpenID4CI
        val supportedCredentials = walletInitiatedOpenID4CI.getSupportedCredentials()

        return getSupportedCredentialsList(supportedCredentials)
    }

    private fun createAuthorizationURLWalletInitiatedFlow(call: MethodCall): String {
        val walletInitiatedOpenID4CI = this.walletInitiatedOpenID4CI
                ?: throw java.lang.Exception("walletInitiatedOpenID4CI interaction is not initialized")

        val scopesFromArgs = call.argument<ArrayList<String>>("scopes")
                ?: throw java.lang.Exception("Missing scopes argument")

        val credentialTypesFromArgs = call.argument<ArrayList<String>>("credentialTypes")
                ?: throw java.lang.Exception("Missing credentialTypes argument")

        val clientID = call.argument<String>("clientID")
                ?: throw java.lang.Exception("Missing clientID argument")

        val redirectURI = call.argument<String>("redirectURI")
                ?: throw java.lang.Exception("Missing redirectURI argument")

        val credentialFormat = call.argument<String>("credentialFormat")
                ?: throw java.lang.Exception("Missing credentialFormat argument")

        val issuerURI = call.argument<String>("issuerURI")
                ?: throw java.lang.Exception("Missing issuerURI argument")


        var scopes = StringArray()

        for (scope in scopesFromArgs) {
            scopes.append(scope)
        }


        var credentialTypes = StringArray()

        for (credentialType in credentialTypesFromArgs) {
            credentialTypes.append(credentialType)
        }

        return walletInitiatedOpenID4CI.createAuthorizationURLWalletInitiatedFlow(
                scopes,
                credentialFormat,
                credentialTypes,
                clientID,
                redirectURI,
                issuerURI
        )
    }

    private fun getSupportedCredentialsList(supportedCredentials: SupportedCredentials): MutableList<Any> {
        val supportedCredentialsList = mutableListOf<Any>()

        for (index in 0 until (supportedCredentials.length())) {

            var typeStrArray = ArrayList<String>()
            for (i in 0 until (supportedCredentials.atIndex(index).types().length())) {
                val type = supportedCredentials.atIndex(index).types().atIndex(i)
                typeStrArray.add(type)
            }


            val localizedCredentialsDisplayRespList = mutableListOf<Any>()
            for (i in 0 until (supportedCredentials.atIndex(index).localizedDisplays().length())) {
                val localizedCredentialsDisplayResp: MutableMap<String, String> = mutableMapOf()
                localizedCredentialsDisplayResp["name"] =
                        supportedCredentials.atIndex(index).localizedDisplays().atIndex(i).name()
                localizedCredentialsDisplayResp["locale"] =
                        supportedCredentials.atIndex(index).localizedDisplays().atIndex(i).locale()
                localizedCredentialsDisplayResp["logo"] =
                        supportedCredentials.atIndex(index).localizedDisplays().atIndex(i).logo().url()
                localizedCredentialsDisplayResp["textColor"] =
                        supportedCredentials.atIndex(index).localizedDisplays().atIndex(i).textColor()
                localizedCredentialsDisplayResp["backgroundColor"] =
                        supportedCredentials.atIndex(index).localizedDisplays().atIndex(i)
                                .backgroundColor()

                localizedCredentialsDisplayRespList.addAll(listOf(localizedCredentialsDisplayResp))
            }

            val supportedCredentialResp: MutableMap<String, Any> = mutableMapOf()
            supportedCredentialResp["format"] = supportedCredentials.atIndex(index).format()
            supportedCredentialResp["types"] = typeStrArray
            supportedCredentialResp["display"] = localizedCredentialsDisplayRespList

            supportedCredentialsList.addAll(listOf(supportedCredentialResp))
        }

        return supportedCredentialsList
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
    private fun requestCredentials(call: MethodCall): List<HashMap<String, String>> {
        val otp = call.argument<String>("otp") ?: throw java.lang.Exception("otp params is missed")

        val attestationVC = call.argument<String>("attestationVC")

        if (attestationVC != null && attestationDID == null) {
            throw java.lang.Exception("AttestationDID should be created first")
        }

        val didDocResolution = this.didDocResolution
                ?: throw java.lang.Exception("DID should be created first")


        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")


        return openID4CI.requestCredentials(didDocResolution.assertionMethod(), otp,
                attestationVC, attestationDID?.assertionMethod())
    }


    private fun requestCredentialAuth(call: MethodCall): String? {
        val redirectURI = call.argument<String>("redirectURIWithParams")
                ?: throw java.lang.Exception("redirectURIWithParams is missed")

        val didDocResolution = this.didDocResolution
                ?: throw java.lang.Exception("DID should be created first")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.requestCredentialWithAuth(didDocResolution.assertionMethod(), redirectURI)
    }

    private fun requestCredentialWithWalletInitiatedFlow(call: MethodCall): String? {
        val redirectURI = call.argument<String>("redirectURIWithParams")
                ?: throw java.lang.Exception("redirectURIWithParams is missed")

        val didDocResolution = this.didDocResolution
                ?: throw java.lang.Exception("DID should be created first")

        val walletInitiatedOpenID4CI = this.walletInitiatedOpenID4CI
                ?: throw java.lang.Exception("walletInitiatedOpenID4CI not initiated. Initialize wallet initiated flow before this.")

        return walletInitiatedOpenID4CI.requestCredentialWithWalletInitiatedFlow(
                didDocResolution.assertionMethod(),
                redirectURI
        ).serialize()
    }


    private fun parseWalletSDKError(call: MethodCall): MutableMap<String, String> {
        val localizedErrorMessage = call.argument<String>("localizedErrorMessage")
                ?: throw java.lang.Exception("localizedErrorMessage is missing")

        val parsedError = Walleterror.parse(localizedErrorMessage)
        println("parsedError ->")
        print(parsedError)

        val parsedErrResp: MutableMap<String, String> = mutableMapOf()
        parsedErrResp["category"] = parsedError.category
        parsedErrResp["details"] = parsedError.details
        parsedErrResp["code"] = parsedError.code
        parsedErrResp["traceID"] = parsedError.traceID

        return parsedErrResp
    }

    /**
     * ResolveDisplay resolves display information for issued credentials based on an issuer's metadata, which is fetched
    using the issuer's (base) URI. The CredentialDisplays returns DisplayData object correspond to the VCs passed in and are in the
    same order. This method requires one or more VCs and the issuer's base URI.
    IssuerURI and array of credentials  are parsed using VcParse to be passed to resolveDisplay which returns the resolved Display Data
     */
    private fun resolveDisplayData(call: MethodCall): MutableMap<String, Any> {
        val issuerURI = call.argument<String>("uri")
                ?: throw java.lang.Exception("issuerURI params is missed")
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
                ?: throw java.lang.Exception("vcCredentials params is missed")

        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")


        val displayData = walletSDK.resolveCredentialsDisplayData(convertToVerifiableCredentialsArray(vcCredentials), issuerURI)


        val credsDisplayData: MutableMap<String, Any> = mutableMapOf()

        credsDisplayData["credentialsDisplay"] = convertCredentialsDisplayDataArray(displayData)
        credsDisplayData["issuerDisplay"] = displayData.issuerDisplay().serialize()

        return credsDisplayData
    }

    private fun getCredentialOfferDisplayData(): MutableMap<String, Any> {
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        val displayData = openID4CI.getCredentialOfferDisplayData()
        // credential issuer
        val issuerDisplay = displayData.issuerDisplay();


        val localizedIssuerDisplay: MutableMap<String, Any> = mutableMapOf()

        localizedIssuerDisplay["name"] = issuerDisplay.name()
        localizedIssuerDisplay["locale"] = issuerDisplay.locale()
        localizedIssuerDisplay["url"] = issuerDisplay.url()
        localizedIssuerDisplay["textColor"] = issuerDisplay.textColor()
        localizedIssuerDisplay["backgroundColor"] = issuerDisplay.backgroundColor()

        if (issuerDisplay.logo() != null) {
            localizedIssuerDisplay["logo"] = issuerDisplay.logo().url()
        } else {
            localizedIssuerDisplay["logo"] = ""
        }

        val issuerMetaDataResp: MutableMap<String, Any> = mutableMapOf()

        issuerMetaDataResp["localizedIssuerDisplay"] = localizedIssuerDisplay
        issuerMetaDataResp["offeredCredentials"] = serializeCredentialsDisplayData(displayData)


        return issuerMetaDataResp
    }

    private fun serializeCredentialsDisplayData(displayData: dev.trustbloc.wallet.sdk.display.Data): MutableList<Any> {

        val resolvedCredDisplayList = mutableListOf<Any>()
        for (i in 0 until (displayData.credentialDisplaysLength())) {
            val credentialDisplay = displayData.credentialDisplayAtIndex(i)

            val resolveDisplayResp: MutableMap<String, Any> = serializeCredentialDisplayData(credentialDisplay)

            resolveDisplayResp["issuerName"] = displayData.issuerDisplay().name()

            resolvedCredDisplayList.addAll(listOf(resolveDisplayResp))
        }
        return resolvedCredDisplayList
    }

    private fun serializeCredentialDisplayData(credentialDisplay: dev.trustbloc.wallet.sdk.display.CredentialDisplay): MutableMap<String, Any> {
        val claimList = mutableListOf<Any>()

        for (i in 0 until credentialDisplay.claimsLength()) {
            val claim = credentialDisplay.claimAtIndex(i)
            val claims: MutableMap<String, Any> = mutableMapOf()
            if (claim.isMasked) {
                claims["value"] = claim.value()
                claims["rawValue"] = claim.rawValue()
            }
            if (claim.hasOrder()) {
                val order = claim.order()
                claims["order"] = order
            }
            claims["rawValue"] = claim.rawValue()
            claims["valueType"] = claim.valueType()
            claims["label"] = claim.label()

            if (claim.valueType() == "attachment") {
                // For type=attachment, ignore the RawValue() and Value(), instead use Attachment() method.
                claims["rawValue"] = ""
                claims["value"] = ""
                val attachmentResp = claim.attachment()
                claims["uri"] = attachmentResp.uri()
            }


            claimList.addAll(listOf(claims))
        }
        var overview = credentialDisplay.overview()
        var resolveDisplayResp: MutableMap<String, Any> = mutableMapOf()
        resolveDisplayResp["claims"] = claimList
        resolveDisplayResp["overviewName"] = overview.name()
        resolveDisplayResp["logo"] = overview.logo().url()
        resolveDisplayResp["textColor"] = overview.textColor()
        resolveDisplayResp["backgroundColor"] = overview.backgroundColor()


        return resolveDisplayResp
    }

    private fun parseCredentialDisplay(call: MethodCall): MutableMap<String, Any> {
        val resolvedCredentialDisplayData = call.argument<String>("resolvedCredentialDisplayData")
                ?: throw java.lang.Exception("resolvedCredentialDisplayData params is missed")

        val displayData = Display.parseCredentialDisplay(resolvedCredentialDisplayData)

        return serializeCredentialDisplayData(displayData)
    }

    private fun parseIssuerDisplay(call: MethodCall): MutableMap<String, Any> {
        val issuerDisplayData = call.argument<String>("issuerDisplayData")
                ?: throw java.lang.Exception("displayData params is missed")

        val issuerDisplay = Display.parseIssuerDisplay(issuerDisplayData)

        val localizedIssuerDisplay: MutableMap<String, Any> = mutableMapOf()

        localizedIssuerDisplay["name"] = issuerDisplay.name()
        localizedIssuerDisplay["locale"] = issuerDisplay.locale()
        localizedIssuerDisplay["url"] = issuerDisplay.url()
        localizedIssuerDisplay["textColor"] = issuerDisplay.textColor()
        localizedIssuerDisplay["backgroundColor"] = issuerDisplay.backgroundColor()

        if (issuerDisplay.logo() != null) {
            localizedIssuerDisplay["logo"] = issuerDisplay.logo().url()
        } else {
            localizedIssuerDisplay["logo"] = ""
        }

        return localizedIssuerDisplay
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
            return parsedVC.issuerID()
        }
        return ""
    }

    private fun wellKnownDidConfig(call: MethodCall): MutableMap<String, Any> {
        val issuerID = call.argument<String>("issuerID")
                ?: throw java.lang.Exception("issuer id is missing")

        val walletSDK = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")

        val validationResult = try {
            Did.validateLinkedDomains(issuerID, walletSDK.didResolver, null)
        } catch (e: Exception) {
            throw java.lang.Exception("error while validating linked domains")
        }

        val didValidateResultResp: MutableMap<String, Any> = mutableMapOf()
        didValidateResultResp["isValid"] = validationResult.isValid
        didValidateResultResp["serviceURL"] = validationResult.serviceURL
        println(didValidateResultResp)

        return didValidateResultResp
    }

    private fun getVerifierDisplayData(): MutableMap<String, Any> {
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
    private fun processAuthorizationRequest(
            call: MethodCall,
            result: MethodChannel.Result
    ): List<String> {
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
            val matchedReq = openID4VP.getMatchedSubmissionRequirements(
                    convertToVerifiableCredentialsArray(storedCredentials)
            )
            var resp = convertVerifiableCredentialsArray(
                    matchedReq.atIndex(0).descriptorAtIndex(0).matchedVCs
            )
            if (resp.isEmpty()) {
                var typeConstraint = matchedReq.atIndex(0).descriptorAtIndex(0).typeConstraint()
                if (typeConstraint.isEmpty()) {
                    var schemas = matchedReq.atIndex(0).descriptorAtIndex(0).schemas()
                    var schemaList = mutableListOf<Any>()
                    for (i in 0..schemas.length()) {
                        val schemasResp: MutableMap<String, Any> = mutableMapOf()
                        if (schemas.atIndex(i) != null) {
                            schemasResp["required"] = schemas.atIndex(i).required()
                            schemasResp["uri"] = schemas.atIndex(i).uri()
                            schemaList.addAll(listOf(schemasResp))
                        }
                    }

                    result.error(
                            "NATIVE_ERR",
                            "No credentials conforming to the following schemas were found",
                            "$schemaList"
                    );
                }
                result.error(
                        "NATIVE_ERR",
                        "No credentials of type $typeConstraint were found",
                        "Required credential $typeConstraint is missing from the wallet"
                );
            }

            return resp

        }
        return listOf()
    }

    private fun getMatchedSubmissionRequirements(call: MethodCall): List<Any> {
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")
        val storedCredentials = call.argument<ArrayList<String>>("storedCredentials")
                ?: throw java.lang.Exception("storedCredentials params is missed")

        return convertSubmissionRequirementArray(
                openID4VP.getMatchedSubmissionRequirements(
                        convertToVerifiableCredentialsArray(
                                storedCredentials
                        )
                )
        )
    }

    private fun evaluateIssuanceTrustInfo(call: MethodCall): HashMap<String, Any> {
        val evaluateIssuanceURL = call.argument<String>("evaluateIssuanceURL")
                ?: throw java.lang.Exception("evaluateIssuanceURL params is missed")
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("OpenID4CI not initiated. Call authorize before this.")

        val result = openID4CI.checkWithTrustRegistry(evaluateIssuanceURL)

        return convertEvaluationResult(result)
    }

    private fun evaluatePresentationTrustInfo(call: MethodCall): HashMap<String, Any> {
        val evaluatePresentationURL = call.argument<String>("evaluatePresentationURL")
                ?: throw java.lang.Exception("evaluatePresentationURL params is missed")
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")

        val result = openID4VP.checkWithTrustRegistry(evaluatePresentationURL)

        return convertEvaluationResult(result)
    }

    private fun noConsentAcknowledgement(): String? {
        val ackRes = openID4VP?.noConsentAcknowledgement()
        try {
              ackRes?.noConsent()
                } catch (e: Exception) {
                    throw e
                }
            return ackRes?.serialize()
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
        val attestationVC = call.argument<String>("attestationVC")
        val selectedCredentials = call.argument<ArrayList<String>>("selectedCredentials")
        val customScopeList = call.argument<MutableMap<String, Any>>("customScopeList")
        val selectedCredentialsArray = if (selectedCredentials != null) {
            convertToVerifiableCredentialsArray(selectedCredentials)
        } else {
            //TODO: remove this after refactoring will be finished
            this.processAuthorizationRequestVCs
                    ?: throw java.lang.Exception("processAuthorizationRequest should be called first.")
        }

        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")

        openID4VP.presentCredential(selectedCredentialsArray, customScopeList, attestationDID?.assertionMethod(), attestationVC)
        this.openID4VP = null
    }

    private fun getAttestationVC(call: MethodCall): String {
        val sdk = this.walletSDK
                ?: throw java.lang.Exception("walletSDK not initiated. Call initSDK().")
        val attestationURL = call.argument<String>("attestationURL")
                ?: throw java.lang.Exception("attestationURL params is missed")

        val disableTLSVerify = call.argument<Boolean>("disableTLSVerify")
                ?: throw java.lang.Exception("disableTLSVerify params is missed")

        val attestationPayload = call.argument<String>("attestationPayload")
                ?: throw java.lang.Exception("authenticationMethod params is missed")

        val attestationToken = call.argument<String>("attestationToken")

        if (attestationDID == null) {
            attestationDID = sdk.createDID("ion", "ED25519")
        }

        return sdk.getAttestationVC(attestationDID!!.assertionMethod(), attestationURL,
                disableTLSVerify, attestationPayload, attestationToken)
    }

    private fun getCustomScope(): ArrayList<String> {
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call startVPInteraction.")

        return openID4VP.getCustomScope()
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
            activityDicResp.put("Status", status)
            activityDicResp.put("Issued By", client)
            activityDicResp.put("Operation", operation);
            activityDicResp.put("Activity Type", activityType);
            activityDicResp.put("Date", date);

            parseActivityList.putAll(activityDicResp)
        }

        return arrayListOf(parseActivityList)
    }

    private fun requireAcknowledgment(): Boolean {
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.requireAcknowledgment()
    }

    private fun acknowledgeSuccess() {
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.acknowledgeSuccess()
    }

    private fun acknowledgeReject() {
        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.acknowledgeReject()
    }


    companion object {
        private const val CHANNEL = "WalletSDKPlugin"
    }
}
