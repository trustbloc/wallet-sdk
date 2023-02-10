package dev.trustbloc.wallet

import android.os.Build
import androidx.annotation.RequiresApi
import dev.trustbloc.wallet.sdk.api.*
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.ld.DocLoader
import dev.trustbloc.wallet.sdk.localkms.KMS
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.walleterror.Walleterror
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import dev.trustbloc.wallet.sdk.mem.ActivityLogger
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import walletsdk.kmsStorage.KmsStore
import walletsdk.openid4ci.OpenID4CI
import walletsdk.openid4vp.OpenID4VP
import java.security.Timestamp
import java.text.SimpleDateFormat
import java.util.*
import kotlin.collections.ArrayList


class MainActivity : FlutterActivity() {
    private var openID4CI: OpenID4CI? = null
    private var openID4VP: OpenID4VP? = null
    private var kms: KMS? = null
    private var didResolver: DIDResolver? = null
    private var documentLoader: LDDocumentLoader? = null
    private var crypto: Crypto? = null
    private var didDocID: String? = null
    private var activityLogger: ActivityLogger? = null
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

                         "fetchDID" -> {
                             try {
                                 val didID = call.argument<String>("didID")
                                 if (didDocID == null) {
                                     didDocID = didID
                                 }
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
                                println("here or no")
                                val activityLoggerResp = activityLogger()
                                result.success(activityLoggerResp)
                            } catch (e: Exception) {
                                result.error("Exception", "Error while fetching activity logger request", e)
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
        activityLogger = ActivityLogger()
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

    private fun issuerURI(): String? {
        return openID4CI?.issuerURI()
    }


    private fun activityLogger(): MutableList<Any> {
        val arrayList = mutableListOf<Any>()
        var aryLength = activityLogger?.length()
        for (i in 0..aryLength!!){
            val status = activityLogger?.atIndex(i)?.status()
            val client = activityLogger?.atIndex(i)?.client()
            val activityType = activityLogger?.atIndex(i)?.type()
            val timestampDate = activityLogger?.atIndex(i)?.unixTimestamp()

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

        val activityLogger = this.activityLogger
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")


        return OpenID4VP(kms, crypto, didResolver, documentLoader, activityLogger)
    }

    private fun authorize(call: MethodCall): Boolean {
        val requestURI = call.argument<String>("requestURI")
                ?: throw java.lang.Exception("requestURI params is missed")

        val didResolver = this.didResolver
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val crypto = this.crypto
                ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val activityLogger = this.activityLogger
            ?: throw java.lang.Exception("SDK is not initialized, call initSDK()")

        val openID4CI = OpenID4CI(
                requestURI,
                crypto,
                didResolver,
                activityLogger
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

    private fun resolveCredentialDisplay(call: MethodCall): String? {
        val issuerURI = call.argument<String>("uri")
            ?: throw java.lang.Exception("issuerURI params is missed")
        val vcCredentials = call.argument<ArrayList<String>>("vcCredentials")
            ?: throw java.lang.Exception("vcCredentials params is missed")

        val openID4CI = this.openID4CI
            ?: throw java.lang.Exception("openID4CI not initiated. Call authorize before this.")

        return openID4CI.resolveCredentialDisplay(issuerURI, vcCredentials)
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
