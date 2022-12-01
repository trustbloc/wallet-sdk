package walletsdk.util

import dev.trustbloc.wallet.sdk.api.JSONArray

fun createJsonArray(objs: List<String>): JSONArray {
    val arr = JSONArray()

    var result = "["
    for (i in objs.indices) {
        if (i != 0) {
            result += ","
        }
        result += objs[i]
    }

    result += "]"

    arr.data = result.toByteArray()
    return arr
}