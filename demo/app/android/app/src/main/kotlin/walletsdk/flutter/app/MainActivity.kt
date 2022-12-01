package dev.trustbloc.wallet

import dev.trustbloc.wallet.sdk.api.CreateDIDOpts
import dev.trustbloc.wallet.sdk.api.Crypto
import dev.trustbloc.wallet.sdk.api.DIDResolver
import dev.trustbloc.wallet.sdk.api.LDDocumentLoader
import dev.trustbloc.wallet.sdk.didcreator.Creator
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.didcreator.Didcreator
import dev.trustbloc.wallet.sdk.didresolver.Didresolver
import dev.trustbloc.wallet.sdk.linkeddomains.DocumentLoader
import dev.trustbloc.wallet.sdk.openid4ci.Interaction
import dev.trustbloc.wallet.sdk.openid4ci.AuthorizeResult
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequestOpts
import io.flutter.plugin.common.MethodCall
import java.util.ArrayList
import walletsdk.openid4vp.OpenID4VP
import java.lang.Override


class MainActivity : FlutterActivity() {
    private var requestURI: String? = null
    private var newInteraction: Interaction? = null
    private var openID4VP: OpenID4VP? = null
    private var kms: KMS? = null;
    private var didResolver: DIDResolver? = null
    private var documentLoader: LDDocumentLoader? = null
    private var crypto: Crypto? = null

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
                                return@setMethodCallHandler
                            } catch (e: Exception) {
                                result.error("Exception", "Error while creating basic sdk services", null)
                                return@setMethodCallHandler
                            }
                        }

                        "createDID" -> {
                            try {
                                val didCreated = createDID()
                                result.success(didCreated)
                                return@setMethodCallHandler
                            } catch (e: Exception) {
                                result.error("Exception", "Error while creating did creator", null)
                                return@setMethodCallHandler
                            }
                        }
                        "authorize" -> {
                            try {
                                requestURI = call.argument("requestURI")
                                val userPinRequired = authorize(requestURI)
                                result.success(userPinRequired)
                                return@setMethodCallHandler
                            } catch (e: Exception) {
                                result.error(
                                        "Exception",
                                        "Error while authorizing the oidc vc flow",
                                        null
                                )
                                return@setMethodCallHandler
                            }
                        }
                        "requestCredential" -> {
                            val otp: String? = call.argument("otp")
                            try {
                                val credentialCreated = requestCredential(otp)
                                result.success(credentialCreated)
                                return@setMethodCallHandler
                            } catch (e: Exception) {
                                result.error("Exception", "Error while requesting credential", e)
                            }
                        }

                        "processAuthorizationRequest" -> {
                            try {
                                processAuthorizationRequest(call);
                                result.success(null)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }

                        "presentCredential" -> {
                            try {
                                presentCredential(call);
                                result.success(null)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while processing authorization request", e)
                            }
                        }
                    }
                }
    }

    private fun initSDK() {
        val kms = Localkms.newKMS(null)
        didResolver = Didresolver.newDIDResolver()
        crypto = kms.crypto
        documentLoader = DocumentLoader()

        this.kms = kms
    }

    private fun processAuthorizationRequest(call: MethodCall) {
        val openID4VP = createOpenID4VP()
        val authorizationRequest = call.argument<String>("authorizationRequest")
                ?: throw java.lang.Exception("authorizationRequest params is missed")
        val storedCredentials = call.argument<ArrayList<String>>("storedCredentials")
                ?: throw java.lang.Exception("storedCredentials params is missed")

        openID4VP.processAuthorizationRequest(authorizationRequest, storedCredentials)
        this.openID4VP = openID4VP
    }

    private fun presentCredential(call: MethodCall) {
        val openID4VP = this.openID4VP
                ?: throw java.lang.Exception("OpenID4VP not initiated. Call processAuthorizationRequest before this.")

        val signingKeyId = call.argument<String>("signingKeyId")
                ?: throw java.lang.Exception("signingKeyId params is missed")

        openID4VP.presentCredential(signingKeyId)
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

    @Throws(Exception::class)
    fun authorize(requestURI: String?): Boolean {
        print(requestURI)
        newInteraction = Interaction(requestURI)
        val authRes: AuthorizeResult = newInteraction!!.authorize()
        return authRes.userPINRequired
    }

    @Throws(Exception::class)
    fun requestCredential(otp: String?): ByteArray? {
        val credReq = CredentialRequestOpts()
        credReq.userPIN = otp
        return newInteraction?.requestCredential(credReq)
    }

    private fun createDID(): String {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val creatorDID: Creator = try {
            Didcreator.newCreatorWithKeyWriter(kms)
        } catch (e: Exception) {
            e.printStackTrace()
            throw IllegalArgumentException(e.message)
        }
        return try {
            val doc: ByteArray = creatorDID.create("key", CreateDIDOpts())
            String(doc)
        } catch (e: Exception) {
            e.printStackTrace()
            throw IllegalArgumentException(e.message)
        }
    }

    companion object {
        private const val CHANNEL = "WalletSDKPlugin"
    }
}
