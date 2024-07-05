/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Foundation
import Walletsdk

func convertVerifiableCredentialsArray(arr: VerifiableCredentialsArray)-> Array<String> {
   var credList: [String] = []
   for i in 0..<arr.length() {
      credList.append((arr.atIndex(i)?.serialize(nil))!)
   }
    
   return credList
}

func convertEvaluationResult(res: TrustregistryEvaluationResult) -> Dictionary<String, Any> {
    var requestedAttestations: [String] = []
    for i in 0..<res.requestedAttestationLength() {
        requestedAttestations.append(res.requestedAttestation(at:i))
    }

    var resDic :[String: Any] = [
        "allowed":   res.allowed,
        "multipleCredentialAllowed": res.multipleCredentialAllowed,
        "errorCode":   res.errorCode,
        "errorMessage":   res.errorMessage,
        "requestedAttestations": requestedAttestations,
        "denyReason": res.denyReason()

    ]

    return resDic;
}

func convertVerifiableCredentialsWithIdArray(arr: VerifiableCredentialsArray)-> Array<Dictionary<String, Any>> {
   var credList: [Dictionary<String, Any>] = []
   for i in 0..<arr.length() {
       credList.append(["id": (arr.atIndex(i)?.id_())!, "content" :(arr.atIndex(i)?.serialize(nil))!])
   }
    
   return credList
}

func convertCredentialsDisplayDataArray(arr: DisplayData)-> Array<String> {
   var credList: [String] = []
    for i in 0..<arr.credentialDisplaysLength() {
      credList.append((arr.credentialDisplay(at: i)?.serialize(nil))!)
   }
    
   return credList
}

func convertSubmissionRequirementArray(requirements: CredentialSubmissionRequirementArray) -> Array<Dictionary<String, Any>> {
    var result: [Dictionary<String, Any>] = []
    for i in 0..<requirements.len() {
        result.append(convertSubmissionRequirement(req: requirements.atIndex(i)!))
    }
     
    return result
    
}

private func convertSubmissionRequirement(req: CredentialSubmissionRequirement)  -> Dictionary<String, Any> {
    
    var result : [String: Any] = [:]
    result["rule"] = req.rule()
    result["name"] = req.name()
    result["min"] = req.min()
    result["max"] = req.max()
    result["count"] = req.count()
    result["nested"] = convertNestedSubmissionRequirement(req: req)
    result["inputDescriptors"] = convertInputDescriptors(req: req)
    return result;
}

func convertInputDescriptor(desc: CredentialInputDescriptor) -> Dictionary<String, Any> {
    var matchedVCs: [String] = []
    var matchedVCsID: [String] = []
    
    for i in 0..<desc.matchedVCs!.length() {
        matchedVCs.append(desc.matchedVCs!.atIndex(i)!.serialize(nil))
    }

    for i in 0..<desc.matchedVCs!.length() {
        matchedVCsID.append(desc.matchedVCs!.atIndex(i)!.id_())
    }
    
    return [
        "id" : desc.id_,
        "name" : desc.name,
        "purpose" : desc.purpose,
        "matchedVCsID" : matchedVCsID,
        "matchedVCs" : matchedVCs,
    ]
}

func convertNestedSubmissionRequirement(req: CredentialSubmissionRequirement) -> Array<Dictionary<String, Any>> {
    var result: [Dictionary<String, Any>] = []

    for i in 0..<req.nestedRequirementLength() {
       result.append(convertSubmissionRequirement(req: req.nestedRequirement(at: i)!))
    }
    
    return result
}

func convertInputDescriptors(req: CredentialSubmissionRequirement)-> Array<Dictionary<String, Any>> {
    var result: [Dictionary<String, Any>] = []
    for i in 0..<req.descriptorLen() {
        result.append(convertInputDescriptor(desc: req.descriptor(at: i)!))
    }
    
    return result
}
