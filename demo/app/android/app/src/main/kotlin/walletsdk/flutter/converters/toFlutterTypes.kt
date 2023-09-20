/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.flutter.converters

import dev.trustbloc.wallet.sdk.credential.InputDescriptor
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirement
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirementArray
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

fun convertVerifiableCredentialsArray(arr: CredentialsArray): List<String> {
    return List(arr.length().toInt()
    ) { i: Int ->
        arr.atIndex(i.toLong()).serialize()
    }
}

fun convertSubmissionRequirementArray(requirements: SubmissionRequirementArray): List<HashMap<String, Any>> {
    return List(requirements.len().toInt()
    ) { i: Int ->
        convertSubmissionRequirement(requirements.atIndex(i.toLong()))
    }
}

private fun convertSubmissionRequirement(req: SubmissionRequirement): HashMap<String, Any> {
    val result = HashMap<String, Any>()
    result.put("rule", req.rule())
    result.put("name", req.name())
    result.put("min", req.min())
    result.put("max", req.max())
    result.put("count", req.count())
    result.put("nested", convertNestedSubmissionRequirement(req))
    result.put("inputDescriptors", convertInputDescriptors(req))
    return result;
}

private fun convertInputDescriptor(desc: InputDescriptor): HashMap<String, Any> {
    val matchedVCsID = List<String>(desc.matchedVCs.length().toInt()) { i: Int ->
        desc.matchedVCs.atIndex(i.toLong()).id()
    }

    val matchedVCs = List<String>(desc.matchedVCs.length().toInt()) { i: Int ->
        desc.matchedVCs.atIndex(i.toLong()).serialize()
    }

    return HashMap(mapOf(
            "id" to desc.id,
            "name" to desc.name,
            "purpose" to desc.purpose,
            "matchedVCsID" to matchedVCsID,
            "matchedVCs" to matchedVCs,
    ))
}

private fun convertNestedSubmissionRequirement(req: SubmissionRequirement): List<HashMap<String, Any>> {
    return List(req.nestedRequirementLength().toInt()
    ) { i: Int ->
        convertSubmissionRequirement(req.nestedRequirementAtIndex(i.toLong()))
    }
}

private fun convertInputDescriptors(req: SubmissionRequirement): List<HashMap<String, Any>> {
    return List(req.descriptorLen().toInt()
    ) { i: Int ->
        convertInputDescriptor(req.descriptorAtIndex(i.toLong()))
    }
}