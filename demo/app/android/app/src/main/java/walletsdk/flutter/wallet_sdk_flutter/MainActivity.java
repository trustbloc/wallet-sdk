package walletsdk.flutter.wallet_sdk_flutter;

import androidx.annotation.NonNull;

import dev.trustbloc.wallet.sdk.api.CreateDIDOpts;
import io.flutter.embedding.android.FlutterActivity;
import io.flutter.embedding.engine.FlutterEngine;
import io.flutter.plugin.common.MethodChannel;
import dev.trustbloc.wallet.sdk.creator.Creator;
import dev.trustbloc.wallet.sdk.creator.DIDCreator;

public class MainActivity extends FlutterActivity {
    private static final String CHANNEL = "WalletSDKPlugin";

    @Override
    public void configureFlutterEngine(@NonNull FlutterEngine flutterEngine) {
        super.configureFlutterEngine(flutterEngine);
        new MethodChannel(flutterEngine.getDartExecutor().getBinaryMessenger(), CHANNEL)
                .setMethodCallHandler(
                        (call, result) -> {
                            if (call.method.equals("createDID")) {
                              DIDCreator creatorDID =  Creator.newDIDCreator(null);
                                try {
                                  byte[] doc =  creatorDID.create("key", new CreateDIDOpts());
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
