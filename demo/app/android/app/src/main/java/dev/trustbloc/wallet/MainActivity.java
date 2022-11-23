package dev.trustbloc.wallet;

import androidx.annotation.NonNull;
import dev.trustbloc.wallet.sdk.api.CreateDIDOpts;
import dev.trustbloc.wallet.sdk.didcreator.Creator;
import io.flutter.embedding.android.FlutterActivity;
import io.flutter.embedding.engine.FlutterEngine;
import io.flutter.plugin.common.MethodChannel;
import dev.trustbloc.wallet.sdk.localkms.KMS;
import dev.trustbloc.wallet.sdk.localkms.Localkms;
import dev.trustbloc.wallet.sdk.didcreator.Didcreator;
import dev.trustbloc.wallet.sdk.openid4ci.Interaction;
import dev.trustbloc.wallet.sdk.openid4ci.AuthorizeResult;
import dev.trustbloc.wallet.sdk.openid4ci.CredentialRequest;

public class MainActivity extends FlutterActivity {
    private static final String CHANNEL = "WalletSDKPlugin";
    String requestURI = null;
    Interaction newInteraction =  null;

    @Override
    public void configureFlutterEngine(@NonNull FlutterEngine flutterEngine) {
        super.configureFlutterEngine(flutterEngine);
        new MethodChannel(flutterEngine.getDartExecutor().getBinaryMessenger(), CHANNEL)
                .setMethodCallHandler(
                        (call, result) -> {
                            switch (call.method) {
                                case "createDID":
                                    try {
                                        String didCreated = createDID();
                                        result.success(didCreated);
                                        return;
                                    } catch (Exception e) {
                                        result.error("Exception", "Error while creating did creator", null);
                                        return;
                                    }
                                case "authorize":
                                    try {
                                        requestURI = call.argument("requestURI");
                                        Boolean userPinRequired = authorize(requestURI);
                                        result.success(userPinRequired);
                                        return;
                                    } catch (Exception e) {
                                        result.error("Exception", "Error while authorizing the oidc vc flow", null);
                                        return;
                                    }
                                case "requestCredential":
                                    String otp = call.argument("otp");
                                    byte[] credentialCreated = null;
                                    try {
                                        credentialCreated = requestCredential(otp);
                                        result.success(credentialCreated);
                                        return;
                                    } catch (Exception e) {
                                        result.error("Exception", "Error while requesting credential", e);

                                    }
                                    break;
                            }
                        }
                );
    }

    public Boolean authorize(String requestURI) throws Exception {
        System.out.print(requestURI);
        newInteraction =  new Interaction(requestURI);
        final AuthorizeResult authRes  = newInteraction.authorize();
        return  authRes.getUserPINRequired();
    }

    public byte[] requestCredential(String otp) throws Exception {
       final CredentialRequest credReq = new CredentialRequest();
       credReq.setUserPIN(otp);
       return newInteraction.requestCredential(credReq);
    }
    private String createDID(){
        KMS localKMS;
        try {
            localKMS = Localkms.newKMS();
        } catch (Exception e) {
            e.printStackTrace();
            throw new IllegalArgumentException(e.getMessage());
        }

        Creator creatorDID;
        try {
            creatorDID = Didcreator.newCreatorWithKeyWriter(localKMS);
        } catch (Exception e) {
            e.printStackTrace();
            throw new IllegalArgumentException(e.getMessage());
        }

        try {
            byte[] doc = creatorDID.create("key", new CreateDIDOpts());
            String docString = new String(doc);
            return docString;
        } catch (Exception e) {
            e.printStackTrace();
            throw new IllegalArgumentException(e.getMessage());

        }
    }
}
