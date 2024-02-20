/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.kmsStorage
import android.content.Context
import android.os.Build
import androidx.annotation.RequiresApi
import dev.trustbloc.wallet.sdk.localkms.Result
import dev.trustbloc.wallet.sdk.localkms.Store
import java.util.Base64


class KmsStore(context: Context) : Store {

    private val preferences = context.getSharedPreferences("KMSLocalStore", Context.MODE_PRIVATE)

    @RequiresApi(Build.VERSION_CODES.O)
    override fun get(keysetID: String?): Result {
        val localResult  = Result()
        val keyVal  = preferences.getString(keysetID, null)
        if(keyVal == null ) {
            localResult.found = false
            return localResult
        }
        localResult.found = true
        localResult.key = Base64.getDecoder().decode(keyVal)
        return localResult
    }

    @RequiresApi(Build.VERSION_CODES.O)
    override fun put(keySetID: String?, key: ByteArray?) {
        try{
            val editor = preferences.edit()
            editor.putString(keySetID, Base64.getEncoder().encodeToString(key))
            editor.apply()
        } catch(e: java.lang.Exception) {
            println(e)
        }
    }

}
