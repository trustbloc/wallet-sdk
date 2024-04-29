/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import android.content.Context
import androidx.test.filters.SmallTest
import androidx.test.platform.app.InstrumentationRegistry
import com.google.common.truth.Truth.assertThat
import dev.trustbloc.wallet.BuildConfig
import dev.trustbloc.wallet.sdk.api.StringArray
import dev.trustbloc.wallet.sdk.attestation.Attestation
import dev.trustbloc.wallet.sdk.credential.Inquirer
import dev.trustbloc.wallet.sdk.did.Resolver
import dev.trustbloc.wallet.sdk.did.ResolverOpts
import dev.trustbloc.wallet.sdk.didion.Didion
import dev.trustbloc.wallet.sdk.localkms.Localkms
import dev.trustbloc.wallet.sdk.oauth2.Oauth2
import dev.trustbloc.wallet.sdk.openid4ci.Acknowledgment
import dev.trustbloc.wallet.sdk.openid4ci.CreateAuthorizationURLOpts
import dev.trustbloc.wallet.sdk.openid4ci.InteractionOpts
import dev.trustbloc.wallet.sdk.openid4ci.IssuerInitiatedInteraction
import dev.trustbloc.wallet.sdk.openid4ci.IssuerInitiatedInteractionArgs
import dev.trustbloc.wallet.sdk.openid4ci.RequestCredentialWithPreAuthOpts
import dev.trustbloc.wallet.sdk.openid4vp.PresentCredentialOpts
import dev.trustbloc.wallet.sdk.otel.Otel
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray
import dev.trustbloc.wallet.sdk.version.Version
import dev.trustbloc.wallet.sdk.api.*
import okhttp3.OkHttpClient
import okhttp3.Request
import org.junit.Before
import org.junit.Test
import walletsdk.kmsStorage.KmsStore
import java.net.URI
import dev.trustbloc.wallet.sdk.openid4vp.Interaction as VPInteraction


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

        val jwk = kms.create(Localkms.KeyTypeED25519)

        val userDID = Didion.createLongForm(jwk)

        // Issue VCs
        val requestURI = BuildConfig.INITIATE_ISSUANCE_URL

        val attestClient = Attestation.newClient(Attestation.newCreateClientArgs(
                "https://localhost:8097/profiles/profileID/profileVersion/wallet/attestation/",
                crypto,
        ).disableHTTPClientTLSVerify().addHeader(Header("Authorization", "Bearer token")))

        val attestationVC = attestClient.getAttestationVC(
                userDID.assertionMethod(),
                """{
                        "type": "urn:attestation:application:trustbloc",
                        "application": {
                            "type":    "wallet-cli",
                            "name":    "wallet-cli",
                            "version": "1.0"
                        },
                        "compliance": {
                            "type": "fcra"				
                        }
                    }""",
        )

        val requiredOpenID4CIArgs = IssuerInitiatedInteractionArgs(requestURI, crypto, didResolver)

        val ciOpts = InteractionOpts()
        ciOpts.addHeader(trace.traceHeader())

        val ciInteraction = IssuerInitiatedInteraction(requiredOpenID4CIArgs, ciOpts)

        val pinRequired = ciInteraction.preAuthorizedCodeGrantParams().pinRequired()
        assertThat(pinRequired).isFalse()

        val opts = RequestCredentialWithPreAuthOpts()
        opts.setAttestationVC(userDID.assertionMethod(), attestationVC.serialize())

        val issuedCreds = ciInteraction.requestCredentialWithPreAuth(userDID.assertionMethod(), opts)

        assertThat(issuedCreds.length()).isGreaterThan(0)

        assertThat(ciInteraction.requireAcknowledgment()).isTrue()

        val acknowledgmentData = ciInteraction.acknowledgment().serialize()

        Acknowledgment(acknowledgmentData).success()

        //Presenting VCs
        val authorizationRequestURI = BuildConfig.INITIATE_VERIFICATION_URL

        val openID4VPInteractionRequiredArgs = dev.trustbloc.wallet.sdk.openid4vp.Args(
                authorizationRequestURI,
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
                inquirer.getSubmissionRequirements(credentialsQuery, issuedCreds)

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

        // Check trust info
        val trustInfo = vpInteraction.trustInfo()
        assertThat(trustInfo.did).isNotEmpty()
        assertThat(trustInfo.domain).isNotEmpty()

        // Credential trust info
        val oneOfTheCreds = requirementDescriptor.matchedVCs.atIndex(0)
        assertThat(oneOfTheCreds.id()).isNotEmpty()
        assertThat(oneOfTheCreds.types().length()).isGreaterThan(0)
        assertThat(oneOfTheCreds.issuerID()).isNotEmpty()
        assertThat(oneOfTheCreds.issuanceDate()).isGreaterThan(0)
        assertThat(oneOfTheCreds.expirationDate()).isGreaterThan(0)

        // Presenting from selected credentials.
        vpInteraction.presentCredentialOpts(selectedCreds, PresentCredentialOpts().addScopeClaim(
                "registration", """{"email":"test@example.com"}""").addScopeClaim("testscope", """{"data": "testdata"}""")
                .setAttestationVC(userDID.assertionMethod(), attestationVC.serialize()))
    }

    @Test
    fun testAuthFlow() {
        val trace = Otel.newTrace()

        assertThat(Version.getVersion()).isEqualTo("testVer")
        assertThat(Version.getGitRevision()).isEqualTo("testRev")
        assertThat(Version.getBuildTime()).isEqualTo("testTime")

        val kms = Localkms.newKMS(KmsStore(instrumentationContext))

        val resolverOpts = ResolverOpts()
        resolverOpts.setResolverServerURI("http://localhost:8072/1.0/identifiers")
        val didResolver = Resolver(resolverOpts)

        val crypto = kms.crypto

        val jwk = kms.create(Localkms.KeyTypeED25519)

        val userDID = Didion.createLongForm(jwk)

        // Issue VCs
        val requestURI = BuildConfig.INITIATE_ISSUANCE_URLS_AUTH_CODE_FLOW
        println("requestURI ->")
        println(requestURI)

        val requiredOpenID4CIArgs = IssuerInitiatedInteractionArgs(requestURI, crypto, didResolver)
        println("requiredOpenID4CIArgs")
        println(requiredOpenID4CIArgs)
        val ciOpts = InteractionOpts()
        ciOpts.addHeader(trace.traceHeader())

        val ciInteraction = IssuerInitiatedInteraction(requiredOpenID4CIArgs, ciOpts)
        var clientID = "oidc4vc_client"
        val redirectURI = "http://127.0.0.1/callback"
        var scopes = StringArray()
        scopes.append("openid").append("profile")

        assertThat(ciInteraction.dynamicClientRegistrationSupported()).isTrue()

        if (ciInteraction.dynamicClientRegistrationSupported()) {
            var dynamicRegistrationEndpoint = ciInteraction.dynamicClientRegistrationEndpoint()
            assertThat(dynamicRegistrationEndpoint).isNotEmpty()

            var clientMetadata = Oauth2.newClientMetadata()
            var grantTypesArr = StringArray()
            grantTypesArr.append("authorization_code")
            clientMetadata.setGrantTypes(grantTypesArr)
            assertThat(clientMetadata.grantTypes()).isNotNull()

            var redirectUri = StringArray()
            redirectUri.append(redirectURI)
            clientMetadata.setRedirectURIs(redirectUri)
            assertThat(clientMetadata.redirectURIs()).isNotNull()

            clientMetadata.setScopes(scopes)
            clientMetadata.setTokenEndpointAuthMethod("none")

            var authorizationCodeGrantParams = ciInteraction.authorizationCodeGrantParams()
            if (authorizationCodeGrantParams.hasIssuerState()) {
                var issuerState = authorizationCodeGrantParams.issuerState()
                clientMetadata.setIssuerState(issuerState)
                assertThat(clientMetadata.issuerState()).isNotEmpty()
            }

            var registrationResp = Oauth2.registerClient(dynamicRegistrationEndpoint, clientMetadata, null)
            clientID = registrationResp.clientID()
            assertThat(clientID).isNotEmpty()

            scopes = registrationResp.registeredMetadata().scopes()
            assertThat(scopes).isNotNull()
        }

        val authCodeGrant = ciInteraction.authorizationCodeGrantTypeSupported()
        assertThat(authCodeGrant).isTrue()

        val createAuthorizationURLOpts = CreateAuthorizationURLOpts().setScopes(scopes)

        val authorizationLink = ciInteraction.createAuthorizationURL(clientID, redirectURI, createAuthorizationURLOpts)
        assertThat(authorizationLink).isNotEmpty()

        var redirectUrl = URI(authorizationLink)

        val client = OkHttpClient.Builder()
                .retryOnConnectionFailure(true)
                .followRedirects(false)
                .build()

        var request = Request.Builder()
                .url(redirectUrl.toString())
                .header("Connection", "close")
                .build()
        val response = client.newCall(request).execute()
        assertThat(response.isRedirect).isTrue()
        var location = response.headers["Location"]
        assertThat(location).contains("cognito-mock.trustbloc.local")
        if (location != null) {
            if (location.contains("cognito-mock.trustbloc.local")) {
                var upr = URI(location.replace("cognito-mock.trustbloc.local", "localhost"));
                assertThat(upr.toString()).contains("localhost")
                var request = Request.Builder()
                        .url(upr.toString())
                        .header("Connection", "close")
                        .build()
                val response = client.newCall(request).clone().execute()
                location = response.headers["location"]
                assertThat(location).contains("oidc/redirect")
                var request2 = Request.Builder()
                        .url(location.toString())
                        .header("Connection", "close")
                        .build()
                val response2 = client.newCall(request2).clone().execute()
                location = response2.headers["location"]
                assertThat(location).contains("127.0.0.1")
                var issuedCreds = ciInteraction.requestCredentialWithAuth(userDID.assertionMethod(), location, null)
                assertThat(issuedCreds.length()).isGreaterThan(0)
            }
        }
    }
}
