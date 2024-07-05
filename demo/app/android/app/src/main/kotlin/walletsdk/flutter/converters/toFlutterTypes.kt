/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletsdk.flutter.converters

import dev.trustbloc.wallet.sdk.credential.InputDescriptor
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirement
import dev.trustbloc.wallet.sdk.credential.SubmissionRequirementArray
import dev.trustbloc.wallet.sdk.display.Data
import dev.trustbloc.wallet.sdk.trustregistry.EvaluationResult
import dev.trustbloc.wallet.sdk.verifiable.CredentialsArray

public fun convertVerifiableCredentialsArray(arr: CredentialsArray): List<String> {
    return List(arr.length().toInt()
    ) { i: Int ->
        arr.atIndex(i.toLong()).serialize()
    }
}

public fun convertVerifiableCredentialsWithIdArray(arr: CredentialsArray): List<HashMap<String, String>> {
    return List(arr.length().toInt()
    ) { i: Int ->
        hashMapOf("id" to arr.atIndex(i.toLong()).id(), "content" to arr.atIndex(i.toLong()).serialize())
    }
}

public fun convertCredentialsDisplayDataArray(arr: Data): List<String> {
    return List(arr.credentialDisplaysLength().toInt()
    ) { i: Int ->
        arr.credentialDisplayAtIndex(i.toLong()).serialize()
    }
}

public fun convertSubmissionRequirementArray(requirements: SubmissionRequirementArray): List<HashMap<String, Any>> {
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

public fun convertEvaluationResult(result: EvaluationResult): HashMap<String, Any> {
    val requestedAttestations = List(result.requestedAttestationLength().toInt()
    ) { i: Int ->
        result.requestedAttestationAtIndex(i.toLong())
    }

    return hashMapOf(
            "allowed" to result.allowed,
            "multipleCredentialAllowed" to  result.multipleCredentialAllowed,
            "errorCode" to result.errorCode,
            "errorMessage" to result.errorMessage,
            "requestedAttestations" to requestedAttestations,
            "denyReason" to result.denyReason(),
    )
}