package dev.trustbloc.wallet

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.Creator
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.ld.DocLoader
import dev.trustbloc.wallet.sdk.walleterror.Walleterror
import dev.trustbloc.wallet.sdk.localkms.Localkms
import io.flutter.plugin.common.MethodCall
import walletsdk.kmsStorage.KmsStore
import walletsdk.openid4ci.OpenID4CI
import java.util.ArrayList
import walletsdk.openid4vp.OpenID4VP
import java.lang.Override


class MainActivity : FlutterActivity() {
    private var openID4CI: OpenID4CI? = null
    private var openID4VP: OpenID4VP? = null
    private var kms: KMS? = null
    private var didResolver: DIDResolver? = null
    private var documentLoader: LDDocumentLoader? = null
    private var crypto: Crypto? = null
    private var didDocID: String? = null
    private var didVerificationMethod: VerificationMethod? = null

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

                        "resolveCredentialDisplay" -> {
                            try {
                                val credentialDisplay = resolveCredentialDisplay()

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

                        "presentCredential" -> {
                            try {
                                presentCredential()
                                result.success(null)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }
                    }
                }
    }

    private fun initSDK() {
        val kmsLocalStore = KmsStore(context)
        val kms = Localkms.newKMS(kmsLocalStore)
        didResolver = Resolver("")
        crypto = kms.crypto
        documentLoader = DocLoader()

        this.kms = kms
    }

    private fun processAuthorizationRequest(call: MethodCall): List<String> {
        val openID4VP = createOpenID4VP()
        val authorizationRequest = call.argument<String>("authorizationRequest")
                ?: throw java.lang.Exception("authorizationRequest params is missed")
        val storedCredentials = call.argument<ArrayList<String>>("storedCredentials")
                ?: throw java.lang.Exception("storedCredentials params is missed")

        this.openID4VP = openID4VP

        return openID4VP.processAuthorizationRequest(authorizationRequest, storedCredentials)
    }

    private fun presentCredential() {
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call processAuthorizationRequest before this.")

        val didVerificationMethod = this.didVerificationMethod
                ?: throw java.lang.Exception("DID should be created first")

        openID4VP.presentCredential(didVerificationMethod)
    }


    private fun createOpenID4VP(): OpenID4VP {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")
        val crypto = this.crypto
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")
        val didResolver = this.didResolver
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")
        val documentLoader = this.documentLoader
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        return OpenID4VP(kms, crypto, didResolver, documentLoader)
    }

    private fun authorize(call: MethodCall): Boolean {
        val requestURI = call.argument<String>("requestURI")
                ?: throw java.lang.Exception("requestURI params is missed")

        val didResolver = this.didResolver
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val crypto = this.crypto
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val openID4CI = OpenID4CI(
                requestURI,
                crypto,
                didResolver,
        )

        val authRes = openID4CI.authorize()

        this.openID4CI = openID4CI

        return authRes.userPINRequired
    }

    private fun requestCredential(call: MethodCall): VerifiableCredential? {
        val otp = call.argument<String>("otp")

        val openID4CI = this.openID4CI
            ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        val didVerificationMethod = this.didVerificationMethod
                ?: throw java.lang.Exception("DID should be created first")

        return openID4CI.requestCredential(otp, didVerificationMethod)
    }

    private fun resolveCredentialDisplay(): String? {
        val openID4CI = this.openID4CI
            ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.resolveCredentialDisplay()
    }


    private fun createDID(call: MethodCall): String {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val creatorDID = Creator(kms as KeyWriter)
        val didMethodType = call.argument<String>("didMethodType")
        val doc = creatorDID.create(didMethodType, CreateDIDOpts())

        didDocID = doc.id()
        didVerificationMethod = doc.assertionMethod()

        return String(doc.content)
    }

    companion object {
        private const val CHANNEL = "WalletSDKPlugin"
    }
}
