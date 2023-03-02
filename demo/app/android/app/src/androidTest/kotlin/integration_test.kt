import android.content.Context
import androidx.test.filters.SmallTest
import androidx.test.platform.app.InstrumentationRegistry
import com.google.common.truth.Truth.assertThat
import dev.trustbloc.wallet.BuildConfig
import dev.trustbloc.wallet.sdk.api.CreateDIDOpts
import dev.trustbloc.wallet.sdk.api.KeyWriter
import dev.trustbloc.wallet.sdk.api.VerifiableCredentialsArray
import dev.trustbloc.wallet.sdk.credential.CredentialsOpt
import dev.trustbloc.wallet.sdk.credential.Inquirer
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.ld.DocLoader
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.mem.ActivityLogger
import dev.trustbloc.wallet.sdk.openid4ci.ClientConfig
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequestOpts
import dev.trustbloc.wallet.sdk.openid4ci.Interaction
import dev.trustbloc.wallet.sdk.openid4vp.ClientConfig as VPClientConfig
import dev.trustbloc.wallet.sdk.openid4vp.Interaction as VPInteraction
import org.junit.Before
import org.junit.Test
import walletsdk.kmsStorage.KmsStore


@SmallTest
class IntegrationTest {
    lateinit var instrumentationContext: Context

    @Before
    fun setup() {
        instrumentationContext = InstrumentationRegistry.getInstrumentation().context
    }

    @Test
    fun fullFlow() {
        val kms = Localkms.newKMS(KmsStore(instrumentationContext))
        val didResolver = Resolver("http://localhost:8072/1.0/identifiers")
        val crypto = kms.crypto
        val documentLoader = DocLoader()
        val activityLogger = ActivityLogger()

        val didCreator = Creator(kms as KeyWriter)
        val userDID = didCreator.create("ion", CreateDIDOpts())

        // Issue VCs
        val cfg = ClientConfig("ClientID", crypto, didResolver, activityLogger)
        val requestURI = BuildConfig.INITIATE_ISSUANCE_URL

        val ciInteraction = Interaction(requestURI, cfg)

        val authorizeResult = ciInteraction.authorize()
        assertThat(authorizeResult.userPINRequired).isFalse()

        val otp = ""
        val issuedCreds = ciInteraction.requestCredential(CredentialRequestOpts(otp), userDID.assertionMethod())
        assertThat(issuedCreds.length()).isGreaterThan(0)

        //Presenting VCs
        val vpConfig = VPClientConfig(kms, crypto, didResolver, documentLoader, activityLogger)
        val authorizationRequestURI = BuildConfig.INITIATE_VERIFICATION_URL
        val vpInteraction = VPInteraction(authorizationRequestURI, vpConfig)

        val credentialsQuery = vpInteraction.getQuery()
        val inquirer = Inquirer(documentLoader)

        // TODO: maybe better to rename getSubmissionRequirements to something like matchCredentialsWithRequirements
        val submissionRequirements =
                inquirer.getSubmissionRequirements(credentialsQuery, CredentialsOpt(issuedCreds))

        assertThat(submissionRequirements.len()).isGreaterThan(0)
        val requirement = submissionRequirements.atIndex(0)
        // rule "all" means that we need to satisfy all input descriptors.
        // In case of multiple input descriptors we need to send one credential per descriptor
        // that satisfy it.
        assertThat(requirement.rule()).isEqualTo("all")

        // In current test case we have only one input descriptor. so we need send only one credential.
        assertThat(submissionRequirements.atIndex(0).descriptorLen()).isEqualTo(1)

        val requirementDescriptor = requirement.descriptorAtIndex(0)
        // matchedVCs contains list of credentials that match given input descriptor.
        assertThat(requirementDescriptor.matchedVCs.length()).isGreaterThan(0)

        val selectedCreds = VerifiableCredentialsArray()
        // Pick first credential from matched creds
        selectedCreds.add(requirementDescriptor.matchedVCs.atIndex(0))

        // TODO: maybe better to rename query to createPresentation.
        // Creating verifiable presentation from selected credentials.
        val verifiablePres = inquirer.query(credentialsQuery, CredentialsOpt(selectedCreds))

        vpInteraction.presentCredential(verifiablePres.content(), userDID.assertionMethod())
    }
}
