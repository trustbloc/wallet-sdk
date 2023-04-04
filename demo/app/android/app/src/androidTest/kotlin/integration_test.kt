import android.content.Context
import androidx.test.filters.SmallTest
import androidx.test.platform.app.InstrumentationRegistry
import com.google.common.truth.Truth.assertThat
import dev.trustbloc.wallet.BuildConfig
import dev.trustbloc.wallet.sdk.api.KeyWriter
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray
import dev.trustbloc.wallet.sdk.credential.CredentialsArg
import dev.trustbloc.wallet.sdk.credential.Inquirer
import dev.trustbloc.wallet.sdk.did.Creator
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.did.ResolverOpts
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.openid4ci.*
import dev.trustbloc.wallet.sdk.openid4vp.Interaction as VPInteraction
import dev.trustbloc.wallet.sdk.version.Version
import dev.trustbloc.wallet.sdk.otel.Otel
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
        val trace = Otel.newTrace()

        assertThat(Version.getVersion()).isEqualTo("testVer")
        assertThat(Version.getGitRevision()).isEqualTo("testRev")
        assertThat(Version.getBuildTime()).isEqualTo("testTime")

        val kms = Localkms.newKMS(KmsStore(instrumentationContext))

        val resolverOpts = ResolverOpts()
        resolverOpts.setResolverServerURI("http://localhost:8072/1.0/identifiers")
        val didResolver = Resolver(resolverOpts)

        val crypto = kms.crypto

        val didCreator = Creator(kms as KeyWriter)
        val userDID = didCreator.create("ion", null)

        // Issue VCs
        val requestURI = BuildConfig.INITIATE_ISSUANCE_URL

        val requiredOpenID4CIArgs = Args(requestURI, "ClientID", crypto, didResolver)

        val ciOpts = Opts()
        ciOpts.addHeader(trace.traceHeader())

        val ciInteraction = Interaction(requiredOpenID4CIArgs, ciOpts)

        val authorizeResult = ciInteraction.authorize()
        assertThat(authorizeResult.userPINRequired).isFalse()

        val issuedCreds = ciInteraction.requestCredential(userDID.assertionMethod())
        assertThat(issuedCreds.length()).isGreaterThan(0)

        //Presenting VCs
        val authorizationRequestURI = BuildConfig.INITIATE_VERIFICATION_URL

        val openID4VPInteractionRequiredArgs = dev.trustbloc.wallet.sdk.openid4vp.Args(
            authorizationRequestURI,
            kms,
            crypto,
            didResolver
        )

        val vpOpts = dev.trustbloc.wallet.sdk.openid4vp.Opts()
        vpOpts.addHeader(trace.traceHeader())

        val vpInteraction = VPInteraction(openID4VPInteractionRequiredArgs, vpOpts)

        val credentialsQuery = vpInteraction.getQuery()

        val inquirer = Inquirer(null)

        // TODO: maybe better to rename getSubmissionRequirements to something like matchCredentialsWithRequirements
        val submissionRequirements =
                inquirer.getSubmissionRequirements(credentialsQuery, CredentialsArg(issuedCreds))

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

        val selectedCreds = CredentialsArray()
        // Pick first credential from matched creds
        selectedCreds.add(requirementDescriptor.matchedVCs.atIndex(0))

        // Presenting from selected credentials.
        vpInteraction.presentCredential(selectedCreds)
    }
}
