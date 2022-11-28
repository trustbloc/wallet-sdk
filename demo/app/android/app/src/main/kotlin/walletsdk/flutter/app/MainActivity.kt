package dev.trustbloc.wallet

import dev.trustbloc.wallet.sdk.api.CreateDIDOpts
import dev.trustbloc.wallet.sdk.didcreator.Creator
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.didcreator.Didcreator
import dev.trustbloc.wallet.sdk.openid4ci.Interaction
import dev.trustbloc.wallet.sdk.openid4ci.AuthorizeResult
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequestOpts
import java.lang.Override


class MainActivity : FlutterActivity() {
    private var requestURI: String? = null
    private var newInteraction: Interaction? = null
    @Override
    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
            .setMethodCallHandler { call, result ->
                when (call.method) {
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
                }
            }
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
        val localKMS: KMS = try {
            Localkms.newKMS()
        } catch (e: Exception) {
            e.printStackTrace()
            throw IllegalArgumentException(e.message)
        }
        val creatorDID: Creator = try {
            Didcreator.newCreatorWithKeyWriter(localKMS)
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
