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

public class MainActivity extends FlutterActivity {
    private static final String CHANNEL = "WalletSDKPlugin";

    @Override
    public void configureFlutterEngine(@NonNull FlutterEngine flutterEngine) {
        super.configureFlutterEngine(flutterEngine);
        new MethodChannel(flutterEngine.getDartExecutor().getBinaryMessenger(), CHANNEL)
                .setMethodCallHandler(
                        (call, result) -> {
                            if (call.method.equals("createDID")) {
                                KMS localKMS;
                                try {
                                    localKMS = Localkms.newKMS();
                                } catch (Exception e) {
                                    e.printStackTrace();
                                    result.error("Exception", "Error while creating kms", null);
                                    return;
                                }

                                Creator creatorDID;
                                try {
                                    creatorDID = Didcreator.newCreatorWithKeyWriter(localKMS);
                                } catch (Exception e) {
                                    e.printStackTrace();
                                    result.error("Exception", "Error while creating did creator", null);
                                    return;
                                }

                                try {
                                    byte[] doc = creatorDID.create("key", new CreateDIDOpts());
                                    String docString = new String(doc);
                                    result.success(docString);
                                } catch (Exception e) {
                                    e.printStackTrace();
                                    result.error("Exception", "Error while creating did", null);
                                }
                            }
                        }
                );
    }
}
