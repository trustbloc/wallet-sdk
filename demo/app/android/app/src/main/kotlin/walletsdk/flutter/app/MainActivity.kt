package dev.trustbloc.wallet

import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.Creator
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.ld.DocLoader
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.localkms.SignerCreator
import io.flutter.plugin.common.MethodCall
import walletsdk.openid4ci.OpenID4CI
import java.util.ArrayList
import walletsdk.openid4vp.OpenID4VP
import java.lang.Override


class MainActivity : FlutterActivity() {
    private var openID4CI: OpenID4CI? = null
    private var openID4VP: OpenID4VP? = null
    private var kms: KMS? = null;
    private var didResolver: DIDResolver? = null
    private var signerCreator: SignerCreator? = null
    private var documentLoader: LDDocumentLoader? = null
    private var crypto: Crypto? = null
    private var didDocID: String? = null

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
                                val didCreated = createDID()
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
                                result.error(
                                        "Exception",
                                        "Error while authorizing the oidc vc flow",
                                        e
                                )

                            }
                        }
                        "requestCredential" -> {
                            try {
                                val credentialCreated = requestCredential(call)
                                result.success(credentialCreated)

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
        didResolver = Resolver("")
        crypto = kms.crypto
        documentLoader = DocLoader()
        signerCreator = Localkms.createSignerCreator(kms)

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

    fun authorize(call: MethodCall): Boolean {
        val requestURI = call.argument<String>("requestURI")
                ?: throw java.lang.Exception("requestURI params is missed")

        val didDocID = this.didDocID ?: throw java.lang.Exception("DID should be created first")

        val didResolver = this.didResolver
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val signerCreator = this.signerCreator
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val openID4CI = OpenID4CI(
                requestURI,
                didDocID,
                signerCreator,
                didResolver,
        )

        val authRes = openID4CI.authorize()

        this.openID4CI = openID4CI

        return authRes.userPINRequired
    }

    private fun requestCredential(call: MethodCall): String? {
        val otp = call.argument<String>("otp")

        val openID4CI = this.openID4CI
                ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        val resp = openID4CI.requestCredential(otp) ?: return null
        return resp
    }

    private fun createDID(): String {
        val kms = this.kms ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val creatorDID = Creator(kms as KeyWriter)

        val doc = creatorDID.create("key", CreateDIDOpts())

        didDocID = doc.id()

        return String(doc.content)
    }

    companion object {
        private const val CHANNEL = "WalletSDKPlugin"
    }
}
